package uniswapclient

import (
	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockUniswapClient struct {
	mock.Mock
}

func (m *MockUniswapClient) GetRouterAddress() *gcommon.Address {
	args := m.Called()
	addr := args.Get(0).(gcommon.Address)
	return &addr
}

func (m *MockUniswapClient) GetAllowance(owner gcommon.Address, token gcommon.Address) (*big.Int, error) {
	args := m.Called(owner, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockUniswapClient) GetTokenBalance(address *gcommon.Address, token gcommon.Address) (*big.Int, error) {
	args := m.Called(address, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockUniswapClient) GetExpectedAmountOut(amountIn *big.Int, path []gcommon.Address) (*big.Int, error) {
	args := m.Called(amountIn, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockUniswapClient) CalculateAmountOutMin(amountOut *big.Int) *big.Int {
	args := m.Called(amountOut)
	return args.Get(0).(*big.Int)
}

func (m *MockUniswapClient) ApproveERC20Token(chainID *big.Int, from *gcommon.Address, token gcommon.Address, spender gcommon.Address, amount *big.Int, nonce uint64) ([]byte, []byte, error) {
	args := m.Called(chainID, from, token, spender, amount, nonce)
	return args.Get(0).([]byte), args.Get(1).([]byte), args.Error(2)
}

func (m *MockUniswapClient) SwapTokens(chainID *big.Int, from *gcommon.Address, amountIn *big.Int, amountOutMin *big.Int, path []gcommon.Address, nonce uint64) ([]byte, []byte, error) {
	args := m.Called(chainID, from, amountIn, amountOutMin, path, nonce)
	return args.Get(0).([]byte), args.Get(1).([]byte), args.Error(2)
}
