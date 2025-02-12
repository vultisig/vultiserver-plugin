package payroll

import (
	"context"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/hibiken/asynq"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethereum/go-ethereum/accounts/abi"
	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/mobile-tss-lib/tss"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/plugin"
	"github.com/vultisig/vultisigner/storage"
)

const PLUGIN_TYPE = "payroll"
const erc20ABI = `[{
    "name": "transfer",
    "type": "function",
    "inputs": [
        {"name": "recipient", "type": "address"},
        {"name": "amount", "type": "uint256"}
    ],
    "outputs": [{"name": "", "type": "bool"}]
}]`

//go:embed frontend
var frontend embed.FS

type PayrollPlugin struct {
	db           storage.DatabaseStorage
	nonceManager *plugin.NonceManager
	rpcClient    *ethclient.Client
	logger       logrus.FieldLogger
}

func NewPayrollPlugin(db storage.DatabaseStorage, logger logrus.FieldLogger, rpcClient *ethclient.Client) *PayrollPlugin {
	return &PayrollPlugin{
		db:           db,
		rpcClient:    rpcClient,
		nonceManager: plugin.NewNonceManager(rpcClient),
		logger:       logger,
	}
}

func (p *PayrollPlugin) SignPluginMessages(e echo.Context) error {
	return nil
}

func (p *PayrollPlugin) ValidatePluginPolicy(policyDoc types.PluginPolicy) error {
	if policyDoc.PluginType != PLUGIN_TYPE {
		return fmt.Errorf("policy does not match plugin type, expected: %s, got: %s", PLUGIN_TYPE, policyDoc.PluginType)
	}

	var payrollPolicy types.PayrollPolicy
	if err := json.Unmarshal(policyDoc.Policy, &payrollPolicy); err != nil {
		return fmt.Errorf("fail to unmarshal payroll policy, err: %w", err)
	}

	if len(payrollPolicy.Recipients) == 0 {
		return fmt.Errorf("no recipients found in payroll policy")
	}

	if len(payrollPolicy.ChainID) != len(payrollPolicy.Recipients) {
		p.logger.WithFields(logrus.Fields{
			"chain_id_length":   payrollPolicy.ChainID,
			"recipients_length": payrollPolicy.Recipients,
		}).Error("chain_id array length does not match number of recipients")
		return fmt.Errorf("chain_id array length must match number of recipients")
	}

	if len(payrollPolicy.TokenID) != len(payrollPolicy.Recipients) {
		return fmt.Errorf("token_id array length must match number of recipients")
	}

	for _, recipient := range payrollPolicy.Recipients {
		mixedCaseAddress, err := gcommon.NewMixedcaseAddressFromString(recipient.Address)
		if err != nil {
			return fmt.Errorf("invalid recipient address: %s", recipient.Address)
		}

		// if the address is not all lowercase, check the checksum
		if strings.ToLower(recipient.Address) != recipient.Address {
			if !mixedCaseAddress.ValidChecksum() {
				return fmt.Errorf("invalid recipient address checksum: %s", recipient.Address)
			}
		}

		if recipient.Amount == "" {
			return fmt.Errorf("amount is required for recipient %s", recipient.Address)
		}

		_, ok := new(big.Int).SetString(recipient.Amount, 10)
		if !ok {
			return fmt.Errorf("invalid amount for recipient %s: %s", recipient.Address, recipient.Amount)
		}
	}

	return nil
}

func (p *PayrollPlugin) ConfigurePlugin(e echo.Context) error {
	return nil
}

func (p *PayrollPlugin) ProposeTransactions(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error) {
	var txs []types.PluginKeysignRequest
	err := p.ValidatePluginPolicy(policy)
	if err != nil {
		return txs, fmt.Errorf("failed to validate plugin policy: %v", err)
	}

	var payrollPolicy types.PayrollPolicy
	if err := json.Unmarshal(policy.Policy, &payrollPolicy); err != nil {
		return txs, fmt.Errorf("fail to unmarshal payroll policy, err: %w", err)
	}

	for i, recipient := range payrollPolicy.Recipients {
		txHash, rawTx, err := p.generatePayrollTransaction(
			recipient.Amount,
			recipient.Address,
			payrollPolicy.ChainID[i],
			payrollPolicy.TokenID[i],
			policy.PublicKey,
		)
		fmt.Printf("Chain ID TEST 1: %s\n", payrollPolicy.ChainID[i])
		if err != nil {
			return []types.PluginKeysignRequest{}, fmt.Errorf("failed to generate transaction hash: %v", err)
		}

		// Create signing request
		signRequest := types.PluginKeysignRequest{
			KeysignRequest: types.KeysignRequest{
				PublicKey:        policy.PublicKey,
				Messages:         []string{txHash},
				SessionID:        uuid.New().String(),
				HexEncryptionKey: "0123456789abcdef0123456789abcdef",
				DerivePath:       "m/44/60/0/0/0",
				IsECDSA:          true, //Todo : do a mapping with chain id to know if it is ecdsa?
				VaultPassword:    "your-secure-password",
			},
			Transaction: hex.EncodeToString(rawTx),
			PluginID:    policy.PluginID,
			PolicyID:    policy.ID,
		}
		txs = append(txs, signRequest)
	}

	signRequest := txs[0]
	txBytes, err := hex.DecodeString(signRequest.Transaction)
	if err != nil {
		p.logger.Errorf("Failed to decode transaction hex: %v", err)
		return []types.PluginKeysignRequest{}, fmt.Errorf("failed to decode transaction hex: %w", err)
	}
	//unmarshal tx from sign req.transaction
	tx := &gtypes.Transaction{}
	err = tx.UnmarshalBinary(txBytes)
	if err != nil {
		p.logger.Errorf("Failed to unmarshal transaction: %v", err)
		return []types.PluginKeysignRequest{}, fmt.Errorf("failed to unmarshal transaction: %d:", err)
	}
	fmt.Printf("Chain ID TEST 2: %s\n", tx.ChainId().String())
	fmt.Printf("len TEST 2: %d\n", len(txs))

	return txs, nil
}

func (p *PayrollPlugin) ValidateTransactionProposal(policy types.PluginPolicy, txs []types.PluginKeysignRequest) error {
	err := p.ValidatePluginPolicy(policy)
	if err != nil {
		return fmt.Errorf("failed to validate plugin policy: %v", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return fmt.Errorf("failed to parse ABI: %v", err)
	}

	var payrollPolicy types.PayrollPolicy
	if err := json.Unmarshal(policy.Policy, &payrollPolicy); err != nil {
		return fmt.Errorf("fail to unmarshal payroll policy, err: %w", err)
	}

	for i, tx := range txs {
		var parsedTx *gtypes.Transaction
		txBytes, err := hex.DecodeString(tx.Transaction)
		if err != nil {
			return fmt.Errorf("failed to decode transaction: %v", err)
		}

		err = rlp.DecodeBytes(txBytes, &parsedTx)
		if err != nil {
			return fmt.Errorf("failed to parse transaction: %v", err)
		}

		txDestination := parsedTx.To()
		if txDestination == nil {
			return fmt.Errorf("transaction destination is nil")
		}

		if strings.ToLower(txDestination.Hex()) != strings.ToLower(payrollPolicy.TokenID[i]) { //todo : why we compare to tokenId and not recipient address?
			return fmt.Errorf("transaction destination does not match token ID")
		}

		txData := parsedTx.Data()
		m, err := parsedABI.MethodById(txData[:4])
		if err != nil {
			return fmt.Errorf("failed to get method by ID: %v", err)
		}

		v := make(map[string]interface{})
		if err := m.Inputs.UnpackIntoMap(v, txData[4:]); err != nil {
			return fmt.Errorf("failed to unpack transaction data: %v", err)
		}

		fmt.Printf("Decoded: %+v\n", v)

		recipientAddress, ok := v["recipient"].(gcommon.Address)
		if !ok {
			return fmt.Errorf("failed to get recipient address")
		}

		var recipientFound bool
		for _, recipient := range payrollPolicy.Recipients {
			if strings.EqualFold(recipientAddress.Hex(), recipient.Address) {
				recipientFound = true
				break
			}
		}

		if !recipientFound {
			return fmt.Errorf("recipient not found in policy")
		}
	}

	return nil
}

func (p *PayrollPlugin) generatePayrollTransaction(amountString string, recipientString string, chainID string, tokenID string, publicKey string) (string, []byte, error) {
	amount := new(big.Int)
	amount.SetString(amountString, 10)
	recipient := gcommon.HexToAddress(recipientString)

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	inputData, err := parsedABI.Pack("transfer", recipient, amount)
	if err != nil {
		return "", nil, fmt.Errorf("failed to pack transfer data: %v", err)
	}

	// create call message to estimate gas
	callMsg := ethereum.CallMsg{
		From:  recipient, //todo : this works, but maybe better to put the correct sender address once we have it
		To:    &recipient,
		Data:  inputData,
		Value: big.NewInt(0),
	}
	// estimate gas limit
	gasLimit, err := p.rpcClient.EstimateGas(context.Background(), callMsg)
	if err != nil {
		return "", nil, fmt.Errorf("failed to estimate gas: %v", err)
	}
	// add 20% to gas limit for safety
	gasLimit = gasLimit * 300 / 100
	// get suggested gas price
	gasPrice, err := p.rpcClient.SuggestGasPrice(context.Background())
	if err != nil {
		return "", nil, fmt.Errorf("failed to get gas price: %v", err)
	}
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(3))
	// Parse chain ID
	chainIDInt := new(big.Int)
	chainIDInt.SetString(chainID, 10)
	fmt.Printf("Chain ID TEST 3: %s\n", chainIDInt.String())

	derivedAddress, err := DeriveAddress(publicKey, "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "m/44/60/0/0/0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to derive address: %v", err)
	}

	nextNonce, err := p.GetNextNonce(derivedAddress.Hex())
	if err != nil {
		return "", nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	// Create unsigned transaction data
	txData := []interface{}{
		uint64(nextNonce),             // nonce
		gasPrice,                      // gas price
		uint64(gasLimit),              // gas limit
		gcommon.HexToAddress(tokenID), // to address
		big.NewInt(0),                 // value
		inputData,                     // data
		chainIDInt,                    // chain id
		uint(0),                       // empty v
		uint(0),                       // empty r
	}

	// Log each component separately
	p.logger.WithFields(logrus.Fields{
		"nonce":     txData[0],
		"gas_price": txData[1].(*big.Int).String(),
		"gas_limit": txData[2],
		"to":        txData[3].(gcommon.Address).Hex(),
		"value":     txData[4].(*big.Int).String(),
		"data_hex":  hex.EncodeToString(txData[5].([]byte)),
		"empty_v":   txData[6],
		"empty_r":   txData[7],
		"recipient": recipient.Hex(),
		"amount":    amount.String(),
	}).Info("Transaction components")

	rawTx, err := rlp.EncodeToBytes(txData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to RLP encode transaction: %v", err)
	}

	txHash := crypto.Keccak256(rawTx)

	p.logger.WithFields(logrus.Fields{
		"raw_tx_hex":   hex.EncodeToString(rawTx),
		"hash_to_sign": hex.EncodeToString(txHash),
	}).Info("Final transaction data")

	/*txBytes, err := hex.DecodeString(string(rawTx))
	if err != nil {
		p.logger.Errorf("Failed to decode transaction hex: %v", err)
		return []types.PluginKeysignRequest{}, fmt.Errorf("failed to decode transaction hex: %w", err)
	}*/
	//unmarshal tx from sign req.transaction
	tx := &gtypes.Transaction{}
	err = tx.UnmarshalBinary(rawTx)
	if err != nil {
		p.logger.Errorf("Failed to unmarshal transaction: %v", err)
		return "", nil, fmt.Errorf("failed to unmarshal transaction: %v: %w", err, asynq.SkipRetry)
	}
	fmt.Printf("Chain ID TEST 4: %s\n", tx.ChainId().String())

	return hex.EncodeToString(txHash), rawTx, nil
}

func (p *PayrollPlugin) Frontend() embed.FS {
	return frontend
}

func (p *PayrollPlugin) GetNextNonce(address string) (uint64, error) {
	return p.nonceManager.GetNextNonce(address)
}

func (p *PayrollPlugin) SigningComplete(ctx context.Context, signature tss.KeysignResponse, signRequest types.PluginKeysignRequest, policy types.PluginPolicy) error {
	R, S, V, originalTx, chainID, _, err := p.convertData(signature, signRequest, policy)
	if err != nil {
		return fmt.Errorf("failed to convert R and S: %v", err)
	}

	innerTx := &gtypes.LegacyTx{
		Nonce:    originalTx.Nonce(),
		GasPrice: originalTx.GasPrice(),
		Gas:      originalTx.Gas(),
		To:       originalTx.To(),
		Value:    originalTx.Value(),
		Data:     originalTx.Data(),
		V:        V,
		R:        R,
		S:        S,
	}

	signedTx := gtypes.NewTx(innerTx)
	signer := gtypes.NewLondonSigner(chainID)
	sender, err := signer.Sender(signedTx)
	if err != nil {
		p.logger.WithError(err).Warn("Could not determine sender")
	} else {
		p.logger.WithField("sender", sender.Hex()).Info("Transaction sender")
	}

	// Check if RPC client is initialized
	if p.rpcClient == nil {
		return fmt.Errorf("RPC client not initialized")
	}

	err = p.rpcClient.SendTransaction(ctx, signedTx)
	if err != nil {
		p.logger.WithError(err).Error("Failed to broadcast transaction")
		return p.handleBroadcastError(err, sender)
	}

	p.logger.WithField("hash", signedTx.Hash().Hex()).Info("Transaction successfully broadcast")

	return p.monitorTransaction(signedTx)
}

func (p *PayrollPlugin) handleBroadcastError(err error, sender gcommon.Address) error {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "insufficient funds"):
		// this is for ETH balance for gas - immediate failure, what to do?
		//goal : retry only when we dectect user send funds
		// for now : we can skip this trigger and wait for next one
		return &types.TransactionError{
			Code:    types.ErrInsufficientFunds,
			Message: fmt.Sprintf("Account %s has insufficient gas", sender.Hex()),
			Err:     err,
		}

	case strings.Contains(errMsg, "nonce too low"):
	case strings.Contains(errMsg, "nonce too high"):
	case strings.Contains(errMsg, "gas price too low"):
	case strings.Contains(errMsg, "gas limit reached"):
		// these are retriable errors - the caller should retry with updated parameters
		//we should not skip this trigger and retry immediately
		return &types.TransactionError{
			Code:    types.ErrRetriable,
			Message: err.Error(),
			Err:     err,
		}

	default:
		return &types.TransactionError{
			Code:    types.ErrRPCConnectionFailed,
			Message: "Unknown RPC error",
			Err:     err,
		}
	}
	return nil
}

func (p *PayrollPlugin) monitorTransaction(tx *gtypes.Transaction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) //how much time should we monitor the tx?
	defer cancel()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	txHash := tx.Hash()
	for {
		select {
		case <-ctx.Done():
			return &types.TransactionError{
				Code:    types.ErrTxTimeout,
				Message: fmt.Sprintf("Transaction monitoring timed out for tx: %s", txHash.Hex()),
			}

		case <-ticker.C:
			// check tx status
			_, isPending, err := p.rpcClient.TransactionByHash(ctx, txHash)
			if err != nil {
				if err == ethereum.NotFound {
					return &types.TransactionError{
						Code:    types.ErrTxDropped,
						Message: fmt.Sprintf("Transaction dropped from mempool: %s", txHash.Hex()),
					}
				}
				continue // keep trying on other RPC errors
			}

			if !isPending {
				receipt, err := p.rpcClient.TransactionReceipt(ctx, txHash)
				if err != nil {
					continue
				}

				if receipt.Status == 0 {
					reason := p.getRevertReason(ctx, tx, receipt.BlockNumber)

					// Check if it's a permanent failure (like insufficient token balance)
					if !p.isRetriableError(reason) {
						return &types.TransactionError{
							Code:    types.ErrPermanentFailure,
							Message: fmt.Sprintf("Transaction permanently failed: %s", reason),
						}
					}

					// It's a retriable error
					return &types.TransactionError{
						Code:    types.ErrRetriable,
						Message: fmt.Sprintf("Transaction failed with retriable error: %s", reason),
					}
				}

				// Transaction successful
				return nil
			}
		}
	}
}

func (p *PayrollPlugin) isRetriableError(reason string) bool {
	// implement logic to determine if the error is retriable based on the reason
	return strings.Contains(reason, "insufficient funds") || strings.Contains(reason, "nonce too low") || strings.Contains(reason, "nonce too high") || strings.Contains(reason, "gas price too low") || strings.Contains(reason, "gas limit reached")
}

func (p *PayrollPlugin) getRevertReason(ctx context.Context, tx *gtypes.Transaction, blockNum *big.Int) string {
	callMsg := ethereum.CallMsg{
		To:       tx.To(),
		Data:     tx.Data(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
	}

	_, err := p.rpcClient.CallContract(ctx, callMsg, blockNum)
	if err != nil {
		// try to parse standard revert reason
		if strings.Contains(err.Error(), "execution reverted:") {
			parts := strings.Split(err.Error(), "execution reverted:")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
		return err.Error()
	}
	return "Unknown revert reason"
}

func (p *PayrollPlugin) convertData(signature tss.KeysignResponse, signRequest types.PluginKeysignRequest, policy types.PluginPolicy) (R *big.Int, S *big.Int, V *big.Int, originalTx *gtypes.Transaction, chainID *big.Int, recoveryID int64, err error) {
	// convert R and S from hex strings to big.Int
	R = new(big.Int)
	R.SetString(signature.R, 16)
	if R == nil {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to parse R value")
	}

	S = new(big.Int)
	S.SetString(signature.S, 16)
	if S == nil {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to parse S value")
	}

	// Decode the hex string to bytes first
	txBytes, err := hex.DecodeString(signRequest.Transaction)
	if err != nil {
		p.logger.Errorf("Failed to decode transaction hex: %v", err)
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to decode transaction hex: %w", err)
	}

	originalTx = new(gtypes.Transaction)
	if err := rlp.DecodeBytes(txBytes, originalTx); err != nil {
		p.logger.Errorf("Failed to unmarshal transaction: %v", err)
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	policybytes := policy.Policy
	payrollPolicy := types.PayrollPolicy{}
	err = json.Unmarshal(policybytes, &payrollPolicy)
	if err != nil {
		p.logger.Errorf("Failed to unmarshal policy: %v", err)
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to unmarshal policy: %w", err)
	}
	chainID = new(big.Int)
	chainID.SetString(payrollPolicy.ChainID[0], 10)

	/*chainID = originalTx.ChainId()
	fmt.Printf("Chain ID TEST: %s\n", chainID.String())*/

	// calculate V according to EIP-155
	recoveryID, err = strconv.ParseInt(signature.RecoveryID, 10, 64)
	if err != nil {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to parse recovery ID: %w", err)
	}

	V = new(big.Int).Set(chainID)
	V.Mul(V, big.NewInt(2))
	V.Add(V, big.NewInt(35+recoveryID))

	return R, S, V, originalTx, chainID, recoveryID, nil
}

func DeriveAddress(compressedPubKeyHex, hexChainCode, derivePath string) (*gcommon.Address, error) {
	derivedPubKeyHex, err := tss.GetDerivedPubKey(compressedPubKeyHex, hexChainCode, derivePath, false)
	if err != nil {
		return nil, err
	}

	derivedPubKeyBytes, err := hex.DecodeString(derivedPubKeyHex)
	if err != nil {
		return nil, err
	}

	derivedPubKey, err := btcec.ParsePubKey(derivedPubKeyBytes)
	if err != nil {
		return nil, err
	}

	uncompressedPubKeyBytes := derivedPubKey.SerializeUncompressed()
	pubKeyBytesWithoutPrefix := uncompressedPubKeyBytes[1:]
	hash := crypto.Keccak256(pubKeyBytesWithoutPrefix)
	address := gcommon.BytesToAddress(hash[12:])

	return &address, nil
}
