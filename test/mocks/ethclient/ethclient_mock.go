package ethclient

import (
	"context"
	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockEthClient struct {
	mock.Mock
}

func (m *MockEthClient) SendTransaction(ctx context.Context, tx *gtypes.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}
func (m *MockEthClient) TransactionReceipt(ctx context.Context, txHash gcommon.Hash) (*gtypes.Receipt, error) {
	args := m.Called(ctx, txHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gtypes.Receipt), args.Error(1)
}
func (m *MockEthClient) CodeAt(ctx context.Context, account gcommon.Address, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, account, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}
