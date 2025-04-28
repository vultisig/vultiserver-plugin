package payroll

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/vultisig/vultiserver-plugin/common"

	"github.com/google/uuid"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vultisig/mobile-tss-lib/tss"
	"github.com/vultisig/vultiserver-plugin/internal/types"
)

func (p *Plugin) ProposeTransactions(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error) {
	var txs []types.PluginKeysignRequest
	err := p.ValidatePluginPolicy(policy)
	if err != nil {
		return txs, fmt.Errorf("failed to validate plugin policy: %v", err)
	}

	var payrollPolicy Policy
	if err := json.Unmarshal(policy.Policy, &payrollPolicy); err != nil {
		return txs, fmt.Errorf("fail to unmarshal payroll policy, err: %w", err)
	}
	signerAddress, err := common.DeriveAddress(policy.GetPublicKey(), policy.ChainCodeHex, policy.DerivePath)
	if err != nil {
		return txs, fmt.Errorf("fail to derive address: %w", err)
	}

	for i, recipient := range payrollPolicy.Recipients {
		txHash, rawTx, err := p.generatePayrollTransaction(
			recipient.Amount,
			recipient.Address,
			payrollPolicy.ChainID[i],
			payrollPolicy.TokenID[i],
			signerAddress,
		)
		if err != nil {
			return []types.PluginKeysignRequest{}, fmt.Errorf("failed to generate transaction hash: %v", err)
		}

		chainIDInt, _ := strconv.ParseInt(payrollPolicy.ChainID[i], 10, 64)

		// Create signing request
		signRequest := types.PluginKeysignRequest{
			KeysignRequest: types.KeysignRequest{
				PublicKey:        policy.GetPublicKey(),
				Messages:         []string{hex.EncodeToString(txHash)},
				SessionID:        uuid.New().String(),
				HexEncryptionKey: common.HexEncryptionKey,
				DerivePath:       policy.DerivePath,
				IsECDSA:          IsECDSA(chainIDInt),
				VaultPassword:    common.VaultPassword,
				// Parties:          []string{common.PluginPartyID, common.VerifierPartyID},
			},
			Transaction: hex.EncodeToString(rawTx),
			PluginType:  policy.PluginType,
			PolicyID:    policy.ID,
		}
		txs = append(txs, signRequest)
	}

	return txs, nil
}

func (p *Plugin) generatePayrollTransaction(amountString, recipientString, chainID, tokenID string, signerAddress *gcommon.Address) ([]byte, []byte, error) {
	amount, ok := new(big.Int).SetString(amountString, 10)
	if !ok {
		return nil, nil, errors.New("fail to parse amount")
	}
	recipient := gcommon.HexToAddress(recipientString)
	tokenAddress := gcommon.HexToAddress(tokenID)

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	transferData, err := parsedABI.Pack("transfer", recipient, amount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to pack transfer data: %v", err)
	}

	// create call message to estimate gas
	callMsg := ethereum.CallMsg{
		From:  *signerAddress,
		To:    &tokenAddress,
		Data:  transferData,
		Value: big.NewInt(0),
	}
	// estimate gas limit
	gasLimit, err := p.rpcClient.EstimateGas(context.Background(), callMsg)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to estimate gas: %v", err)
	}

	// add 20% to gas limit for safety
	gasLimit += gasLimit * 20 / 100

	// get suggested gas price
	gasPrice, err := p.rpcClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get gas price: %v", err)
	}
	// Parse chain ID

	chainIDInt, ok := new(big.Int).SetString(chainID, 10)
	if !ok {
		return nil, nil, errors.New("fail to parse chain id")
	}

	nextNonce, err := p.GetNextNonce(signerAddress.Hex())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	tx := gtypes.NewTransaction(nextNonce, tokenAddress, big.NewInt(0), gasLimit, gasPrice, transferData)
	// Create unsigned transaction data
	V := new(big.Int).Set(chainIDInt)
	V = V.Mul(V, big.NewInt(2))
	V = V.Add(V, big.NewInt(35))
	txData := []interface{}{
		tx.Nonce(),    // nonce
		tx.GasPrice(), // gas price
		tx.Gas(),      // gas limit
		tx.To(),       // to
		tx.Value(),    // value
		tx.Data(),     // data
		V,             // chain id
		uint(0),       // empty r
		uint(0),       // empty s
	}

	rawTx, err := rlp.EncodeToBytes(txData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to RLP encode transaction: %v", err)
	}

	signer := gtypes.NewEIP155Signer(chainIDInt)
	txHash := signer.Hash(tx).Bytes()

	return txHash, rawTx, nil
}

func (p *Plugin) SigningComplete(ctx context.Context, signature tss.KeysignResponse, signRequest types.PluginKeysignRequest, policy types.PluginPolicy) error {
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

func (p *Plugin) convertData(signature tss.KeysignResponse, signRequest types.PluginKeysignRequest, policy types.PluginPolicy) (R *big.Int, S *big.Int, V *big.Int, originalTx *gtypes.Transaction, chainID *big.Int, recoveryID int64, err error) {
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
	payrollPolicy := Policy{}
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
