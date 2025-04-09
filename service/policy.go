package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/internal/syncer"
	"github.com/vultisig/vultisigner/internal/types"
	"reflect"
)

type Policy interface {
	CreatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) (*types.PluginPolicy, error)
	UpdatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) (*types.PluginPolicy, error)
	DeletePolicyWithSync(ctx context.Context, policyID, signature string) error
	GetPluginPolicies(ctx context.Context, pluginType, publicKey string) ([]types.PluginPolicy, error)
	GetPluginPolicy(ctx context.Context, policyID string) (types.PluginPolicy, error)
	GetPluginPolicyTransactionHistory(ctx context.Context, policyID string) ([]types.TransactionHistory, error)
}

type PolicyServiceStorage interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
	InsertPluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error)
	UpdatePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error)
	UpdateTimeTriggerTx(ctx context.Context, policyID string, trigger types.TimeTrigger, dbTx pgx.Tx) error
	DeletePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, id string) error
	GetAllPluginPolicies(ctx context.Context, publicKey string, pluginType string) ([]types.PluginPolicy, error)
	GetPluginPolicy(ctx context.Context, id string) (types.PluginPolicy, error)
	GetTransactionHistory(ctx context.Context, policyID uuid.UUID, transactionType string, take int, skip int) ([]types.TransactionHistory, error)
}

type SchedulerService interface {
	CreateTimeTrigger(ctx context.Context, policy types.PluginPolicy, tx pgx.Tx) error
	GetTriggerFromPolicy(policy types.PluginPolicy) (*types.TimeTrigger, error)
}

type PolicyService struct {
	db        PolicyServiceStorage
	syncer    syncer.PolicySyncer
	scheduler SchedulerService
	logger    *logrus.Logger
}

func NewPolicyService(db PolicyServiceStorage, syncer syncer.PolicySyncer, scheduler SchedulerService, logger *logrus.Logger) (*PolicyService, error) {
	if db == nil {
		return nil, fmt.Errorf("database storage cannot be nil")
	}
	return &PolicyService{
		db:        db,
		syncer:    syncer,
		scheduler: scheduler,
		logger:    logger,
	}, nil
}

func (s *PolicyService) CreatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) (*types.PluginPolicy, error) {
	var newPolicy *types.PluginPolicy

	err := s.db.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error

		// Insert the policy in the database
		newPolicy, err = s.db.InsertPluginPolicyTx(ctx, tx, policy)
		if err != nil {
			return fmt.Errorf("failed to insert policy: %w", err)
		}

		// Create time trigger if scheduler is available
		if s.scheduler != nil && !reflect.ValueOf(s.scheduler).IsNil() {
			if err := s.scheduler.CreateTimeTrigger(ctx, policy, tx); err != nil {
				return fmt.Errorf("failed to create time trigger: %w", err)
			}
		}

		// Sync the policy if syncer is available
		if s.syncer != nil && !reflect.ValueOf(s.syncer).IsNil() {
			err := s.syncer.CreatePolicySync(policy)
			if err != nil {
				return fmt.Errorf("failed to sync create policy: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newPolicy, nil
}

func (s *PolicyService) UpdatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) (*types.PluginPolicy, error) {
	var updatedPolicy *types.PluginPolicy

	err := s.db.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error

		updatedPolicy, err = s.db.UpdatePluginPolicyTx(ctx, tx, policy)
		if err != nil {
			return fmt.Errorf("failed to update policy: %w", err)
		}

		if s.scheduler != nil && !reflect.ValueOf(s.scheduler).IsNil() {
			trigger, err := s.scheduler.GetTriggerFromPolicy(policy)
			if err != nil {
				return fmt.Errorf("failed to get trigger from policy: %w", err)
			}

			if err := s.db.UpdateTimeTriggerTx(ctx, policy.ID, *trigger, tx); err != nil {
				return fmt.Errorf("failed to update time trigger: %w", err)
			}
		}

		if s.syncer != nil && !reflect.ValueOf(s.syncer).IsNil() {
			if err := s.syncer.UpdatePolicySync(policy); err != nil {
				return fmt.Errorf("failed to sync update policy: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updatedPolicy, nil
}

func (s *PolicyService) DeletePolicyWithSync(ctx context.Context, policyID, signature string) error {
	return s.db.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {

		if err := s.db.DeletePluginPolicyTx(ctx, tx, policyID); err != nil {
			return fmt.Errorf("failed to delete policy: %w", err)
		}
		if s.syncer != nil && !reflect.ValueOf(s.syncer).IsNil() {
			if err := s.syncer.DeletePolicySync(policyID, signature); err != nil {
				return fmt.Errorf("failed to sync delete policy: %w", err)
			}
		}
		return nil
	})
}

func (s *PolicyService) GetPluginPolicies(ctx context.Context, pluginType, publicKey string) ([]types.PluginPolicy, error) {
	policies, err := s.db.GetAllPluginPolicies(ctx, pluginType, publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get policies: %w", err)
	}
	return policies, nil
}

func (s *PolicyService) GetPluginPolicy(ctx context.Context, policyID string) (types.PluginPolicy, error) {
	policy, err := s.db.GetPluginPolicy(ctx, policyID)
	if err != nil {
		return types.PluginPolicy{}, fmt.Errorf("failed to get policy: %w", err)
	}
	return policy, nil
}

func (s *PolicyService) GetPluginPolicyTransactionHistory(ctx context.Context, policyID string) ([]types.TransactionHistory, error) {
	// Convert string to UUID
	policyUUID, err := uuid.Parse(policyID)
	if err != nil {
		return []types.TransactionHistory{}, fmt.Errorf("invalid policy_id: %s", policyID)
	}

	history, err := s.db.GetTransactionHistory(ctx, policyUUID, "SWAP", 30, 0) // take the last 30 records and skip the first 0
	if err != nil {
		return []types.TransactionHistory{}, fmt.Errorf("failed to get policy history: %w", err)
	}

	return history, nil
}
