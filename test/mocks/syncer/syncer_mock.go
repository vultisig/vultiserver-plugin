package syncer

import (
	"github.com/stretchr/testify/mock"

	"github.com/vultisig/vultiserver-plugin/internal/syncer"
	"github.com/vultisig/vultiserver-plugin/internal/types"
)

type MockSyncer struct {
	mock.Mock
}

func (m *MockSyncer) SyncTransaction(action syncer.Action, jwtToken string, tx types.TransactionHistory) error {
	args := m.Called(action, jwtToken, tx)
	return args.Error(0)
}

func (m *MockSyncer) CreatePolicySync(policy types.PluginPolicy) error {
	args := m.Called(policy)
	return args.Error(0)
}

func (m *MockSyncer) UpdatePolicySync(policy types.PluginPolicy) error {
	args := m.Called(policy)
	return args.Error(0)
}

func (m *MockSyncer) DeletePolicySync(policyID string, signature string) error {
	args := m.Called(policyID, signature)
	return args.Error(0)
}
