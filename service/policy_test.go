package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vultisig/vultisigner/internal/syncer"
	"github.com/vultisig/vultisigner/internal/types"
	"testing"
)

type MockPolicyStorage struct {
	mock.Mock
}

func (m *MockPolicyStorage) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	args := m.Called(ctx, fn)

	if val, ok := args.Get(0).(bool); ok && val {
		return fn(ctx, nil)
	}

	return args.Error(1)
}

func (m *MockPolicyStorage) InsertPluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error) {
	args := m.Called(ctx, dbTx, policy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.PluginPolicy), args.Error(1)
}

func (m *MockPolicyStorage) UpdatePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error) {
	args := m.Called(ctx, dbTx, policy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.PluginPolicy), args.Error(1)
}

func (m *MockPolicyStorage) UpdateTimeTriggerTx(ctx context.Context, policyID string, trigger types.TimeTrigger, dbTx pgx.Tx) error {
	args := m.Called(ctx, policyID, trigger, dbTx)
	return args.Error(0)
}

func (m *MockPolicyStorage) DeletePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, id string) error {
	args := m.Called(ctx, dbTx, id)
	return args.Error(0)
}

func (m *MockPolicyStorage) GetAllPluginPolicies(ctx context.Context, pluginType string, publicKey string) ([]types.PluginPolicy, error) {
	args := m.Called(ctx, pluginType, publicKey)
	return args.Get(0).([]types.PluginPolicy), args.Error(1)
}

func (m *MockPolicyStorage) GetPluginPolicy(ctx context.Context, id string) (types.PluginPolicy, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(types.PluginPolicy), args.Error(1)
}

func (m *MockPolicyStorage) GetTransactionHistory(ctx context.Context, policyID uuid.UUID, transactionType string, take int, skip int) ([]types.TransactionHistory, error) {
	args := m.Called(ctx, policyID, transactionType, take, skip)
	return args.Get(0).([]types.TransactionHistory), args.Error(1)
}

// MockPolicySyncer is a mock implementation of the PolicySyncer interface
type MockPolicySyncer struct {
	mock.Mock
}

func (m *MockPolicySyncer) CreatePolicySync(policy types.PluginPolicy) error {
	args := m.Called(policy)
	return args.Error(0)
}

func (m *MockPolicySyncer) UpdatePolicySync(policy types.PluginPolicy) error {
	args := m.Called(policy)
	return args.Error(0)
}

func (m *MockPolicySyncer) DeletePolicySync(policyID string, signature string) error {
	args := m.Called(policyID, signature)
	return args.Error(0)
}

func (m *MockPolicySyncer) SyncTransaction(action syncer.Action, jwtToken string, tx types.TransactionHistory) error {
	return nil
}

// MockSchedulerService is a mock implementation of the SchedulerService interface
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

func createSamplePolicy(id string) types.PluginPolicy {
	return types.PluginPolicy{
		ID:         id,
		PluginType: "testPlugin",
		PublicKey:  "testPublicKey",
		Policy:     []byte(`{"key":"value"}`),
	}
}

// Tests for the Policy Service

func TestNewPolicyService(t *testing.T) {
	testCases := []struct {
		name      string
		db        PolicyServiceStorage
		syncer    *MockPolicySyncer
		scheduler *MockSchedulerService
		logger    *logrus.Logger
		wantErr   bool
	}{
		{
			name:      "Valid initialization",
			db:        &MockPolicyStorage{},
			syncer:    &MockPolicySyncer{},
			scheduler: &MockSchedulerService{},
			logger:    logrus.StandardLogger(),
			wantErr:   false,
		},
		{
			name:      "Nil database",
			db:        nil,
			syncer:    &MockPolicySyncer{},
			scheduler: &MockSchedulerService{},
			logger:    logrus.StandardLogger(),
			wantErr:   true,
		},
		{
			name:      "Nil syncer is allowed",
			db:        &MockPolicyStorage{},
			syncer:    nil,
			scheduler: &MockSchedulerService{},
			logger:    logrus.StandardLogger(),
			wantErr:   false,
		},
		{
			name:      "Nil scheduler is allowed",
			db:        &MockPolicyStorage{},
			syncer:    &MockPolicySyncer{},
			scheduler: nil,
			logger:    logrus.StandardLogger(),
			wantErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, err := NewPolicyService(tc.db, tc.syncer, tc.scheduler, tc.logger)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, service)
			} else {
				require.NoError(t, err)
				require.NotNil(t, service)
			}
		})
	}
}

func TestCreatePolicyWithSync(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name         string
		policy       types.PluginPolicy
		mockSetup    func(*MockPolicyStorage, *MockPolicySyncer, *MockSchedulerService)
		wantErr      bool
		errorMessage string
	}{
		{
			name:   "Successful policy creation",
			policy: createSamplePolicy("policy1"),
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer, scheduler *MockSchedulerService) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				insertedPolicy := createSamplePolicy("policy1")
				db.On("InsertPluginPolicyTx", ctx, nil, mock.AnythingOfType("types.PluginPolicy")).Return(&insertedPolicy, nil)

				scheduler.On("CreateTimeTrigger", ctx, mock.AnythingOfType("types.PluginPolicy"), nil).Return(nil)

				syncer.On("CreatePolicySync", mock.MatchedBy(func(p types.PluginPolicy) bool {
					return p.ID == "policy1"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "Database error",
			policy: createSamplePolicy("policy2"),
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer, scheduler *MockSchedulerService) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				db.On("InsertPluginPolicyTx", ctx, nil, mock.AnythingOfType("types.PluginPolicy")).Return(nil, fmt.Errorf("database error"))
			},
			wantErr:      true,
			errorMessage: "failed to insert policy",
		},
		{
			name:   "Scheduler error",
			policy: createSamplePolicy("policy3"),
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer, scheduler *MockSchedulerService) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				insertedPolicy := createSamplePolicy("policy3")
				db.On("InsertPluginPolicyTx", ctx, nil, mock.AnythingOfType("types.PluginPolicy")).Return(&insertedPolicy, nil)

				scheduler.On("CreateTimeTrigger", ctx, mock.AnythingOfType("types.PluginPolicy"), nil).Return(fmt.Errorf("scheduler error"))
			},
			wantErr:      true,
			errorMessage: "failed to create time trigger",
		},
		{
			name:   "Syncer error",
			policy: createSamplePolicy("policy4"),
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer, scheduler *MockSchedulerService) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				insertedPolicy := createSamplePolicy("policy3")
				db.On("InsertPluginPolicyTx", ctx, nil, mock.AnythingOfType("types.PluginPolicy")).Return(&insertedPolicy, nil)

				scheduler.On("CreateTimeTrigger", ctx, mock.AnythingOfType("types.PluginPolicy"), nil).Return(nil)

				syncer.On("CreatePolicySync", mock.AnythingOfType("types.PluginPolicy")).Return(fmt.Errorf("syncer error"))
			},
			wantErr:      true,
			errorMessage: "failed to sync create policy",
		},
		{
			name:   "transaction error",
			policy: createSamplePolicy("policy5"),
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer, scheduler *MockSchedulerService) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(nil, fmt.Errorf("transaction error"))
			},
			wantErr:      true,
			errorMessage: "transaction error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockDB := new(MockPolicyStorage)
			mockSyncer := new(MockPolicySyncer)
			mockScheduler := new(MockSchedulerService)

			tc.mockSetup(mockDB, mockSyncer, mockScheduler)

			policyService, err := NewPolicyService(mockDB, mockSyncer, mockScheduler, logrus.StandardLogger())
			require.NoError(t, err)

			result, err := policyService.CreatePolicyWithSync(ctx, tc.policy)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.policy.ID, result.ID)
			}

			mockDB.AssertExpectations(t)
			mockSyncer.AssertExpectations(t)
			mockScheduler.AssertExpectations(t)
		})
	}
}

func TestUpdatePolicyWithSync(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name         string
		policy       types.PluginPolicy
		mockSetup    func(db *MockPolicyStorage, syncer *MockPolicySyncer, scheduler *MockSchedulerService)
		wantErr      bool
		errorMessage string
	}{
		{
			name:   "successful policy update",
			policy: createSamplePolicy("policy1"),
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer, scheduler *MockSchedulerService) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				updatedPolicy := createSamplePolicy("policy1")
				db.On("UpdatePluginPolicyTx", ctx, nil, mock.AnythingOfType("types.PluginPolicy")).Return(&updatedPolicy, nil)

				trigger := &types.TimeTrigger{PolicyID: "policy1"}
				scheduler.On("GetTriggerFromPolicy", mock.AnythingOfType("types.PluginPolicy")).Return(trigger, nil)

				db.On("UpdateTimeTriggerTx", ctx, "policy1", *trigger, nil).Return(nil)

				syncer.On("UpdatePolicySync", mock.AnythingOfType("types.PluginPolicy")).Return(nil)
			},
			wantErr: false,
		},

		{
			name:   "database update policy error",
			policy: createSamplePolicy("policy2"),
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer, scheduler *MockSchedulerService) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				db.On("UpdatePluginPolicyTx", ctx, nil, mock.AnythingOfType("types.PluginPolicy")).Return(nil, fmt.Errorf("database update error"))
			},
			wantErr:      true,
			errorMessage: "failed to update policy: database update error",
		},
		{
			name:   "get trigger error",
			policy: createSamplePolicy("policy3"),
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer, scheduler *MockSchedulerService) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				updatedPolicy := createSamplePolicy("policy3")
				db.On("UpdatePluginPolicyTx", ctx, nil, mock.AnythingOfType("types.PluginPolicy")).Return(&updatedPolicy, nil)

				scheduler.On("GetTriggerFromPolicy", mock.AnythingOfType("types.PluginPolicy")).Return(nil, fmt.Errorf("trigger error"))
			},
			wantErr:      true,
			errorMessage: "failed to get trigger from policy",
		},
		{
			name:   "database update time trigger error",
			policy: createSamplePolicy("policy4"),
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer, scheduler *MockSchedulerService) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				updatedPolicy := createSamplePolicy("policy4")
				db.On("UpdatePluginPolicyTx", ctx, nil, mock.AnythingOfType("types.PluginPolicy")).Return(&updatedPolicy, nil)

				trigger := &types.TimeTrigger{PolicyID: "policy4"}
				scheduler.On("GetTriggerFromPolicy", mock.AnythingOfType("types.PluginPolicy")).Return(trigger, nil)

				db.On("UpdateTimeTriggerTx", ctx, "policy4", *trigger, nil).Return(fmt.Errorf("database time trigger update error"))
			},
			wantErr:      true,
			errorMessage: "failed to update time trigger",
		},
		{
			name:   "syncer error",
			policy: createSamplePolicy("policy5"),
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer, scheduler *MockSchedulerService) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				updatedPolicy := createSamplePolicy("policy5")
				db.On("UpdatePluginPolicyTx", ctx, nil, mock.AnythingOfType("types.PluginPolicy")).Return(&updatedPolicy, nil)

				trigger := &types.TimeTrigger{PolicyID: "policy5"}
				scheduler.On("GetTriggerFromPolicy", mock.AnythingOfType("types.PluginPolicy")).Return(trigger, nil)

				db.On("UpdateTimeTriggerTx", ctx, "policy5", *trigger, nil).Return(nil)

				syncer.On("UpdatePolicySync", mock.AnythingOfType("types.PluginPolicy")).Return(fmt.Errorf("syncer update error"))
			},
			wantErr:      true,
			errorMessage: "failed to sync update policy",
		},
		{
			name:   "transaction error",
			policy: createSamplePolicy("policy6"),
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer, scheduler *MockSchedulerService) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(nil, fmt.Errorf("transaction error"))
			},
			wantErr:      true,
			errorMessage: "transaction error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockDB := new(MockPolicyStorage)
			mockSyncer := new(MockPolicySyncer)
			mockScheduler := new(MockSchedulerService)

			tc.mockSetup(mockDB, mockSyncer, mockScheduler)

			policyService, err := NewPolicyService(mockDB, mockSyncer, mockScheduler, logrus.StandardLogger())
			require.NoError(t, err)

			result, err := policyService.UpdatePolicyWithSync(ctx, tc.policy)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.policy.ID, result.ID)
			}

			mockDB.AssertExpectations(t)
			mockSyncer.AssertExpectations(t)
			mockScheduler.AssertExpectations(t)
		})
	}
}

func TestDeletePolicyWithSync(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name         string
		policyID     string
		signature    string
		mockSetup    func(*MockPolicyStorage, *MockPolicySyncer)
		expectErr    bool
		errorMessage string
	}{
		{
			name:      "successful policy deletion",
			policyID:  "policy1",
			signature: "validSignature",
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				db.On("DeletePluginPolicyTx", ctx, nil, "policy1").
					Return(nil)

				syncer.On("DeletePolicySync", "policy1", "validSignature").
					Return(nil)
			},
			expectErr: false,
		},
		{
			name:      "database delete error",
			policyID:  "policy2",
			signature: "validSignature",
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				db.On("DeletePluginPolicyTx", ctx, nil, "policy2").
					Return(fmt.Errorf("delete error"))
			},
			expectErr:    true,
			errorMessage: "failed to delete policy",
		},
		{
			name:      "syncer error",
			policyID:  "policy3",
			signature: "validSignature",
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				db.On("DeletePluginPolicyTx", ctx, nil, "policy3").
					Return(nil)

				syncer.On("DeletePolicySync", "policy3", "validSignature").
					Return(fmt.Errorf("syncer error"))
			},
			expectErr:    true,
			errorMessage: "failed to sync delete policy",
		},
		{
			name:      "transaction error",
			policyID:  "policy4",
			signature: "validSignature",
			mockSetup: func(db *MockPolicyStorage, syncer *MockPolicySyncer) {
				// Setup transaction mock with error
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(nil, fmt.Errorf("transaction error"))
			},
			expectErr:    true,
			errorMessage: "transaction error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(MockPolicyStorage)
			mockSyncer := new(MockPolicySyncer)

			tc.mockSetup(mockDB, mockSyncer)

			policyService, err := NewPolicyService(mockDB, mockSyncer, nil, logrus.StandardLogger())
			require.NoError(t, err)

			err = policyService.DeletePolicyWithSync(ctx, tc.policyID, tc.signature)

			if tc.expectErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
			mockSyncer.AssertExpectations(t)
		})
	}
}

func TestGetPluginPolicies(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name         string
		pluginType   string
		publicKey    string
		mockSetup    func(*MockPolicyStorage)
		expectErr    bool
		errorMessage string
		expectedLen  int
	}{
		{
			name:       "successful get policies",
			pluginType: "testType",
			publicKey:  "testKey",
			mockSetup: func(db *MockPolicyStorage) {
				policies := []types.PluginPolicy{
					createSamplePolicy("policy1"),
					createSamplePolicy("policy2"),
				}
				db.On("GetAllPluginPolicies", ctx, "testType", "testKey").
					Return(policies, nil)
			},
			expectErr:   false,
			expectedLen: 2,
		},
		{
			name:       "no policies found",
			pluginType: "emptyType",
			publicKey:  "emptyKey",
			mockSetup: func(db *MockPolicyStorage) {
				db.On("GetAllPluginPolicies", ctx, "emptyType", "emptyKey").
					Return([]types.PluginPolicy{}, nil)
			},
			expectErr:   false,
			expectedLen: 0,
		},
		{
			name:       "database error",
			pluginType: "errorType",
			publicKey:  "errorKey",
			mockSetup: func(db *MockPolicyStorage) {
				db.On("GetAllPluginPolicies", ctx, "errorType", "errorKey").
					Return([]types.PluginPolicy{}, fmt.Errorf("database error"))
			},
			expectErr:    true,
			errorMessage: "failed to get policies",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(MockPolicyStorage)

			tc.mockSetup(mockDB)

			policyService, err := NewPolicyService(mockDB, nil, nil, logrus.StandardLogger())
			require.NoError(t, err)

			policies, err := policyService.GetPluginPolicies(ctx, tc.pluginType, tc.publicKey)

			if tc.expectErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
				require.Len(t, policies, tc.expectedLen)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetPluginPolicy(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name         string
		policyID     string
		mockSetup    func(*MockPolicyStorage)
		expectErr    bool
		errorMessage string
	}{
		{
			name:     "successful get policy",
			policyID: "policy1",
			mockSetup: func(db *MockPolicyStorage) {
				policy := createSamplePolicy("policy1")
				db.On("GetPluginPolicy", ctx, "policy1").
					Return(policy, nil)
			},
			expectErr: false,
		},
		{
			name:     "policy not found",
			policyID: "nonexistent",
			mockSetup: func(db *MockPolicyStorage) {
				db.On("GetPluginPolicy", ctx, "nonexistent").
					Return(types.PluginPolicy{}, fmt.Errorf("not found"))
			},
			expectErr:    true,
			errorMessage: "failed to get policy",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(MockPolicyStorage)

			tc.mockSetup(mockDB)

			policyService, err := NewPolicyService(mockDB, nil, nil, logrus.StandardLogger())
			require.NoError(t, err)

			policy, err := policyService.GetPluginPolicy(ctx, tc.policyID)

			if tc.expectErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
				require.Empty(t, policy.ID)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.policyID, policy.ID)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetPluginPolicyTransactionHistory(t *testing.T) {
	ctx := context.Background()
	policyID := "23cc1630-227e-4887-9d13-238eb8c6fa02"
	policyUUID, err := uuid.Parse(policyID)
	require.NoError(t, err)

	testCases := []struct {
		name         string
		policyID     string
		mockSetup    func(*MockPolicyStorage)
		expectErr    bool
		errorMessage string
		expectedLen  int
	}{
		{
			name:     "successful get transaction history",
			policyID: policyID,
			mockSetup: func(db *MockPolicyStorage) {
				transactionHistory := []types.TransactionHistory{
					{
						TxHash: "txHash",
						TxBody: "txBody",
					},
					{
						TxHash: "txHash2",
						TxBody: "txBody2",
					},
				}

				db.On("GetTransactionHistory", ctx, policyUUID, "SWAP", 30, 0).
					Return(transactionHistory, nil)
			},
			expectErr:   false,
			expectedLen: 2,
		},
		{
			name:         "invalid policy UUID",
			policyID:     "nonexistent",
			mockSetup:    func(db *MockPolicyStorage) {},
			expectErr:    true,
			errorMessage: "invalid policy_id",
			expectedLen:  0,
		},
		{
			name:     "database error",
			policyID: policyID,
			mockSetup: func(db *MockPolicyStorage) {
				db.On("GetTransactionHistory", ctx, policyUUID, "SWAP", 30, 0).
					Return([]types.TransactionHistory{}, fmt.Errorf("database error"))
			},
			expectErr:    true,
			errorMessage: "failed to get policy history",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(MockPolicyStorage)

			tc.mockSetup(mockDB)

			policyService, err := NewPolicyService(mockDB, nil, nil, logrus.StandardLogger())
			require.NoError(t, err)

			policies, err := policyService.GetPluginPolicyTransactionHistory(ctx, tc.policyID)

			if tc.expectErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
				require.Len(t, policies, tc.expectedLen)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
