package dca

import (
	"context"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/mobile-tss-lib/tss"
	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/signing"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
	"github.com/vultisig/vultisigner/uniswap"
)

const (
	PLUGIN_TYPE    = "dca"
	PLUGIN_VERSION = "0.0.1"
	POLICY_VERSION = "0.0.1"
)

type DCAPlugin struct {
	uniswapClient *uniswap.Client
	rpcClient     *ethclient.Client
	db            storage.DatabaseStorage
	logger        *logrus.Logger
}

func NewDCAPlugin(uniswapCfg *uniswap.Config, db storage.DatabaseStorage, logger *logrus.Logger) (*DCAPlugin, error) {
	pluginConfig, err := config.ReadConfig("config-plugin")
	if err != nil {
		return nil, fmt.Errorf("fail to read plugin config: %w", err)
	}

	rpcClient, err := ethclient.Dial(pluginConfig.Server.Plugin.Eth.Rpc)
	if err != nil {
		return nil, fmt.Errorf("fail to connect to RPC client: %w", err)
	}

	uniswapClient, err := uniswap.NewClient(uniswapCfg)
	if err != nil {
		return nil, fmt.Errorf("fail to initialize Uniswap client: %w", err)
	}

	return &DCAPlugin{
		uniswapClient: uniswapClient,
		rpcClient:     rpcClient,
		db:            db,
		logger:        logger,
	}, nil
}

// TODO: do we actually need this?
func (p *DCAPlugin) SignPluginMessages(e echo.Context) error {
	p.logger.Debug("DCA: SIGN PLUGIN MESSAGES")
	return nil
}

func (p *DCAPlugin) SigningComplete(
	ctx context.Context,
	signature tss.KeysignResponse,
	signRequest types.PluginKeysignRequest,
	policy types.PluginPolicy,
) error {
	var dcaPolicy types.DCAPolicy
	if err := json.Unmarshal(policy.Policy, &dcaPolicy); err != nil {
		return fmt.Errorf("fail to unmarshal DCA policy: %w", err)
	}

	chainID, ok := new(big.Int).SetString(dcaPolicy.ChainID, 10)
	if !ok {
		return errors.New("fail to parse chain ID")
	}

	// currently we are only signing one transaction
	txHash := signRequest.Messages[0]
	if len(txHash) == 0 {
		return errors.New("transaction hash is missing")
	}

	signedTx, _, err := signing.SignLegacyTx(signature, txHash, signRequest.Transaction, chainID)
	if err != nil {
		return fmt.Errorf("fail to sign transaction: %w", err)
	}

	err = p.rpcClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	receipt, err := bind.WaitMined(context.Background(), p.rpcClient, signedTx)
	if err != nil {
		return fmt.Errorf("fail to wait for transaction to be mined: %w", err)
	}

	p.logger.Info("transaction receipt status: ", receipt.Status)
	return nil
}

func (p *DCAPlugin) SetupPluginPolicy(policyDoc *types.PluginPolicy) error {
	if policyDoc == nil {
		return fmt.Errorf("no policy to set up")
	}

	if policyDoc.ID == "" {
		policyDoc.ID = uuid.NewString()
	}

	if policyDoc.PolicyVersion == "" {
		policyDoc.PolicyVersion = POLICY_VERSION
	}

	if policyDoc.PluginVersion == "" {
		policyDoc.PluginVersion = PLUGIN_VERSION
	}

	if policyDoc.PluginID == "" {
		policyDoc.PluginID = uuid.NewString()
	}

	return nil
}

func (p *DCAPlugin) ValidatePluginPolicy(policyDoc types.PluginPolicy) error {
	if policyDoc.PluginType != PLUGIN_TYPE {
		return fmt.Errorf("policy does not match plugin type, expected: %s, got: %s", PLUGIN_TYPE, policyDoc.PluginType)
	}

	var dcaPolicy types.DCAPolicy
	if err := json.Unmarshal(policyDoc.Policy, &dcaPolicy); err != nil {
		return fmt.Errorf("fail to unmarshal DCA policy: %w", err)
	}

	mixedCaseTokenIn, err := gcommon.NewMixedcaseAddressFromString(dcaPolicy.SourceTokenID)
	if err != nil {
		return fmt.Errorf("invalid source token address: %s", dcaPolicy.SourceTokenID)
	}
	if strings.ToLower(dcaPolicy.SourceTokenID) != dcaPolicy.SourceTokenID {
		if !mixedCaseTokenIn.ValidChecksum() {
			return fmt.Errorf("invalid source token address checksum: %s", dcaPolicy.SourceTokenID)
		}
	}

	mixedCaseTokenOut, err := gcommon.NewMixedcaseAddressFromString(dcaPolicy.DestinationTokenID)
	if err != nil {
		return fmt.Errorf("invalid destination token address: %s", dcaPolicy.DestinationTokenID)
	}
	if strings.ToLower(dcaPolicy.DestinationTokenID) != dcaPolicy.DestinationTokenID {
		if !mixedCaseTokenOut.ValidChecksum() {
			return fmt.Errorf("invalid destination token address checksum: %s", dcaPolicy.DestinationTokenID)
		}
	}

	sourceAddrPolicy := gcommon.HexToAddress(dcaPolicy.SourceTokenID)
	destAddrPolicy := gcommon.HexToAddress(dcaPolicy.DestinationTokenID)
	if sourceAddrPolicy == gcommon.HexToAddress("0x0") || destAddrPolicy == gcommon.HexToAddress("0x0") {
		return fmt.Errorf("invalid token addresses")
	}

	if dcaPolicy.SourceTokenID == dcaPolicy.DestinationTokenID {
		return fmt.Errorf("source token and destination token addresses are the same")
	}

	if dcaPolicy.TotalAmount == "" {
		return fmt.Errorf("total amount is required")
	}
	totalAmount, ok := new(big.Int).SetString(dcaPolicy.TotalAmount, 10)
	if !ok {
		return fmt.Errorf("invalid total amount %s", dcaPolicy.TotalAmount)
	}
	if totalAmount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("total amount must be greater than 0")
	}

	if dcaPolicy.TotalOrders == "" {
		return fmt.Errorf("total orders is required")
	}
	totalOrders, ok := new(big.Int).SetString(dcaPolicy.TotalOrders, 10)
	if !ok {
		return fmt.Errorf("invalid total orders %s", dcaPolicy.TotalOrders)
	}
	if totalOrders.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("total orders must be greater than 0")
	}

	if dcaPolicy.PriceRange.Min != "" && dcaPolicy.PriceRange.Max != "" {
		minPrice, ok := new(big.Int).SetString(dcaPolicy.PriceRange.Min, 10)
		if !ok {
			return fmt.Errorf("invalid min price %s", dcaPolicy.PriceRange.Min)
		}
		maxPrice, ok := new(big.Int).SetString(dcaPolicy.PriceRange.Max, 10)
		if !ok {
			return fmt.Errorf("invalid max price %s", dcaPolicy.PriceRange.Max)
		}
		if minPrice.Cmp(maxPrice) > 0 {
			return fmt.Errorf("min price should be equal or lower than max price")
		}
	}

	if dcaPolicy.ChainID == "" {
		return fmt.Errorf("chain id is required")
	}

	return nil
}

// TODO: do we actually need this?
func (p *DCAPlugin) ConfigurePlugin(e echo.Context) error {
	return nil
}

func (p *DCAPlugin) ProposeTransactions(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error) {
	p.logger.Debug("DCA: PROPOSE TRANSACTIONS")

	var txs []types.PluginKeysignRequest

	// validate policy
	err := p.ValidatePluginPolicy(policy)
	if err != nil {
		return txs, fmt.Errorf("fail to validate plugin policy: %w", err)
	}

	var dcaPolicy types.DCAPolicy
	if err := json.Unmarshal(policy.Policy, &dcaPolicy); err != nil {
		return txs, fmt.Errorf("fail to unmarshal dca policy, err: %w", err)
	}

	// build transactions
	// TODO: get theses values from the vault
	derivePath := "m/44'/60'/0'/0/0"                                                       // ethereum
	hexChainCode := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"     // vault's chain code
	hexEncryptionKey := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef" // vault's encryption key
	vaultPassword := "your-secure-password"                                                // vault's password

	signerAddress, err := common.DeriveAddress(policy.PublicKey, hexChainCode, derivePath)
	if err != nil {
		return []types.PluginKeysignRequest{}, fmt.Errorf("fail to derive address: %w", err)
	}

	chainID, ok := new(big.Int).SetString(dcaPolicy.ChainID, 10)
	if !ok {
		return []types.PluginKeysignRequest{}, fmt.Errorf("fail to parse chain id: %w", err)
	}
	rawTxsData, err := p.generateSwapTransactions(chainID, signerAddress, dcaPolicy.SourceTokenID, dcaPolicy.DestinationTokenID, dcaPolicy.TotalAmount)
	if err != nil {
		return []types.PluginKeysignRequest{}, fmt.Errorf("fail to generate transaction hash: %w", err)
	}

	for _, data := range rawTxsData {
		signRequest := types.PluginKeysignRequest{
			KeysignRequest: types.KeysignRequest{
				PublicKey:        policy.PublicKey,
				Messages:         []string{hex.EncodeToString(data.TxHash)},
				SessionID:        uuid.New().String(),
				HexEncryptionKey: hexEncryptionKey,
				DerivePath:       derivePath,
				IsECDSA:          true, // TODO:
				VaultPassword:    vaultPassword,
				StartSession:     false,
				Parties:          []string{common.PluginPartyID, common.VerifierPartyID},
			},
			Transaction: hex.EncodeToString(data.RlpTxBytes),
			PluginID:    policy.PluginID,
			PolicyID:    policy.ID,
		}
		txs = append(txs, signRequest)
	}

	return txs, nil
}

func (p *DCAPlugin) ValidateTransactionProposal(policy types.PluginPolicy, txs []types.PluginKeysignRequest) error {
	p.logger.Info("DCA: VALIDATE TRANSACTION PROPOSAL")

	if len(txs) == 0 {
		return fmt.Errorf("no transactions provided for validation")
	}
	// TODO: get these values from the vault.
	hexChainCode := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef" // vault's chain code
	derivePath := "m/44'/60'/0'/0/0"                                                   // ethereum

	if err := p.ValidatePluginPolicy(policy); err != nil {
		return fmt.Errorf("failed to validate plugin policy: %w", err)
	}

	var dcaPolicy types.DCAPolicy
	if err := json.Unmarshal(policy.Policy, &dcaPolicy); err != nil {
		return fmt.Errorf("failed to unmarshal DCA policy: %w", err)
	}

	// Validate policy params.
	policyChainID, ok := new(big.Int).SetString(dcaPolicy.ChainID, 10)
	if !ok {
		return fmt.Errorf("failed to parse chain ID: %s", dcaPolicy.ChainID)
	}

	sourceAddrPolicy := gcommon.HexToAddress(dcaPolicy.SourceTokenID)
	destAddrPolicy := gcommon.HexToAddress(dcaPolicy.DestinationTokenID)

	totalAmount, ok := new(big.Int).SetString(dcaPolicy.TotalAmount, 10)
	if !ok {
		return fmt.Errorf("invalid total amount")
	}

	signerAddress, err := common.DeriveAddress(policy.PublicKey, hexChainCode, derivePath)
	if err != nil {
		return fmt.Errorf("failed to derive address: %w", err)
	}
	// Validate each transaction
	for _, tx := range txs {
		if err := p.validateTransaction(tx, totalAmount, policyChainID, &sourceAddrPolicy, &destAddrPolicy, signerAddress); err != nil {
			return fmt.Errorf("failed to validate transaction: %w", err)
		}
	}
	return nil
}

func (p *DCAPlugin) validateTransaction(keysignRequest types.PluginKeysignRequest, totalAmountPolicy, policyChainID *big.Int, sourceAddrPolicy, destAddrPolicy, signerAddress *gcommon.Address) error {
	// Parse the transaction
	var tx *gtypes.Transaction
	txBytes, err := hex.DecodeString(keysignRequest.Transaction)
	if err != nil {
		return fmt.Errorf("failed to decode transaction bytes: %w", err)
	}
	err = rlp.DecodeBytes(txBytes, &tx)
	if err != nil {
		return fmt.Errorf("fail to parse RLP transaction: %w", err)
	}

	// Validate chain ID
	if tx.ChainId().Cmp(policyChainID) != 0 {
		return fmt.Errorf("chain ID mismatch: expected %s, got %s", policyChainID.String(), tx.ChainId().String())
	}

	// Validate destination address
	txDestination := tx.To()
	if txDestination == nil {
		return fmt.Errorf("transaction missing destination address")
	}
	if p.uniswapClient.GetRouterAddress() != txDestination {
		// TODO: change when mint and approve transactions are removed.
		p.logger.Warn("invalid router address", "address", txDestination)
	}

	// Validate gas parameters
	if tx.Gas() == 0 {
		return fmt.Errorf("invalid gas limit: must be greater than zero")
	}
	if tx.GasPrice().Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("invalid gas price: must be greater than zero")
	}

	// Validate transaction data
	if len(tx.Data()) == 0 {
		return fmt.Errorf("transaction contains empty payload")
	}

	parsedRouterABI, err := p.getRouterABI()
	if err != nil {
		return fmt.Errorf("failed to parse router ABI: %w", err)
	}
	method, err := parsedRouterABI.MethodById(tx.Data())
	if err != nil {
		p.logger.Warn("failed to find method in router ABI")
	}
	if method != nil && method.Name != "swapExactTokensForTokens" {
		return fmt.Errorf("unexpected transaction method: expected 'swapExactTokensForTokens', got %s'", method.Name)
	}

	// Validate swap parameters if it's a swap transaction
	if method != nil && method.Name == "swapExactTokensForTokens" {
		p.logger.Info("DCA: method is swapExactTokensForTokens")
		if err := p.validateSwapParameters(tx, method, totalAmountPolicy, sourceAddrPolicy, destAddrPolicy, signerAddress); err != nil {
			return fmt.Errorf("failed to validate swap parameters: %w", err)
		}
	}
	return nil
}
func (p *DCAPlugin) validateSwapParameters(tx *gtypes.Transaction, method *abi.Method, totalAmountPolicy *big.Int, sourceAddrPolicy, destAddrPolicy, signerAddress *gcommon.Address) error {
	inputData := tx.Data()[4:]
	decodedParams, err := method.Inputs.Unpack(inputData)
	if err != nil {
		return fmt.Errorf("failed to decode transaction swap parameters: %w", err)
	}
	path, ok := decodedParams[2].([]gcommon.Address)
	if !ok || len(path) < 2 {
		return fmt.Errorf("invalid swap path: must contain at least 2 tokens")
	}

	if path[0] != *sourceAddrPolicy || path[len(path)-1] != *destAddrPolicy {
		return fmt.Errorf("swap path tokens mismatch: expected source=%s, destination=%s", *sourceAddrPolicy, *destAddrPolicy)
	}

	// TODO: change this validation to not compare the total amount, but to validate it is in valid range.
	amountIn, ok := decodedParams[0].(*big.Int)
	if !ok {
		return fmt.Errorf("failed to parse swap amount: invalid format")
	}
	if amountIn.Cmp(totalAmountPolicy) > 0 {
		return fmt.Errorf("swap amount exceeds policy limit: max=%s, got=%s", totalAmountPolicy.String(), amountIn.String())
	}

	// Validate destination address matches signer
	to, ok := decodedParams[3].(gcommon.Address)
	if !ok || to != *signerAddress {
		return fmt.Errorf("invalid swap destination: expected=%s, got=%s", *signerAddress, to.String())
	}

	return nil
}
func (p *DCAPlugin) getRouterABI() (abi.ABI, error) {
	routerABI := `[
        {
            "name": "swapExactTokensForTokens",
            "type": "function",
            "inputs": [
                {
                    "name": "amountIn",
                    "type": "uint256"
                },
                {
                    "name": "amountOutMin",
                    "type": "uint256"
                },
                {
                    "name": "path",
                    "type": "address[]"
                },
                {
                    "name": "to",
                    "type": "address"
                },
                {
                    "name": "deadline",
                    "type": "uint256"
                }
            ]
        }
    ]`
	return abi.JSON(strings.NewReader(routerABI))
}

func (p *DCAPlugin) Frontend() embed.FS {
	return embed.FS{}
}

type RawTxData struct {
	TxHash     []byte
	RlpTxBytes []byte
}

func (p *DCAPlugin) generateSwapTransactions(chainID *big.Int, signerAddress *gcommon.Address, srcToken, destToken string, strAmount string) ([]RawTxData, error) {
	srcTokenAddress := gcommon.HexToAddress(srcToken)
	destTokenAddress := gcommon.HexToAddress(destToken)
	tokensPair := []gcommon.Address{srcTokenAddress, destTokenAddress}

	swapAmountIn, ok := new(big.Int).SetString(strAmount, 10)
	if !ok {
		return []RawTxData{}, fmt.Errorf("fail to parse swap amount")
	}

	// fetch token pair amount out
	expectedAmountOut, err := p.uniswapClient.GetExpectedAmountOut(swapAmountIn, tokensPair)
	if err != nil {
		return []RawTxData{}, fmt.Errorf("fail to get expected amount out: %w", err)
	}
	p.logger.Info("Expected amount out: ", expectedAmountOut.String())

	// TODO: remove, probably we dont need slippage percentage
	slippagePercentage := 1.0
	amountOutMin := p.uniswapClient.CalculateAmountOutMin(expectedAmountOut, slippagePercentage)

	// TODO: validate the price range (if specified)

	rawTxsData := []RawTxData{}

	// TODO:
	// remove, it is not responsibility of the plugin
	// the user should have the coresponding amount of WETH
	// in their wallet

	// mint WETH
	p.logger.Info("Minting WETH")
	p.logTokenBalances(p.uniswapClient, signerAddress, srcTokenAddress, destTokenAddress)
	txHash, rawTx, err := p.uniswapClient.MintWETH(chainID, signerAddress, swapAmountIn, srcTokenAddress)
	if err != nil {
		return []RawTxData{}, fmt.Errorf("fail to mint WETH: %w", err)
	}
	rawTxsData = append(rawTxsData, RawTxData{txHash, rawTx})
	p.logTokenBalances(p.uniswapClient, signerAddress, srcTokenAddress, destTokenAddress)

	// TODO:
	// approve should be done by the Vault user, during policy
	// configuration, since it is not responsibility of the plugin
	// to approve tokens so there will be only one transaction to sign

	p.logger.Info("Approving Uniswap Router to spend: ", srcTokenAddress.Hex())
	// approve Router to spend input token
	txHash, rawTx, err = p.uniswapClient.ApproveERC20Token(chainID, signerAddress, srcTokenAddress, *p.uniswapClient.GetRouterAddress(), swapAmountIn)
	if err != nil {
		return []RawTxData{}, fmt.Errorf("fail to approve token: %w", err)
	}
	rawTxsData = append(rawTxsData, RawTxData{txHash, rawTx})

	// swap tokens
	txHash, rawTx, err = p.uniswapClient.SwapTokens(chainID, signerAddress, swapAmountIn, amountOutMin, tokensPair)
	if err != nil {
		return []RawTxData{}, fmt.Errorf("fail to swap tokens: %w", err)
	}
	rawTxsData = append(rawTxsData, RawTxData{txHash, rawTx})
	p.logTokenBalances(p.uniswapClient, signerAddress, srcTokenAddress, destTokenAddress)

	return rawTxsData, nil
}

func (p *DCAPlugin) logTokenBalances(client *uniswap.Client, signerAddress *gcommon.Address, tokenInAddress, tokenOutAddress gcommon.Address) {
	tokenInBalance, err := client.GetTokenBalance(signerAddress, tokenInAddress)
	if err != nil {
		p.logger.Error("Input token balance: ", err)
		return
	}
	p.logger.Info("Input token balance: ", tokenInBalance.String())

	tokenOutBalance, err := client.GetTokenBalance(signerAddress, tokenOutAddress)
	if err != nil {
		p.logger.Error("Output token balance: ", err)
		return
	}
	p.logger.Info("Output token balance: ", tokenOutBalance.String())
}
