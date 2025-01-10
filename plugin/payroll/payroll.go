package payroll

import (
	"context"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
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
	logger       *logrus.Logger
}

func NewPayrollPlugin(db storage.DatabaseStorage, rpcClient *ethclient.Client) *PayrollPlugin {
	return &PayrollPlugin{
		db:           db,
		rpcClient:    rpcClient,
		nonceManager: plugin.NewNonceManager(rpcClient),
		logger:       logrus.New(),
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

	for _, recipient := range payrollPolicy.Recipients {
		txHash, rawTx, err := p.generatePayrollTransaction(recipient.Amount, recipient.Address, payrollPolicy.ChainID, payrollPolicy.TokenID)
		if err != nil {
			return []types.PluginKeysignRequest{}, fmt.Errorf("failed to generate transaction hash: %v", err)
		}

		// Create signing request
		signRequest := types.PluginKeysignRequest{
			KeysignRequest: types.KeysignRequest{
				PublicKey:        policy.PublicKey,
				Messages:         []string{txHash}, //check how to correctly construct tx hash which depends on blockchain infos like nounce
				SessionID:        uuid.New().String(),
				HexEncryptionKey: "0123456789abcdef0123456789abcdef",
				DerivePath:       "m/44/60/0/0/0",
				IsECDSA:          true,
				VaultPassword:    "your-secure-password",
			},
			Transaction: hex.EncodeToString(rawTx),
			PluginID:    policy.PluginID,
			PolicyID:    policy.ID,
		}
		txs = append(txs, signRequest)
	}

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

	for _, tx := range txs {
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

		if strings.ToLower(txDestination.Hex()) != strings.ToLower(payrollPolicy.TokenID) {
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

func (p *PayrollPlugin) generatePayrollTransaction(amountString string, recipientString string, chainID string, tokenID string) (string, []byte, error) {
	amount := new(big.Int)
	amount.SetString(amountString, 10)
	recipient := gcommon.HexToAddress(recipientString)

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// Create transfer data
	inputData, err := parsedABI.Pack("transfer", recipient, amount)
	if err != nil {
		return "", nil, fmt.Errorf("failed to pack transfer data: %v", err)
	}

	// Create transaction
	tx := gtypes.NewTransaction(
		0,                             // nonce  //TODO : to be updated.
		gcommon.HexToAddress(tokenID), // USDC contract
		big.NewInt(0),                 // value, if it is not eth. If it is eth, we have to set the value. How to tell to send eth at plugin creation?
		100000,                        // gas limit
		big.NewInt(2000000000),        // gas price (2 gwei)
		inputData,
	)

	// Get the raw transaction bytes
	rawTx, err := tx.MarshalBinary()
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal transaction: %v", err)
	}

	// Calculate transaction hash
	txHash := tx.Hash().Hex()[2:]

	return txHash, rawTx, nil

}

func (p *PayrollPlugin) Frontend() embed.FS {
	return frontend
}

func (p *PayrollPlugin) GetNextNonce(address string) (uint64, error) {
	return p.nonceManager.GetNextNonce(address)
}

func (p *PayrollPlugin) SigningComplete(signedTx types.SignedTransaction) error {
	tx := new(gtypes.Transaction)
	if err := tx.UnmarshalBinary([]byte(signedTx.RawTx)); err != nil {
		return fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	// get sender for logging
	signer := gtypes.NewLondonSigner(tx.ChainId())
	sender, err := signer.Sender(tx)
	if err != nil {
		p.logger.Warnf("Could not determine sender: %v", err)
	}

	// send tx
	err = p.rpcClient.SendTransaction(context.Background(), tx)
	if err != nil {
		return p.handleTransactionError(err, tx, sender) //Todo:the worker gets back the tx error, and acts accordingly
	}

	return p.monitorTransaction(tx)
}

func (p *PayrollPlugin) handleTransactionError(err error, tx *gtypes.Transaction, sender gcommon.Address) error {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "nonce too low"): //Todo : check that the error messages are correct
		return &types.TransactionError{
			Code:    types.ErrNonce,
			Message: fmt.Sprintf("Transaction with nonce %d already exists. Retry with new nonce", tx.Nonce()),
			Err:     err,
		}

	case strings.Contains(errMsg, "nonce too high"):
		return &types.TransactionError{
			Code:    types.ErrNonce,
			Message: "Gap in nonces detected. Previous transaction might be pending",
			Err:     err,
		}

	case strings.Contains(errMsg, "insufficient funds"):
		return &types.TransactionError{
			Code:    types.ErrInsufficientFunds,
			Message: fmt.Sprintf("Account %s has insufficient funds", sender.Hex()),
			Err:     err,
		}

	case strings.Contains(errMsg, "gas price too low"):
		return &types.TransactionError{
			Code:    types.ErrGasPriceUnderpriced,
			Message: fmt.Sprintf("Current gas price %s is too low", tx.GasPrice().String()),
			Err:     err,
		}

	case strings.Contains(errMsg, "gas limit reached"):
		return &types.TransactionError{
			Code:    types.ErrGasTooLow,
			Message: fmt.Sprintf("Gas limit %d is too low", tx.Gas()),
			Err:     err,
		}

	default:
		return &types.TransactionError{
			Code:    types.ErrRPCConnectionFailed,
			Message: "Unknown RPC error",
			Err:     err,
		}
	}
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
					// try to get revert reason
					reason := p.getRevertReason(ctx, tx, receipt.BlockNumber)
					return &types.TransactionError{
						Code:    types.ErrExecutionReverted,
						Message: fmt.Sprintf("Transaction reverted: %s", reason),
					}
				}

				// Transaction successful
				return nil
			}
		}
	}
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
