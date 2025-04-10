package database

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/vultisig/vultisigner/internal/types"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) CountTransactions(ctx context.Context, policyUUID uuid.UUID, status types.TransactionStatus, txType string) (int64, error) {
	args := m.Called(ctx, policyUUID, status, txType)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDB) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	args := m.Called(ctx, fn)

	if val, ok := args.Get(0).(bool); ok && val {
		return fn(ctx, nil)
	}

	return args.Error(1)
}

func (m *MockDB) InsertPluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error) {
	args := m.Called(ctx, dbTx, policy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.PluginPolicy), args.Error(1)
}

func (m *MockDB) UpdatePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error) {
	args := m.Called(ctx, dbTx, policy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.PluginPolicy), args.Error(1)
}

func (m *MockDB) UpdateTimeTriggerTx(ctx context.Context, policyID string, trigger types.TimeTrigger, dbTx pgx.Tx) error {
	args := m.Called(ctx, policyID, trigger, dbTx)
	return args.Error(0)
}

func (m *MockDB) DeletePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, id string) error {
	args := m.Called(ctx, dbTx, id)
	return args.Error(0)
}

func (m *MockDB) GetAllPluginPolicies(ctx context.Context, pluginType string, publicKey string) ([]types.PluginPolicy, error) {
	args := m.Called(ctx, pluginType, publicKey)
	return args.Get(0).([]types.PluginPolicy), args.Error(1)
}

func (m *MockDB) GetTransactionHistory(ctx context.Context, policyID uuid.UUID, transactionType string, take int, skip int) ([]types.TransactionHistory, error) {
	args := m.Called(ctx, policyID, transactionType, take, skip)
	return args.Get(0).([]types.TransactionHistory), args.Error(1)
}

func (m *MockDB) GetPluginPolicy(ctx context.Context, policyID string) (types.PluginPolicy, error) {
	args := m.Called(ctx, policyID)
	return args.Get(0).(types.PluginPolicy), args.Error(1)
}

func (m *MockDB) UpdateTriggerStatus(ctx context.Context, policyID string, status types.TimeTriggerStatus) error {
	args := m.Called(ctx, policyID, status)
	return args.Error(0)
}

func (m *MockDB) UpdateTimeTriggerLastExecution(ctx context.Context, policyID string) error {
	args := m.Called(ctx, policyID)
	return args.Error(0)
}

func (m *MockDB) CreateTransactionHistoryTx(ctx context.Context, tx pgx.Tx, txHistory types.TransactionHistory) (uuid.UUID, error) {
	args := m.Called(ctx, tx, txHistory)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockDB) UpdateTransactionStatusTx(ctx context.Context, dbTx pgx.Tx, txID uuid.UUID, status types.TransactionStatus, metadata map[string]interface{}) error {
	args := m.Called(ctx, dbTx, txID, status, metadata)
	return args.Error(0)
}

func (m *MockDB) GetPendingTimeTriggers(ctx context.Context) ([]types.TimeTrigger, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.TimeTrigger), args.Error(1)
}

func (m *MockDB) GetTriggerStatus(ctx context.Context, policyID string) (types.TimeTriggerStatus, error) {
	args := m.Called(ctx, policyID)
	return args.Get(0).(types.TimeTriggerStatus), args.Error(1)
}

func (m *MockDB) DeleteTimeTrigger(ctx context.Context, policyID string) error {
	args := m.Called(ctx, policyID)
	return args.Error(0)
}

func (m *MockDB) CreateTimeTriggerTx(ctx context.Context, tx pgx.Tx, trigger types.TimeTrigger) error {
	args := m.Called(ctx, tx, trigger)
	return args.Error(0)
}
