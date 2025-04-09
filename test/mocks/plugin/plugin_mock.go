package plugin

import (
	"context"
	"embed"
	"github.com/stretchr/testify/mock"
	"github.com/vultisig/mobile-tss-lib/tss"
	"github.com/vultisig/vultisigner/internal/types"
)

type MockPlugin struct {
	mock.Mock
}

func (m *MockPlugin) ProposeTransactions(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error) {
	args := m.Called(policy)
	return args.Get(0).([]types.PluginKeysignRequest), args.Error(1)
}

func (m *MockPlugin) SigningComplete(ctx context.Context, signature tss.KeysignResponse, request types.PluginKeysignRequest, policy types.PluginPolicy) error {
	args := m.Called(ctx, signature, request, policy)
	return args.Error(0)
}

func (m *MockPlugin) FrontendSchema() embed.FS {
	args := m.Called()
	return args.Get(0).(embed.FS)
}

func (m *MockPlugin) ValidatePluginPolicy(policyDoc types.PluginPolicy) error {
	args := m.Called(policyDoc)
	return args.Error(0)
}

func (m *MockPlugin) ValidateProposedTransactions(policy types.PluginPolicy, txs []types.PluginKeysignRequest) error {
	args := m.Called(policy, txs)
	return args.Error(0)
}
