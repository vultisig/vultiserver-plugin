package service

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultiserver-plugin/internal/syncer"
	"github.com/vultisig/vultiserver-plugin/internal/types"
)

type Policy interface {
	CreatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) (*types.PluginPolicy, error)
	UpdatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) (*types.PluginPolicy, error)
	DeletePolicyWithSync(ctx context.Context, policyID, signature string) error
	GetPluginPolicies(ctx context.Context, pluginType, publicKey string, take int, skip int) (types.PluginPolicyPaginatedList, error)
	GetPluginPolicy(ctx context.Context, policyID string) (types.PluginPolicy, error)
	GetPluginPolicyTransactionHistory(ctx context.Context, policyID string, take int, skip int) (types.TransactionHistoryPaginatedList, error)
	CreatePricingPolicyWithSync(ctx context.Context, pluginPricingDto types.PluginPricingCreateDto) (*types.PluginPricing, error)
}

type PolicyServiceStorage interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
	InsertPluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error)
	UpdatePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error)
	UpdateTimeTriggerTx(ctx context.Context, policyID string, trigger types.TimeTrigger, dbTx pgx.Tx) error
	DeletePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, id string) error
	GetAllPluginPolicies(ctx context.Context, pluginType string, publicKey string, take int, skip int) (types.PluginPolicyPaginatedList, error)
	GetPluginPolicy(ctx context.Context, id string) (types.PluginPolicy, error)
	GetTransactionHistory(ctx context.Context, policyID uuid.UUID, transactionType string, take int, skip int) (types.TransactionHistoryPaginatedList, error)
	CreatePluginPricing(ctx context.Context, pluginPricingDto types.PluginPricingCreateDto) (*types.PluginPricing, error)
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

func (s *PolicyService) GetPluginPolicies(ctx context.Context, pluginType, publicKey string, take int, skip int) (types.PluginPolicyPaginatedList, error) {
	policies, err := s.db.GetAllPluginPolicies(ctx, pluginType, publicKey, take, skip)
	if err != nil {
		return types.PluginPolicyPaginatedList{}, fmt.Errorf("failed to get policies: %w", err)
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

func (s *PolicyService) GetPluginPolicyTransactionHistory(ctx context.Context, policyID string, take int, skip int) (types.TransactionHistoryPaginatedList, error) {
	policyUUID, err := uuid.Parse(policyID)
	if err != nil {
		return types.TransactionHistoryPaginatedList{}, fmt.Errorf("invalid policy_id: %s", policyID)
	}

	history, err := s.db.GetTransactionHistory(ctx, policyUUID, "SWAP", take, skip)
	if err != nil {
		return types.TransactionHistoryPaginatedList{}, fmt.Errorf("failed to get policy history: %w", err)
	}

	return history, nil
}

func (s *PolicyService) CreatePricingPolicyWithSync(
	ctx context.Context, pluginPricingDto types.PluginPricingCreateDto,
) (*types.PluginPricing, error) {
	pluginPricing, err := s.db.CreatePluginPricing(ctx, pluginPricingDto)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin pricing: %w", err)
	}

	// syncer should exist only on plugin instance
	if s.syncer != nil {
		err := s.syncer.CreatePricingPolicySync(pluginPricingDto)
		if err != nil {
			// Note: previous result is not rolled back here
			return nil, fmt.Errorf("failed to sync create pricing policy with verifier: %w", err)
		}
	}

	return pluginPricing, nil
}
