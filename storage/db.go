package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vultisig/vultisigner/internal/types"
)

type DatabaseStorage interface {
	Close() error

	GetPluginPolicy(ctx context.Context, id string) (types.PluginPolicy, error)
	GetAllPluginPolicies(ctx context.Context, publicKey string, pluginType string) ([]types.PluginPolicy, error)
	DeletePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, id string) error
	InsertPluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error)
	UpdatePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error)

	CreateTimeTriggerTx(ctx context.Context, dbTx pgx.Tx, trigger types.TimeTrigger) error
	GetPendingTimeTriggers(ctx context.Context) ([]types.TimeTrigger, error)
	UpdateTimeTriggerLastExecution(ctx context.Context, policyID string) error
	UpdateTimeTriggerTx(ctx context.Context, policyID string, trigger types.TimeTrigger, dbTx pgx.Tx) error

	DeleteTimeTrigger(ctx context.Context, policyID string) error
	UpdateTriggerStatus(ctx context.Context, policyID string, status types.TimeTriggerStatus) error
	GetTriggerStatus(ctx context.Context, policyID string) (types.TimeTriggerStatus, error)

	CreateTransactionHistoryTx(ctx context.Context, dbTx pgx.Tx, tx types.TransactionHistory) (uuid.UUID, error)
	UpdateTransactionStatusTx(ctx context.Context, dbTx pgx.Tx, txID uuid.UUID, status types.TransactionStatus, metadata map[string]interface{}) error
	CreateTransactionHistory(ctx context.Context, tx types.TransactionHistory) (uuid.UUID, error)
	UpdateTransactionStatus(ctx context.Context, txID uuid.UUID, status types.TransactionStatus, metadata map[string]interface{}) error
	GetTransactionHistory(ctx context.Context, policyID uuid.UUID, take int, skip int) ([]types.TransactionHistory, error)
	GetTransactionByHash(ctx context.Context, txHash string) (*types.TransactionHistory, error)

	Pool() *pgxpool.Pool
}
