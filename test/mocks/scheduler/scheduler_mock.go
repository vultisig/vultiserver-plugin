package scheduler

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/vultisig/vultiserver-plugin/internal/types"
)

type MockSchedulerService struct {
	mock.Mock
}

func (m *MockSchedulerService) CreateTimeTrigger(ctx context.Context, policy types.PluginPolicy, tx pgx.Tx) error {
	args := m.Called(ctx, policy, tx)
	return args.Error(0)
}

func (m *MockSchedulerService) GetTriggerFromPolicy(policy types.PluginPolicy) (*types.TimeTrigger, error) {
	args := m.Called(policy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TimeTrigger), args.Error(1)
}
