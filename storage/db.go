package storage

import (
	"github.com/google/uuid"
	"github.com/vultisig/vultisigner/internal/types"
)

type DatabaseStorage interface {
	Close() error

	InsertPluginPolicy(policyDoc types.PluginPolicy) (types.PluginPolicy, error)
	GetPluginPolicy(id string) (types.PluginPolicy, error)
	GetAllPluginPolicies(publicKey string, pluginType string) ([]types.PluginPolicy, error)
	UpdatePluginPolicy(policyDoc types.PluginPolicy) (types.PluginPolicy, error)
	DeletePluginPolicy(id string) error

	CreateTimeTrigger(trigger types.TimeTrigger) error
	GetPendingTriggers() ([]types.TimeTrigger, error)
	UpdateTriggerExecution(policyID string) error

	CreateTransactionHistory(tx types.TransactionHistory) (uuid.UUID, error)
	UpdateTransactionStatus(txID uuid.UUID, status types.TransactionStatus, metadata map[string]interface{}) error
	GetTransactionHistory(policyID uuid.UUID) ([]types.TransactionHistory, error)
}
