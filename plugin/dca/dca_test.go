package dca

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vultisig/mobile-tss-lib/tss"
	"github.com/vultisig/vultiserver-plugin/common"
	"github.com/vultisig/vultiserver-plugin/internal/types"
	"github.com/vultisig/vultiserver-plugin/test/mocks/database"
	"github.com/vultisig/vultiserver-plugin/test/mocks/ethclient"
	"github.com/vultisig/vultiserver-plugin/test/mocks/uniswapclient"

	pg "github.com/vultisig/vultiserver-plugin/plugin"

	"math/big"
	"testing"
)

func createSamplePolicy(progress string) types.PluginPolicy {
	return types.PluginPolicy{
		Progress: progress,
	}
}

func createValidPolicy() types.PluginPolicy {
	dcaPolicy := Policy{
		ChainID:            "1",
		SourceTokenID:      "0x1111111111111111111111111111111111111111",
		DestinationTokenID: "0x2222222222222222222222222222222222222222",
		TotalAmount:        "1000",
		TotalOrders:        "10",
		Schedule: pg.Schedule{
			Frequency: "daily",
			Interval:  "1",
		},
	}

	policyBytes, _ := json.Marshal(dcaPolicy)

	return types.PluginPolicy{
		ID:            "098c8562-91bd-46da-92a2-d66b38823cb0",
		PluginType:    PluginType,
		PluginVersion: PluginVersion,
		PolicyVersion: PolicyVersion,
		ChainCodeHex:  "8769353fa9b5baaf9ceef4c0c747c57d67933ed9865612ce5d8b771708bfaa1d",
		PublicKey:     "02e23a52d46f02064f60305a5397ed808f4e2dcc4210a3ddc1c4ca9a6ac6d02fb3",
		DerivePath:    common.DerivePathMap["1"],
		IsEcdsa:       true,
		Policy:        policyBytes,
		Active:        true,
		Progress:      "IN PROGRESS",
	}
}

func rlpUnsignedTxAndHash(tx *gtypes.Transaction, chainID *big.Int) ([]byte, []byte, error) {
	// post EIP-155 transaction
	V := new(big.Int).Set(chainID)
	V = V.Mul(V, big.NewInt(2))
	V = V.Add(V, big.NewInt(35))
	rawTx, err := rlp.EncodeToBytes([]interface{}{
		tx.Nonce(),
		tx.GasPrice(),
		tx.Gas(),
		tx.To(),
		tx.Value(),
		tx.Data(),
		V,       // chain id
		uint(0), // r
		uint(0), // s
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to rlp encode transaction: %v", err)
	}

	signer := gtypes.NewEIP155Signer(chainID)
	txHash := signer.Hash(tx).Bytes()

	return txHash, rawTx, nil
}

func createSwapTransaction(t *testing.T, chainID *big.Int, amountIn *big.Int, amountOutMin *big.Int, path []gcommon.Address, to gcommon.Address, routerAddress gcommon.Address) ([]byte, []byte, error) {
	plugin := Plugin{}
	parsedABI, err := plugin.getSwapABI()
	require.NoError(t, err)

	deadline := big.NewInt(1714500000)
	data, err := parsedABI.Pack("swapExactTokensForTokens", amountIn, amountOutMin, path, to, deadline)
	require.NoError(t, err)
	gasPrice := big.NewInt(2000000000)
	gasLimit := uint64(200000)
	tx := gtypes.NewTransaction(0, routerAddress, big.NewInt(0), gasLimit, gasPrice, data)

	hash, rawTx, err := rlpUnsignedTxAndHash(tx, chainID)
	require.NoError(t, err)
	return hash, rawTx, nil
}
func createApproveTransaction(t *testing.T, chainID *big.Int, spender gcommon.Address, amount *big.Int, tokenAddress gcommon.Address) ([]byte, []byte, error) {
	plugin := Plugin{}
	parsedABI, err := plugin.getApproveABI()
	require.NoError(t, err)
	data, err := parsedABI.Pack("approve", spender, amount)
	require.NoError(t, err)

	gasPrice := big.NewInt(2000000000) // 2 Gwei
	gasLimit := uint64(60000)

	tx := gtypes.NewTransaction(0, tokenAddress, big.NewInt(0), gasLimit, gasPrice, data)

	hash, rawTx, err := rlpUnsignedTxAndHash(tx, chainID)
	require.NoError(t, err)
	return hash, rawTx, nil
}
func createInvalidTransaction(t *testing.T, chainID *big.Int, invalidCase string) ([]byte, []byte, error) {
	gasPrice := big.NewInt(2000000000) // 2 Gwei
	gasLimit := uint64(60000)

	dummyAddr := gcommon.HexToAddress("0x1111111111111111111111111111111111111111")

	var txHash, rawTx []byte
	var err error
	switch invalidCase {
	case "empty_data":
		tx := gtypes.NewTransaction(0, dummyAddr, big.NewInt(0), gasLimit, gasPrice, nil)
		txHash, rawTx, err = rlpUnsignedTxAndHash(tx, chainID)
	case "wrong_chain_id":
		tx := gtypes.NewTransaction(0, dummyAddr, big.NewInt(0), gasLimit, gasPrice, nil)
		txHash, rawTx, err = rlpUnsignedTxAndHash(tx, big.NewInt(2))
	case "zero_gas_price":
		tx := gtypes.NewTransaction(0, dummyAddr, big.NewInt(0), gasLimit, big.NewInt(0), nil)
		txHash, rawTx, err = rlpUnsignedTxAndHash(tx, chainID)
	case "zero_gas_limit":
		tx := gtypes.NewTransaction(0, dummyAddr, big.NewInt(0), uint64(0), gasPrice, nil)
		txHash, rawTx, err = rlpUnsignedTxAndHash(tx, chainID)
	}

	require.NoError(t, err)
	return txHash, rawTx, nil
}
func createKeysignRequest(txHash []byte, rlpTxBytes []byte, policyID string) types.PluginKeysignRequest {
	return types.PluginKeysignRequest{
		KeysignRequest: types.KeysignRequest{
			PublicKey:        "02e23a52d46f02064f60305a5397ed808f4e2dcc4210a3ddc1c4ca9a6ac6d02fb3",
			Messages:         []string{hex.EncodeToString(txHash)},
			SessionID:        "sessionID",
			HexEncryptionKey: "8769353fa9b5baaf9ceef4c0c747c57d67933ed9865612ce5d8b771708bfaa1d",
			DerivePath:       "m/44'/60'/0'/0/0",
			IsECDSA:          true,
			VaultPassword:    "password",
			// StartSession:     false,
			Parties: []string{"party1", "party2"},
		},
		Transaction:     hex.EncodeToString(rlpTxBytes),
		PluginID:        "pluginID",
		PolicyID:        policyID,
		TransactionType: "SWAP",
	}
}

// Tests

func TestValidatePluginPolicy(t *testing.T) {
	mockUniswap := new(uniswapclient.MockUniswapClient)
	mockEth := new(ethclient.MockEthClient)
	mockDB := new(database.MockDB)
	logger := logrus.StandardLogger()

	plugin := &Plugin{
		uniswapClient: mockUniswap,
		db:            mockDB,
		logger:        logger,
		rpcClient:     mockEth,
	}

	testCases := []struct {
		name      string
		policy    types.PluginPolicy
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Valid policy",
			policy:    createValidPolicy(),
			wantError: false,
		},
		{
			name: "Invalid plugin type",
			policy: func() types.PluginPolicy {
				p := createValidPolicy()
				p.PluginType = "invalid"
				return p
			}(),
			wantError: true,
			errorMsg:  "policy does not match plugin type",
		},
		{
			name: "Invalid plugin version",
			policy: func() types.PluginPolicy {
				p := createValidPolicy()
				p.PluginVersion = "invalid"
				return p
			}(),
			wantError: true,
			errorMsg:  "policy does not match plugin version",
		},
		{
			name: "Invalid policy version",
			policy: func() types.PluginPolicy {
				p := createValidPolicy()
				p.PolicyVersion = "invalid"
				return p
			}(),
			wantError: true,
			errorMsg:  "policy does not match policy version",
		},
		{
			name: "Missing chain code",
			policy: func() types.PluginPolicy {
				p := createValidPolicy()
				p.ChainCodeHex = ""
				return p
			}(),
			wantError: true,
			errorMsg:  "policy does not contain chain_code_hex",
		},
		{
			name: "Missing public key",
			policy: func() types.PluginPolicy {
				p := createValidPolicy()
				p.PublicKey = ""
				return p
			}(),
			wantError: true,
			errorMsg:  "policy does not contain public_key",
		},
		{
			name: "Invalid source and destination tokens",
			policy: func() types.PluginPolicy {
				p := createValidPolicy()
				dcaPolicy := Policy{
					ChainID:            "1",
					SourceTokenID:      "0x1111111111111111111111111111111111111111",
					DestinationTokenID: "0x1111111111111111111111111111111111111111", // Same as source
					TotalAmount:        "1000",
					TotalOrders:        "10",
					Schedule: pg.Schedule{
						Frequency: "daily",
						Interval:  "1",
					},
				}
				p.Policy, _ = json.Marshal(dcaPolicy)
				return p
			}(),
			wantError: true,
			errorMsg:  "source token and destination token addresses are the same",
		},
		{
			name: "Invalid total amount",
			policy: func() types.PluginPolicy {
				p := createValidPolicy()
				dcaPolicy := Policy{
					ChainID:            "1",
					SourceTokenID:      "0x1111111111111111111111111111111111111111",
					DestinationTokenID: "0x2222222222222222222222222222222222222222",
					TotalAmount:        "-1",
					TotalOrders:        "10",
					Schedule: pg.Schedule{
						Frequency: "daily",
						Interval:  "1",
					},
				}
				p.Policy, _ = json.Marshal(dcaPolicy)
				return p
			}(),
			wantError: true,
			errorMsg:  "total amount must be greater than 0",
		},
		{
			name: "Invalid total orders",
			policy: func() types.PluginPolicy {
				p := createValidPolicy()
				dcaPolicy := Policy{
					ChainID:            "1",
					SourceTokenID:      "0x1111111111111111111111111111111111111111",
					DestinationTokenID: "0x2222222222222222222222222222222222222222",
					TotalAmount:        "1000",
					TotalOrders:        "0",
					Schedule: pg.Schedule{
						Frequency: "daily",
						Interval:  "1",
					},
				}
				p.Policy, _ = json.Marshal(dcaPolicy)
				return p
			}(),
			wantError: true,
			errorMsg:  "total orders must be greater than 0",
		},
		{
			name: "Invalid price range",
			policy: func() types.PluginPolicy {
				p := createValidPolicy()
				dcaPolicy := Policy{
					ChainID:            "1",
					SourceTokenID:      "0x1111111111111111111111111111111111111111",
					DestinationTokenID: "0x2222222222222222222222222222222222222222",
					TotalAmount:        "1000",
					TotalOrders:        "10",
					PriceRange: PriceRange{
						Min: "200",
						Max: "100", // Min > Max
					},
					Schedule: pg.Schedule{
						Frequency: "daily",
						Interval:  "1",
					},
				}
				p.Policy, _ = json.Marshal(dcaPolicy)
				return p
			}(),
			wantError: true,
			errorMsg:  "min price should be equal or lower than max price",
		},
		{
			name: "Invalid derive path",
			policy: func() types.PluginPolicy {
				p := createValidPolicy()
				p.DerivePath = "wrong/path"
				return p
			}(),
			wantError: true,
			errorMsg:  "policy does not match derive path",
		},
		{
			name: "Invalid frequency",
			policy: func() types.PluginPolicy {
				p := createValidPolicy()
				dcaPolicy := Policy{
					ChainID:            "1",
					SourceTokenID:      "0x1111111111111111111111111111111111111111",
					DestinationTokenID: "0x2222222222222222222222222222222222222222",
					TotalAmount:        "1000",
					TotalOrders:        "10",
					Schedule: pg.Schedule{
						Frequency: "invalid",
						Interval:  "1",
					},
				}
				p.Policy, _ = json.Marshal(dcaPolicy)
				return p
			}(),
			wantError: true,
			errorMsg:  "invalid frequency",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := plugin.ValidatePluginPolicy(tc.policy)

			if tc.wantError {
				require.Error(t, err)
				if tc.errorMsg != "" {
					require.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCalculateSwapAmountPerOrder(t *testing.T) {
	mockUniswap := new(uniswapclient.MockUniswapClient)
	mockEth := new(ethclient.MockEthClient)
	mockDB := new(database.MockDB)
	logger := logrus.StandardLogger()

	plugin := &Plugin{
		uniswapClient: mockUniswap,
		db:            mockDB,
		logger:        logger,
		rpcClient:     mockEth,
	}

	testCases := []struct {
		name           string
		totalAmount    *big.Int
		totalOrders    *big.Int
		completedSwaps int64
		expected       *big.Int
	}{
		{
			name:           "Valid swap amount",
			totalAmount:    big.NewInt(100),
			totalOrders:    big.NewInt(1),
			completedSwaps: 0,
			expected:       big.NewInt(100),
		},
		{
			name:           "With remainder - first swap",
			totalAmount:    big.NewInt(1001),
			totalOrders:    big.NewInt(10),
			completedSwaps: 0,
			expected:       big.NewInt(101),
		},
		{
			name:           "With remainder - later swap no remainder case",
			totalAmount:    big.NewInt(1003),
			totalOrders:    big.NewInt(10),
			completedSwaps: 5,
			expected:       big.NewInt(100),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := plugin.calculateSwapAmountPerOrder(tc.totalAmount, tc.totalOrders, tc.completedSwaps)
			require.Equal(t, tc.expected, result)
		})
	}

}

func TestGetCompletedSwapTransactionsCount(t *testing.T) {
	ctx := context.Background()
	validPolicyID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	policyUUID, err := uuid.Parse(validPolicyID)
	require.NoError(t, err)

	testCases := []struct {
		name         string
		policyID     string
		mockSetup    func(db *database.MockDB)
		expected     int64
		wantErr      bool
		errorMessage string
	}{
		{
			name:     "Valid policy ID with completed swaps",
			policyID: validPolicyID,
			mockSetup: func(db *database.MockDB) {
				db.On("CountTransactions", ctx, policyUUID, types.StatusMined, "SWAP").Return(int64(5), nil)
			},
			expected: 5,
			wantErr:  false,
		},
		{
			name:     "Valid policy ID with no completed swaps",
			policyID: validPolicyID,
			mockSetup: func(db *database.MockDB) {
				db.On("CountTransactions", ctx, policyUUID, types.StatusMined, "SWAP").Return(int64(0), nil)
			},
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "Database error",
			policyID: validPolicyID,
			mockSetup: func(db *database.MockDB) {
				db.On("CountTransactions", ctx, policyUUID, types.StatusMined, "SWAP").Return(int64(0), fmt.Errorf("database error"))
			},
			expected:     0,
			wantErr:      true,
			errorMessage: `database error`,
		},
		{
			name:         "Invalid policy ID",
			policyID:     "invalid-uuid",
			mockSetup:    func(db *database.MockDB) {},
			expected:     0,
			wantErr:      true,
			errorMessage: `invalid policy_id`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUniswap := new(uniswapclient.MockUniswapClient)
			mockEth := new(ethclient.MockEthClient)
			mockDB := new(database.MockDB)
			logger := logrus.StandardLogger()

			tc.mockSetup(mockDB)
			plugin := &Plugin{
				uniswapClient: mockUniswap,
				rpcClient:     mockEth,
				db:            mockDB,
				logger:        logger,
			}

			count, err := plugin.getCompletedSwapTransactionsCount(ctx, tc.policyID)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, count)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestProposeTransactions(t *testing.T) {
	ctx := context.Background()
	validPolicyID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	policyUUID, err := uuid.Parse(validPolicyID)
	require.NoError(t, err)
	policy := createValidPolicy()
	policy.ID = validPolicyID

	testCases := []struct {
		name         string
		policy       types.PluginPolicy
		mockSetup    func(*database.MockDB, *uniswapclient.MockUniswapClient)
		expected     int
		wantErr      bool
		errorMessage string
	}{
		{
			name:   "All swaps completed",
			policy: policy,
			mockSetup: func(db *database.MockDB, mockUniswap *uniswapclient.MockUniswapClient) {
				dcaPolicy := Policy{}
				err := json.Unmarshal(policy.Policy, &dcaPolicy)
				require.NoError(t, err)

				totalOrders, _ := new(big.Int).SetString(dcaPolicy.TotalOrders, 10)

				db.On("CountTransactions", ctx, policyUUID, types.StatusMined, "SWAP").Return(totalOrders.Int64(), nil)
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)
				updatedPolicy := createSamplePolicy("DONE")
				db.On("UpdatePluginPolicyTx", ctx, nil, mock.AnythingOfType("types.PluginPolicy")).Return(&updatedPolicy, nil)
			},
			expected:     0,
			wantErr:      true,
			errorMessage: `policy completed all swaps`,
		},
		{
			name:   "Propose approve and swap",
			policy: policy,
			mockSetup: func(db *database.MockDB, uniswap *uniswapclient.MockUniswapClient) {
				db.On("CountTransactions", ctx, policyUUID, types.StatusMined, "SWAP").
					Return(int64(0), nil)

				routerAddr := gcommon.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
				uniswap.On("GetRouterAddress").Return(routerAddr)

				sourceAddr := gcommon.HexToAddress("0x1111111111111111111111111111111111111111")

				uniswap.On("GetAllowance", mock.Anything, sourceAddr).
					Return(big.NewInt(0), nil)

				uniswap.On("ApproveERC20Token", mock.Anything, mock.Anything, sourceAddr, routerAddr, mock.Anything, uint64(0)).
					Return([]byte("txhash1"), []byte("rawtx1"), nil)

				uniswap.On("GetExpectedAmountOut", mock.Anything, mock.Anything).
					Return(big.NewInt(90), nil)

				uniswap.On("CalculateAmountOutMin", big.NewInt(90)).
					Return(big.NewInt(89))

				uniswap.On("SwapTokens",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, uint64(1)).
					Return([]byte("txhash2"), []byte("rawtx2"), nil)
			},
			expected: 2,
			wantErr:  false,
		},
		{
			name:   "Propose only SWAP (sufficient allowance case)",
			policy: policy,
			mockSetup: func(db *database.MockDB, uniswap *uniswapclient.MockUniswapClient) {
				db.On("CountTransactions", ctx, policyUUID, types.StatusMined, "SWAP").
					Return(int64(0), nil)

				sourceAddr := gcommon.HexToAddress("0x1111111111111111111111111111111111111111")

				uniswap.On("GetAllowance", mock.Anything, sourceAddr).
					Return(big.NewInt(10000), nil)

				uniswap.On("GetExpectedAmountOut", mock.Anything, mock.Anything).
					Return(big.NewInt(90), nil)

				uniswap.On("CalculateAmountOutMin", big.NewInt(90)).
					Return(big.NewInt(89))

				uniswap.On("SwapTokens",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, uint64(0)).
					Return([]byte("txhash2"), []byte("rawtx2"), nil)
			},
			expected: 1,
			wantErr:  false,
		},
		{
			name:   "Error getting allowance",
			policy: policy,
			mockSetup: func(db *database.MockDB, uniswap *uniswapclient.MockUniswapClient) {
				db.On("CountTransactions", ctx, policyUUID, types.StatusMined, "SWAP").
					Return(int64(0), nil)

				sourceAddr := gcommon.HexToAddress("0x1111111111111111111111111111111111111111")

				uniswap.On("GetAllowance", mock.Anything, sourceAddr).
					Return(nil, errors.New("failed to get allowance"))
			},
			expected:     0,
			wantErr:      true,
			errorMessage: `failed to get allowance`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockUniswap := new(uniswapclient.MockUniswapClient)
			mockEth := new(ethclient.MockEthClient)
			mockDB := new(database.MockDB)
			logger := logrus.StandardLogger()

			tc.mockSetup(mockDB, mockUniswap)
			plugin := &Plugin{
				uniswapClient: mockUniswap,
				rpcClient:     mockEth,
				db:            mockDB,
				logger:        logger,
			}

			results, err := plugin.ProposeTransactions(tc.policy)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
				require.Len(t, results, tc.expected)
			}

			mockDB.AssertExpectations(t)
			mockUniswap.AssertExpectations(t)
		})
	}
}

func TestValidateInterval(t *testing.T) {
	testCases := []struct {
		name      string
		interval  string
		frequency string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Valid daily interval",
			interval:  "1",
			frequency: "daily",
			wantError: false,
		},
		{
			name:      "Valid weekly interval",
			interval:  "2",
			frequency: "weekly",
			wantError: false,
		},
		{
			name:      "Valid monthly interval",
			interval:  "3",
			frequency: "monthly",
			wantError: false,
		},
		{
			name:      "Minutely interval too small",
			interval:  "10",
			frequency: "minutely",
			wantError: true,
			errorMsg:  "minutely interval must be at least 15 minutes",
		},
		{
			name:      "Hourly interval too large",
			interval:  "24",
			frequency: "hourly",
			wantError: true,
			errorMsg:  "hourly interval must be at most 23 hours",
		},
		{
			name:      "Daily interval too large",
			interval:  "32",
			frequency: "daily",
			wantError: true,
			errorMsg:  "daily interval must be at most 31 days",
		},
		{
			name:      "Weekly interval too large",
			interval:  "53",
			frequency: "weekly",
			wantError: true,
			errorMsg:  "weekly interval must be at most 52 weeks",
		},
		{
			name:      "Monthly interval too large",
			interval:  "13",
			frequency: "monthly",
			wantError: true,
			errorMsg:  "monthly interval must be at most 12 months",
		},
		{
			name:      "Invalid frequency",
			interval:  "1",
			frequency: "invalid",
			wantError: true,
			errorMsg:  "invalid frequency",
		},
		{
			name:      "Invalid interval format",
			interval:  "not-a-number",
			frequency: "daily",
			wantError: true,
			errorMsg:  "invalid interval",
		},
		{
			name:      "Zero interval",
			interval:  "0",
			frequency: "daily",
			wantError: true,
			errorMsg:  "interval must be greater than 0",
		},
		{
			name:      "Negative interval",
			interval:  "-1",
			frequency: "daily",
			wantError: true,
			errorMsg:  "interval must be greater than 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateInterval(tc.interval, tc.frequency)

			if tc.wantError {
				require.Error(t, err)
				if tc.errorMsg != "" {
					require.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateProposedTransactions(t *testing.T) {
	ctx := context.Background()
	validPolicyID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	policyUUID, _ := uuid.Parse(validPolicyID)

	policy := createValidPolicy()
	policy.ID = validPolicyID

	chainID := big.NewInt(1)

	sourceAddr := gcommon.HexToAddress("0x1111111111111111111111111111111111111111")
	destAddr := gcommon.HexToAddress("0x2222222222222222222222222222222222222222")
	routerAddr := gcommon.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	signerAddr := gcommon.HexToAddress("0x7238f7c96DB71bf2bEda4909f023DAE40DEf3248")

	path := []gcommon.Address{sourceAddr, destAddr}
	swapAmount := big.NewInt(100)
	amountOutMin := big.NewInt(90)

	swapHash, swapRawTx, _ := createSwapTransaction(t, chainID, swapAmount, amountOutMin, path, signerAddr, routerAddr)
	approveHash, approveRawTx, _ := createApproveTransaction(t, chainID, routerAddr, swapAmount, sourceAddr)
	invalidTxHash, invalidRawTx, _ := createInvalidTransaction(t, chainID, "empty_data")

	validSwapRequest := createKeysignRequest(swapHash, swapRawTx, validPolicyID)
	validApproveRequest := createKeysignRequest(approveHash, approveRawTx, validPolicyID)
	invalidRequest := createKeysignRequest(invalidTxHash, invalidRawTx, validPolicyID)

	testCases := []struct {
		name         string
		policy       types.PluginPolicy
		txRequests   []types.PluginKeysignRequest
		mockSetup    func(*database.MockDB, *uniswapclient.MockUniswapClient)
		wantErr      bool
		errorMessage string
	}{
		{
			name:         "No transactions provided",
			policy:       policy,
			txRequests:   []types.PluginKeysignRequest{},
			mockSetup:    func(*database.MockDB, *uniswapclient.MockUniswapClient) {},
			wantErr:      true,
			errorMessage: "no transactions provided for validation",
		},
		{
			name:         "Policy validation fails",
			policy:       func() types.PluginPolicy { p := policy; p.PluginType = "invalid"; return p }(),
			txRequests:   []types.PluginKeysignRequest{validSwapRequest},
			mockSetup:    func(*database.MockDB, *uniswapclient.MockUniswapClient) {},
			wantErr:      true,
			errorMessage: "failed to validate plugin policy",
		},
		{
			name:       "All Swaps completed",
			policy:     policy,
			txRequests: []types.PluginKeysignRequest{validSwapRequest},
			mockSetup: func(db *database.MockDB, uniswap *uniswapclient.MockUniswapClient) {
				dcaPolicy := Policy{}
				err := json.Unmarshal(policy.Policy, &dcaPolicy)
				require.NoError(t, err)
				totalOrders, _ := new(big.Int).SetString(dcaPolicy.TotalOrders, 10)

				db.On("CountTransactions", mock.Anything, policyUUID, types.StatusMined, "SWAP").
					Return(totalOrders.Int64(), nil)

				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)
				updatedPolicy := createSamplePolicy("DONE")
				db.On("UpdatePluginPolicyTx", ctx, nil, mock.AnythingOfType("types.PluginPolicy")).Return(&updatedPolicy, nil)
			},
			wantErr:      true,
			errorMessage: "policy completed all swaps",
		},
		{
			name:       "Valid swap transaction",
			policy:     policy,
			txRequests: []types.PluginKeysignRequest{validSwapRequest},
			mockSetup: func(db *database.MockDB, uniswap *uniswapclient.MockUniswapClient) {
				// No completed swaps yet
				db.On("CountTransactions", mock.Anything, policyUUID, types.StatusMined, "SWAP").
					Return(int64(0), nil)

				// Router address for router calls
				routerAddr := gcommon.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
				uniswap.On("GetRouterAddress").Return(routerAddr).Times(1)
			},
			wantErr: false,
		},
		{
			name:       "Valid approve transaction",
			policy:     policy,
			txRequests: []types.PluginKeysignRequest{validApproveRequest},
			mockSetup: func(db *database.MockDB, uniswap *uniswapclient.MockUniswapClient) {
				// No completed swaps yet
				db.On("CountTransactions", mock.Anything, policyUUID, types.StatusMined, "SWAP").
					Return(int64(0), nil)

				routerAddr := gcommon.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
				uniswap.On("GetRouterAddress").Return(routerAddr).Times(2)
			},
			wantErr: false,
		},
		{
			name:       "Invalid transaction (empty data)",
			policy:     policy,
			txRequests: []types.PluginKeysignRequest{invalidRequest},
			mockSetup: func(db *database.MockDB, uniswap *uniswapclient.MockUniswapClient) {
				db.On("CountTransactions", mock.Anything, policyUUID, types.StatusMined, "SWAP").
					Return(int64(0), nil)

			},
			wantErr:      true,
			errorMessage: "transaction contains empty payload",
		},
		{
			name:       "Database error",
			policy:     policy,
			txRequests: []types.PluginKeysignRequest{validSwapRequest},
			mockSetup: func(db *database.MockDB, uniswap *uniswapclient.MockUniswapClient) {
				// Database error
				db.On("CountTransactions", mock.Anything, policyUUID, types.StatusMined, "SWAP").
					Return(int64(0), errors.New("database error"))
			},
			wantErr:      true,
			errorMessage: "fail to get completed swaps",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(database.MockDB)
			mockUniswap := new(uniswapclient.MockUniswapClient)
			mockEth := new(ethclient.MockEthClient)
			logger := logrus.StandardLogger()

			tc.mockSetup(mockDB, mockUniswap)

			plugin := &Plugin{
				uniswapClient: mockUniswap,
				logger:        logger,
				db:            mockDB,
				rpcClient:     mockEth,
			}

			err := plugin.ValidateProposedTransactions(tc.policy, tc.txRequests)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
			mockUniswap.AssertExpectations(t)
		})
	}
}

func TestValidateTransaction(t *testing.T) {
	chainID := big.NewInt(1)

	sourceAddr := gcommon.HexToAddress("0x1111111111111111111111111111111111111111")
	destAddr := gcommon.HexToAddress("0x2222222222222222222222222222222222222222")
	routerAddr := gcommon.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	signerAddr := gcommon.HexToAddress("0x7238f7c96DB71bf2bEda4909f023DAE40DEf3248")
	wrongAddr := gcommon.HexToAddress("0x4444444444444444444444444444444444444444")

	path := []gcommon.Address{sourceAddr, destAddr}
	swapAmount := big.NewInt(100)
	amountOutMin := big.NewInt(90)

	swapHash, swapRawBytesTx, _ := createSwapTransaction(t, chainID, swapAmount, amountOutMin, path, signerAddr, routerAddr)
	approveHash, approveRawBytesTx, _ := createApproveTransaction(t, chainID, routerAddr, swapAmount, sourceAddr)
	invalidTxHash, invalidTxRawBytes, _ := createInvalidTransaction(t, chainID, "empty_data")

	wrongChainIDTxHex, wrongChainIDTxRaw, _ := createInvalidTransaction(t, chainID, "wrong_chain_id")

	zeroGasLimitTxHex, zeroGasLimitTxRaw, _ := createInvalidTransaction(t, chainID, "zero_gas_limit")

	zeroGasPriceTxHex, zeroGasPriceTxRaw, _ := createInvalidTransaction(t, chainID, "zero_gas_price")

	testCases := []struct {
		name              string
		txHex             []byte
		txRlpBytes        []byte
		completedSwaps    int64
		policyTotalAmount *big.Int
		policyTotalOrders *big.Int
		policyChainID     *big.Int
		sourceAddr        *gcommon.Address
		destAddr          *gcommon.Address
		signerAddr        *gcommon.Address
		mockSetup         func(*uniswapclient.MockUniswapClient)
		wantError         bool
		errorMsg          string
	}{
		{
			name:              "Valid swap transaction",
			txHex:             swapHash,
			txRlpBytes:        swapRawBytesTx,
			completedSwaps:    0,
			policyTotalAmount: big.NewInt(1000),
			policyTotalOrders: big.NewInt(10),
			policyChainID:     chainID,
			sourceAddr:        &sourceAddr,
			destAddr:          &destAddr,
			signerAddr:        &signerAddr,
			mockSetup: func(uniswap *uniswapclient.MockUniswapClient) {
				uniswap.On("GetRouterAddress").Return(routerAddr)
			},
			wantError: false,
		},
		{
			name:              "Valid approve transaction",
			txHex:             approveHash,
			txRlpBytes:        approveRawBytesTx,
			completedSwaps:    0,
			policyTotalAmount: big.NewInt(1000),
			policyTotalOrders: big.NewInt(10),
			policyChainID:     chainID,
			sourceAddr:        &sourceAddr,
			destAddr:          &destAddr,
			signerAddr:        &signerAddr,
			mockSetup: func(uniswap *uniswapclient.MockUniswapClient) {
				uniswap.On("GetRouterAddress").Return(routerAddr)
			},
			wantError: false,
		},
		{
			name:              "Invalid transaction RLP",
			txHex:             swapHash,
			txRlpBytes:        []byte{1, 2, 34, 35, 36},
			completedSwaps:    0,
			policyTotalAmount: big.NewInt(1000),
			policyTotalOrders: big.NewInt(10),
			policyChainID:     chainID,
			sourceAddr:        &sourceAddr,
			destAddr:          &destAddr,
			signerAddr:        &signerAddr,
			mockSetup:         func(uniswap *uniswapclient.MockUniswapClient) {},
			wantError:         true,
			errorMsg:          "fail to parse RLP transaction",
		},
		{
			name:              "Wrong chain ID",
			txHex:             wrongChainIDTxHex,
			txRlpBytes:        wrongChainIDTxRaw,
			completedSwaps:    0,
			policyTotalAmount: big.NewInt(1000),
			policyTotalOrders: big.NewInt(10),
			policyChainID:     chainID,
			sourceAddr:        &sourceAddr,
			destAddr:          &destAddr,
			signerAddr:        &signerAddr,
			mockSetup:         func(uniswap *uniswapclient.MockUniswapClient) {},
			wantError:         true,
			errorMsg:          "chain ID mismatch",
		},
		{
			name:              "Zero gas limit",
			txHex:             zeroGasLimitTxHex,
			txRlpBytes:        zeroGasLimitTxRaw,
			completedSwaps:    0,
			policyTotalAmount: big.NewInt(1000),
			policyTotalOrders: big.NewInt(10),
			policyChainID:     chainID,
			sourceAddr:        &sourceAddr,
			destAddr:          &destAddr,
			signerAddr:        &signerAddr,
			mockSetup:         func(uniswap *uniswapclient.MockUniswapClient) {},
			wantError:         true,
			errorMsg:          "invalid gas limit",
		},
		{
			name:              "Zero gas price",
			txHex:             zeroGasPriceTxHex,
			txRlpBytes:        zeroGasPriceTxRaw,
			completedSwaps:    0,
			policyTotalAmount: big.NewInt(1000),
			policyTotalOrders: big.NewInt(10),
			policyChainID:     chainID,
			sourceAddr:        &sourceAddr,
			destAddr:          &destAddr,
			signerAddr:        &signerAddr,
			mockSetup:         func(uniswap *uniswapclient.MockUniswapClient) {},
			wantError:         true,
			errorMsg:          "invalid gas price",
		},
		{
			name:              "Empty transaction data",
			txHex:             invalidTxHash,
			txRlpBytes:        invalidTxRawBytes,
			completedSwaps:    0,
			policyTotalAmount: big.NewInt(1000),
			policyTotalOrders: big.NewInt(10),
			policyChainID:     chainID,
			sourceAddr:        &sourceAddr,
			destAddr:          &destAddr,
			signerAddr:        &signerAddr,
			mockSetup:         func(uniswap *uniswapclient.MockUniswapClient) {},
			wantError:         true,
			errorMsg:          "transaction contains empty payload",
		},
		{
			name:              "Unsupported destination",
			txHex:             swapHash,
			txRlpBytes:        swapRawBytesTx,
			completedSwaps:    0,
			policyTotalAmount: big.NewInt(1000),
			policyTotalOrders: big.NewInt(10),
			policyChainID:     chainID,
			sourceAddr:        &sourceAddr,
			destAddr:          &destAddr,
			signerAddr:        &signerAddr,
			mockSetup: func(uniswap *uniswapclient.MockUniswapClient) {
				uniswap.On("GetRouterAddress").Return(wrongAddr) // Different from tx destination
			},
			wantError: true,
			errorMsg:  "unsupported transaction",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUniswap := new(uniswapclient.MockUniswapClient)
			mockEth := new(ethclient.MockEthClient)
			mockDB := new(database.MockDB)
			logger := logrus.StandardLogger()

			tc.mockSetup(mockUniswap)

			plugin := Plugin{
				uniswapClient: mockUniswap,
				logger:        logger,
				db:            mockDB,
				rpcClient:     mockEth,
			}

			keysignRequest := createKeysignRequest(tc.txHex, tc.txRlpBytes, "policyID")

			err := plugin.validateTransaction(
				keysignRequest,
				tc.completedSwaps,
				tc.policyTotalAmount,
				tc.policyTotalOrders,
				tc.policyChainID,
				tc.sourceAddr,
				tc.destAddr,
				tc.signerAddr,
			)

			if tc.wantError {
				require.Error(t, err)
				if tc.errorMsg != "" {
					require.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}

			mockUniswap.AssertExpectations(t)
		})
	}
}

func TestSigningComplete(t *testing.T) {
	ctx := context.Background()

	// Create a test policy
	policy := createValidPolicy()

	// Create sample signature response
	signature := tss.KeysignResponse{
		Msg:          "message",
		R:            "R",
		S:            "S",
		DerSignature: "detsig",
		RecoveryID:   "0",
	}

	testCases := []struct {
		name         string
		signRequest  types.PluginKeysignRequest
		mockSetup    func(*ethclient.MockEthClient, *database.MockDB)
		waitMined    func(ctx context.Context, backend bind.DeployBackend, tx *gtypes.Transaction) (*gtypes.Receipt, error)
		signLegacyTx func(keysignResponse tss.KeysignResponse, rawTx string, chainID *big.Int) (*gtypes.Transaction, *gcommon.Address, error)
		wantError    bool
		errorMsg     string
	}{
		{
			name: "Successful transaction",
			mockSetup: func(ethClient *ethclient.MockEthClient, db *database.MockDB) {
				ethClient.On("SendTransaction", mock.Anything, mock.AnythingOfType("*types.Transaction")).
					Return(nil)
				db.On("CountTransactions", mock.Anything, mock.AnythingOfType("uuid.UUID"), types.StatusMined, "SWAP").
					Return(int64(0), nil)
			},
			waitMined: func(ctx context.Context, backend bind.DeployBackend, tx *gtypes.Transaction) (*gtypes.Receipt, error) {
				return &gtypes.Receipt{Status: 1}, nil
			},
			signLegacyTx: func(keysignResponse tss.KeysignResponse, rawTx string, chainID *big.Int) (*gtypes.Transaction, *gcommon.Address, error) {
				return &gtypes.Transaction{}, nil, nil
			},
			signRequest: func() types.PluginKeysignRequest {
				swapHash, swapRawTx, _ := createSwapTransaction(t,
					big.NewInt(1),
					big.NewInt(100),
					big.NewInt(90),
					[]gcommon.Address{gcommon.HexToAddress("0xtest"), gcommon.HexToAddress("0xtest2")},
					gcommon.HexToAddress("0xToaddress"),
					gcommon.HexToAddress("0xRouterAddress"),
				)

				signRequest := createKeysignRequest(swapHash, swapRawTx, policy.ID)
				return signRequest
			}(),
			wantError: false,
		},
		{
			name: "Successful last transaction",
			mockSetup: func(ethClient *ethclient.MockEthClient, db *database.MockDB) {
				ethClient.On("SendTransaction", mock.Anything, mock.AnythingOfType("*types.Transaction")).
					Return(nil)
				db.On("CountTransactions", mock.Anything, mock.AnythingOfType("uuid.UUID"), types.StatusMined, "SWAP").
					Return(int64(9), nil)
				db.On("WithTransaction", mock.Anything, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)
				db.On("UpdatePluginPolicyTx", mock.Anything, nil, mock.AnythingOfType("types.PluginPolicy")).
					Return(&types.PluginPolicy{}, nil)
			},
			waitMined: func(ctx context.Context, backend bind.DeployBackend, tx *gtypes.Transaction) (*gtypes.Receipt, error) {
				return &gtypes.Receipt{Status: 1}, nil
			},
			signLegacyTx: func(keysignResponse tss.KeysignResponse, rawTx string, chainID *big.Int) (*gtypes.Transaction, *gcommon.Address, error) {
				return &gtypes.Transaction{}, nil, nil
			},
			signRequest: func() types.PluginKeysignRequest {
				swapHash, swapRawTx, _ := createSwapTransaction(t,
					big.NewInt(1),
					big.NewInt(100),
					big.NewInt(90),
					[]gcommon.Address{gcommon.HexToAddress("0xtest"), gcommon.HexToAddress("0xtest2")},
					gcommon.HexToAddress("0xToaddress"),
					gcommon.HexToAddress("0xRouterAddress"),
				)

				signRequest := createKeysignRequest(swapHash, swapRawTx, policy.ID)
				return signRequest
			}(),
			wantError: false,
		},
		{
			name:      "Empty transaction hash",
			mockSetup: func(ethClient *ethclient.MockEthClient, db *database.MockDB) {},
			signRequest: func() types.PluginKeysignRequest {
				swapHash, swapRawTx, _ := createSwapTransaction(t,
					big.NewInt(1),
					big.NewInt(100),
					big.NewInt(90),
					[]gcommon.Address{gcommon.HexToAddress("0xtest"), gcommon.HexToAddress("0xtest2")},
					gcommon.HexToAddress("0xToaddress"),
					gcommon.HexToAddress("0xRouterAddress"),
				)

				signRequest := createKeysignRequest(swapHash, swapRawTx, policy.ID)
				signRequest.Messages[0] = ""
				return signRequest
			}(),
			wantError: true,
			errorMsg:  "transaction hash is missing",
		},
		{
			name:      "SignLegacyTx fails",
			mockSetup: func(ethClient *ethclient.MockEthClient, db *database.MockDB) {},
			signRequest: func() types.PluginKeysignRequest {
				swapHash, swapRawTx, _ := createSwapTransaction(t,
					big.NewInt(1),
					big.NewInt(100),
					big.NewInt(90),
					[]gcommon.Address{gcommon.HexToAddress("0xtest"), gcommon.HexToAddress("0xtest2")},
					gcommon.HexToAddress("0xToaddress"),
					gcommon.HexToAddress("0xRouterAddress"),
				)

				signRequest := createKeysignRequest(swapHash, swapRawTx, policy.ID)
				return signRequest
			}(),
			signLegacyTx: func(keysignResponse tss.KeysignResponse, rawTx string, chainID *big.Int) (*gtypes.Transaction, *gcommon.Address, error) {
				return &gtypes.Transaction{}, nil, errors.New("failed to sign")
			},
			wantError: true,
			errorMsg:  "fail to sign transaction",
		},
		{
			name: "SendTransaction fails",
			mockSetup: func(ethClient *ethclient.MockEthClient, db *database.MockDB) {
				ethClient.On("SendTransaction", mock.Anything, mock.AnythingOfType("*types.Transaction")).
					Return(errors.New("network error"))
			},
			signRequest: func() types.PluginKeysignRequest {
				swapHash, swapRawTx, _ := createSwapTransaction(t,
					big.NewInt(1),
					big.NewInt(100),
					big.NewInt(90),
					[]gcommon.Address{gcommon.HexToAddress("0xtest"), gcommon.HexToAddress("0xtest2")},
					gcommon.HexToAddress("0xToaddress"),
					gcommon.HexToAddress("0xRouterAddress"),
				)

				signRequest := createKeysignRequest(swapHash, swapRawTx, policy.ID)
				return signRequest
			}(),
			signLegacyTx: func(keysignResponse tss.KeysignResponse, rawTx string, chainID *big.Int) (*gtypes.Transaction, *gcommon.Address, error) {
				return &gtypes.Transaction{}, nil, nil
			},
			wantError: true,
			errorMsg:  "failed to send transaction",
		},
		{
			name: "WaitMined fails",
			mockSetup: func(ethClient *ethclient.MockEthClient, db *database.MockDB) {
				ethClient.On("SendTransaction", mock.Anything, mock.AnythingOfType("*types.Transaction")).
					Return(nil)
			},
			signRequest: func() types.PluginKeysignRequest {
				swapHash, swapRawTx, _ := createSwapTransaction(t,
					big.NewInt(1),
					big.NewInt(100),
					big.NewInt(90),
					[]gcommon.Address{gcommon.HexToAddress("0xtest"), gcommon.HexToAddress("0xtest2")},
					gcommon.HexToAddress("0xToaddress"),
					gcommon.HexToAddress("0xRouterAddress"),
				)

				signRequest := createKeysignRequest(swapHash, swapRawTx, policy.ID)
				return signRequest
			}(),
			signLegacyTx: func(keysignResponse tss.KeysignResponse, rawTx string, chainID *big.Int) (*gtypes.Transaction, *gcommon.Address, error) {
				return &gtypes.Transaction{}, nil, nil
			},
			waitMined: func(ctx context.Context, backend bind.DeployBackend, tx *gtypes.Transaction) (*gtypes.Receipt, error) {
				return nil, errors.New("waitMined fails")
			},
			wantError: true,
			errorMsg:  "fail to wait for transaction to be mined",
		},
		{
			name: "Transaction reverted",
			mockSetup: func(ethClient *ethclient.MockEthClient, db *database.MockDB) {
				ethClient.On("SendTransaction", mock.Anything, mock.AnythingOfType("*types.Transaction")).
					Return(nil)
			},
			signRequest: func() types.PluginKeysignRequest {
				swapHash, swapRawTx, _ := createSwapTransaction(t,
					big.NewInt(1),
					big.NewInt(100),
					big.NewInt(90),
					[]gcommon.Address{gcommon.HexToAddress("0xtest"), gcommon.HexToAddress("0xtest2")},
					gcommon.HexToAddress("0xToaddress"),
					gcommon.HexToAddress("0xRouterAddress"),
				)

				signRequest := createKeysignRequest(swapHash, swapRawTx, policy.ID)
				return signRequest
			}(),
			signLegacyTx: func(keysignResponse tss.KeysignResponse, rawTx string, chainID *big.Int) (*gtypes.Transaction, *gcommon.Address, error) {
				return &gtypes.Transaction{}, nil, nil
			},
			waitMined: func(ctx context.Context, backend bind.DeployBackend, tx *gtypes.Transaction) (*gtypes.Receipt, error) {
				return &gtypes.Receipt{Status: 0}, nil
			},
			wantError: true,
			errorMsg:  "transaction reverted",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockUniswap := new(uniswapclient.MockUniswapClient)
			mockEth := new(ethclient.MockEthClient)
			mockDB := new(database.MockDB)
			logger := logrus.StandardLogger()

			// Setup mocks
			tc.mockSetup(mockEth, mockDB)

			// Create plugin
			plugin := Plugin{
				waitMined:     tc.waitMined,
				signLegacyTx:  tc.signLegacyTx,
				uniswapClient: mockUniswap,
				logger:        logger,
				db:            mockDB,
				rpcClient:     mockEth,
			}

			// Call method
			err := plugin.SigningComplete(ctx, signature, tc.signRequest, policy)

			// Check expectations
			if tc.wantError {
				require.Error(t, err)
				if tc.errorMsg != "" {
					require.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}

			// Verify mocks
			mockEth.AssertExpectations(t)
			mockDB.AssertExpectations(t)
		})
	}
}
