package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vultisig/vultiserver-plugin/internal/types"
	"github.com/vultisig/vultiserver-plugin/test/mocks/database"
	"github.com/vultisig/vultiserver-plugin/test/mocks/scheduler"
	"github.com/vultisig/vultiserver-plugin/test/mocks/syncer"

	"testing"
)

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
		syncer    *syncer.MockSyncer
		scheduler *scheduler.MockSchedulerService
		logger    *logrus.Logger
		wantErr   bool
	}{
		{
			name:      "Valid initialization",
			db:        &database.MockDB{},
			syncer:    &syncer.MockSyncer{},
			scheduler: &scheduler.MockSchedulerService{},
			logger:    logrus.StandardLogger(),
			wantErr:   false,
		},
		{
			name:      "Nil database",
			db:        nil,
			syncer:    &syncer.MockSyncer{},
			scheduler: &scheduler.MockSchedulerService{},
			logger:    logrus.StandardLogger(),
			wantErr:   true,
		},
		{
			name:      "Nil syncer is allowed",
			db:        &database.MockDB{},
			syncer:    nil,
			scheduler: &scheduler.MockSchedulerService{},
			logger:    logrus.StandardLogger(),
			wantErr:   false,
		},
		{
			name:      "Nil scheduler is allowed",
			db:        &database.MockDB{},
			syncer:    &syncer.MockSyncer{},
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
		mockSetup    func(*database.MockDB, *syncer.MockSyncer, *scheduler.MockSchedulerService)
		wantErr      bool
		errorMessage string
	}{
		{
			name:   "Successful policy creation",
			policy: createSamplePolicy("policy1"),
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer, scheduler *scheduler.MockSchedulerService) {
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
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer, scheduler *scheduler.MockSchedulerService) {
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
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer, scheduler *scheduler.MockSchedulerService) {
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
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer, scheduler *scheduler.MockSchedulerService) {
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
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer, scheduler *scheduler.MockSchedulerService) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(nil, fmt.Errorf("transaction error"))
			},
			wantErr:      true,
			errorMessage: "transaction error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockDB := new(database.MockDB)
			mockSyncer := new(syncer.MockSyncer)
			mockScheduler := new(scheduler.MockSchedulerService)

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
		mockSetup    func(db *database.MockDB, syncer *syncer.MockSyncer, scheduler *scheduler.MockSchedulerService)
		wantErr      bool
		errorMessage string
	}{
		{
			name:   "successful policy update",
			policy: createSamplePolicy("policy1"),
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer, scheduler *scheduler.MockSchedulerService) {
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
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer, scheduler *scheduler.MockSchedulerService) {
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
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer, scheduler *scheduler.MockSchedulerService) {
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
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer, scheduler *scheduler.MockSchedulerService) {
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
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer, scheduler *scheduler.MockSchedulerService) {
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
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer, scheduler *scheduler.MockSchedulerService) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(nil, fmt.Errorf("transaction error"))
			},
			wantErr:      true,
			errorMessage: "transaction error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockDB := new(database.MockDB)
			mockSyncer := new(syncer.MockSyncer)
			mockScheduler := new(scheduler.MockSchedulerService)

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
		mockSetup    func(*database.MockDB, *syncer.MockSyncer)
		expectErr    bool
		errorMessage string
	}{
		{
			name:      "successful policy deletion",
			policyID:  "policy1",
			signature: "validSignature",
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer) {
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
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer) {
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
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer) {
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
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer) {
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
			mockDB := new(database.MockDB)
			mockSyncer := new(syncer.MockSyncer)

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
	take := 10
	skip := 0

	testCases := []struct {
		name         string
		pluginType   string
		publicKey    string
		mockSetup    func(*database.MockDB)
		expectErr    bool
		errorMessage string
		expectedLen  int
	}{
		{
			name:       "successful get policies",
			pluginType: "testType",
			publicKey:  "testKey",
			mockSetup: func(db *database.MockDB) {
				policies := types.PluginPolicyPaginatedList{
					Policies: []types.PluginPolicy{
						createSamplePolicy("policy1"),
						createSamplePolicy("policy2"),
					},
					TotalCount: 2,
				}
				db.On("GetAllPluginPolicies", ctx, "testType", "testKey", take, skip).
					Return(policies, nil)
			},
			expectErr:   false,
			expectedLen: 2,
		},
		{
			name:       "no policies found",
			pluginType: "emptyType",
			publicKey:  "emptyKey",
			mockSetup: func(db *database.MockDB) {
				db.On("GetAllPluginPolicies", ctx, "emptyType", "emptyKey", take, skip).
					Return(types.PluginPolicyPaginatedList{
						Policies:   []types.PluginPolicy{},
						TotalCount: 0,
					}, nil)
			},
			expectErr:   false,
			expectedLen: 0,
		},
		{
			name:       "database error",
			pluginType: "errorType",
			publicKey:  "errorKey",
			mockSetup: func(db *database.MockDB) {
				db.On("GetAllPluginPolicies", ctx, "errorType", "errorKey", take, skip).
					Return(types.PluginPolicyPaginatedList{}, fmt.Errorf("database error"))
			},
			expectErr:    true,
			errorMessage: "failed to get policies",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(database.MockDB)

			tc.mockSetup(mockDB)

			policyService, err := NewPolicyService(mockDB, nil, nil, logrus.StandardLogger())
			require.NoError(t, err)

			policies, err := policyService.GetPluginPolicies(ctx, tc.pluginType, tc.publicKey, 10, 0)

			if tc.expectErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
				require.Len(t, policies.Policies, tc.expectedLen)
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
		mockSetup    func(*database.MockDB)
		expectErr    bool
		errorMessage string
	}{
		{
			name:     "successful get policy",
			policyID: "policy1",
			mockSetup: func(db *database.MockDB) {
				policy := createSamplePolicy("policy1")
				db.On("GetPluginPolicy", ctx, "policy1").
					Return(policy, nil)
			},
			expectErr: false,
		},
		{
			name:     "policy not found",
			policyID: "nonexistent",
			mockSetup: func(db *database.MockDB) {
				db.On("GetPluginPolicy", ctx, "nonexistent").
					Return(types.PluginPolicy{}, fmt.Errorf("not found"))
			},
			expectErr:    true,
			errorMessage: "failed to get policy",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(database.MockDB)

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
	take := 30
	skip := 0

	testCases := []struct {
		name         string
		policyID     string
		mockSetup    func(*database.MockDB)
		expectErr    bool
		errorMessage string
		expectedLen  int
	}{
		{
			name:     "successful get transaction history",
			policyID: policyID,
			mockSetup: func(db *database.MockDB) {
				transactionHistoryList := types.TransactionHistoryPaginatedList{
					History: []types.TransactionHistory{
						{
							TxHash: "txHash",
							TxBody: "txBody",
						},
						{
							TxHash: "txHash2",
							TxBody: "txBody2",
						},
					},
					TotalCount: 2,
				}

				db.On("GetTransactionHistory", ctx, policyUUID, "SWAP", take, skip).
					Return(transactionHistoryList, nil)
			},
			expectErr:   false,
			expectedLen: 2,
		},
		{
			name:         "invalid policy UUID",
			policyID:     "nonexistent",
			mockSetup:    func(db *database.MockDB) {},
			expectErr:    true,
			errorMessage: "invalid policy_id",
			expectedLen:  0,
		},
		{
			name:     "database error",
			policyID: policyID,
			mockSetup: func(db *database.MockDB) {
				db.On("GetTransactionHistory", ctx, policyUUID, "SWAP", take, skip).
					Return(types.TransactionHistoryPaginatedList{}, fmt.Errorf("database error"))
			},
			expectErr:    true,
			errorMessage: "failed to get policy history",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(database.MockDB)

			tc.mockSetup(mockDB)

			policyService, err := NewPolicyService(mockDB, nil, nil, logrus.StandardLogger())
			require.NoError(t, err)

			policies, err := policyService.GetPluginPolicyTransactionHistory(ctx, tc.policyID, take, skip)

			if tc.expectErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
				require.Len(t, policies.History, tc.expectedLen)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
