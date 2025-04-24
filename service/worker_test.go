package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	rsyncer "github.com/vultisig/vultiserver-plugin/internal/syncer"
	"github.com/vultisig/vultiserver-plugin/internal/tasks"
	"github.com/vultisig/vultiserver-plugin/internal/types"
	"github.com/vultisig/vultiserver-plugin/test/mocks/database"
	"github.com/vultisig/vultiserver-plugin/test/mocks/plugin"
	"github.com/vultisig/vultiserver-plugin/test/mocks/queueclient"
	"github.com/vultisig/vultiserver-plugin/test/mocks/syncer"

	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GenerateToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	return args.Get(0).(*Claims), args.Error(1)
}

type MockInspector struct {
	mock.Mock
}

func (m *MockInspector) GetTaskInfo(queueName, taskID string) (*asynq.TaskInfo, error) {
	args := m.Called(queueName, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asynq.TaskInfo), args.Error(1)
}

func createTestWorkerService() (*WorkerService, *database.MockDB, *plugin.MockPlugin, *syncer.MockSyncer, *MockAuthService, *MockInspector, *queueclient.MockQueueClient) {
	mockDB := new(database.MockDB)
	mockPlugin := new(plugin.MockPlugin)
	mockSyncer := new(syncer.MockSyncer)
	mockAuthService := new(MockAuthService)
	mockInspector := new(MockInspector)
	mockQueueClient := new(queueclient.MockQueueClient)

	worker := &WorkerService{
		db:           mockDB,
		plugin:       mockPlugin,
		syncer:       mockSyncer,
		authService:  mockAuthService,
		inspector:    mockInspector,
		queueClient:  mockQueueClient,
		logger:       logrus.New(),
		verifierPort: 8080, // Default for tests
	}

	return worker, mockDB, mockPlugin, mockSyncer, mockAuthService, mockInspector, mockQueueClient
}

func TestHandlePluginTransaction(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name         string
		payload      types.PluginTriggerEvent
		mockSetup    func(*database.MockDB, *plugin.MockPlugin, *MockAuthService, *queueclient.MockQueueClient, *MockInspector, *syncer.MockSyncer)
		wantErr      bool
		errorMessage string
	}{
		{
			name:    "successful plugin transaction",
			payload: types.PluginTriggerEvent{PolicyID: "f1674509-df78-4982-8a7f-29c37c4ebe1c"},
			mockSetup: func(db *database.MockDB, plugin *plugin.MockPlugin, auth *MockAuthService, queue *queueclient.MockQueueClient, inspector *MockInspector, syncer *syncer.MockSyncer) {
				// Setup trigger status updates
				db.On("UpdateTriggerStatus", mock.Anything, "f1674509-df78-4982-8a7f-29c37c4ebe1c", types.StatusTimeTriggerPending).Return(nil)
				db.On("UpdateTimeTriggerLastExecution", mock.Anything, "f1674509-df78-4982-8a7f-29c37c4ebe1c").Return(nil)

				// Setup GetPluginPolicy
				policy := types.PluginPolicy{
					ID:             "f1674509-df78-4982-8a7f-29c37c4ebe1c",
					PublicKeyEcdsa: "public-key-123",
					PluginType:     "test-plugin",
				}
				db.On("GetPluginPolicy", mock.Anything, "f1674509-df78-4982-8a7f-29c37c4ebe1c").Return(policy, nil)

				// Setup ProposeTransactions
				signRequest := types.PluginKeysignRequest{
					PolicyID: "f1674509-df78-4982-8a7f-29c37c4ebe1c",
					KeysignRequest: types.KeysignRequest{
						PublicKey: "public-key-123",
						Messages:  []string{"message-hash-123"},
					},
					Transaction: "raw-tx-data",
				}
				plugin.On("ProposeTransactions", policy).Return([]types.PluginKeysignRequest{signRequest}, nil)

				// Setup auth token
				auth.On("GenerateToken").Return("jwt-token-123", nil)

				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)
				txID := uuid.New()
				db.On("CreateTransactionHistoryTx", ctx, nil, mock.AnythingOfType("types.TransactionHistory")).
					Return(txID, nil)
				syncer.On("SyncTransaction", rsyncer.CreateAction, "jwt-token-123", mock.AnythingOfType("types.TransactionHistory")).
					Return(nil)

				// Setup task enqueue
				taskInfo := &asynq.TaskInfo{ID: "task-123"}
				queue.On("Enqueue", mock.AnythingOfType("*asynq.Task"), mock.Anything).Return(taskInfo, nil)

				// Setup task result
				result := &asynq.TaskInfo{
					ID:     "task-123",
					State:  asynq.TaskStateCompleted,
					Result: []byte(`{"key1":{"r":"123","s":"456","v":"789"}}`),
				}
				inspector.On("GetTaskInfo", tasks.QUEUE_NAME, "task-123").Return(result, nil)

				// Setup signing complete
				plugin.On("SigningComplete", mock.Anything, mock.AnythingOfType("tss.KeysignResponse"), mock.AnythingOfType("types.PluginKeysignRequest"), policy).Return(nil)

				// Setup WithTransaction for UpdateTransactionStatusTx (SIGNED)
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)
				db.On("UpdateTransactionStatusTx", mock.Anything, nil, mock.AnythingOfType("uuid.UUID"), types.StatusSigned, mock.AnythingOfType("map[string]interface {}")).Return(nil)
				syncer.On("SyncTransaction", rsyncer.UpdateAction, "jwt-token-123", mock.AnythingOfType("types.TransactionHistory")).Return(nil)

				// Setup WithTransaction for UpdateTransactionStatusTx (MINED)
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)
				db.On("UpdateTransactionStatusTx", mock.Anything, nil, mock.AnythingOfType("uuid.UUID"), types.StatusMined, mock.AnythingOfType("map[string]interface {}")).Return(nil)
				syncer.On("SyncTransaction", rsyncer.UpdateAction, "jwt-token-123", mock.AnythingOfType("types.TransactionHistory")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "failed to get plugin policy",
			payload: types.PluginTriggerEvent{PolicyID: "policy-not-found"},
			mockSetup: func(db *database.MockDB, plugin *plugin.MockPlugin, auth *MockAuthService, queue *queueclient.MockQueueClient, inspector *MockInspector, syncer *syncer.MockSyncer) {
				// Setup trigger status updates for defer
				db.On("UpdateTriggerStatus", mock.Anything, "policy-not-found", types.StatusTimeTriggerPending).Return(nil)
				db.On("UpdateTimeTriggerLastExecution", mock.Anything, "policy-not-found").Return(nil)

				// Setup GetPluginPolicy to fail
				db.On("GetPluginPolicy", mock.Anything, "policy-not-found").Return(types.PluginPolicy{}, errors.New("policy not found"))
			},
			wantErr:      true,
			errorMessage: "db.GetPluginPolicy failed",
		},
		{
			name:    "failed to propose transactions",
			payload: types.PluginTriggerEvent{PolicyID: "policy-234"},
			mockSetup: func(db *database.MockDB, plugin *plugin.MockPlugin, auth *MockAuthService, queue *queueclient.MockQueueClient, inspector *MockInspector, syncer *syncer.MockSyncer) {
				// Setup trigger status updates
				db.On("UpdateTriggerStatus", mock.Anything, "policy-234", types.StatusTimeTriggerPending).Return(nil)
				db.On("UpdateTimeTriggerLastExecution", mock.Anything, "policy-234").Return(nil)

				// Setup GetPluginPolicy
				policy := types.PluginPolicy{
					ID:             "policy-234",
					PublicKeyEcdsa: "public-key-234",
					PluginType:     "test-plugin",
				}
				db.On("GetPluginPolicy", mock.Anything, "policy-234").Return(policy, nil)

				plugin.On("ProposeTransactions", policy).Return([]types.PluginKeysignRequest{}, errors.New("failed to propose transactions"))
			},
			wantErr:      true,
			errorMessage: "ProposeTransactions failed",
		},
		{
			name:    "Failed to Generate auth token",
			payload: types.PluginTriggerEvent{PolicyID: "policy-234"},
			mockSetup: func(db *database.MockDB, plugin *plugin.MockPlugin, auth *MockAuthService, queue *queueclient.MockQueueClient, inspector *MockInspector, syncer *syncer.MockSyncer) {
				// Setup trigger status updates
				db.On("UpdateTriggerStatus", mock.Anything, "policy-234", types.StatusTimeTriggerPending).Return(nil)
				db.On("UpdateTimeTriggerLastExecution", mock.Anything, "policy-234").Return(nil)

				// Setup GetPluginPolicy
				policy := types.PluginPolicy{
					ID:             "policy-234",
					PublicKeyEcdsa: "public-key-234",
					PluginType:     "test-plugin",
				}
				db.On("GetPluginPolicy", mock.Anything, "policy-234").Return(policy, nil)

				// Setup ProposeTransactions
				signRequest := types.PluginKeysignRequest{
					PolicyID: "f1674509-df78-4982-8a7f-29c37c4ebe1c",
					KeysignRequest: types.KeysignRequest{
						PublicKey: "public-key-123",
						Messages:  []string{"message-hash-123"},
					},
					Transaction: "raw-tx-data",
				}
				plugin.On("ProposeTransactions", policy).Return([]types.PluginKeysignRequest{signRequest}, nil)

				// Setup auth token
				auth.On("GenerateToken").Return("", errors.New("failed to generate token"))
			},
			wantErr:      true,
			errorMessage: "GenerateToken failed",
		},
		{
			name:    "Failed to parse policy UUID",
			payload: types.PluginTriggerEvent{PolicyID: "policy-234"},
			mockSetup: func(db *database.MockDB, plugin *plugin.MockPlugin, auth *MockAuthService, queue *queueclient.MockQueueClient, inspector *MockInspector, syncer *syncer.MockSyncer) {
				// Setup trigger status updates
				db.On("UpdateTriggerStatus", mock.Anything, "policy-234", types.StatusTimeTriggerPending).Return(nil)
				db.On("UpdateTimeTriggerLastExecution", mock.Anything, "policy-234").Return(nil)

				// Setup GetPluginPolicy
				policy := types.PluginPolicy{
					ID:             "policy-234",
					PublicKeyEcdsa: "public-key-234",
					PluginType:     "test-plugin",
				}
				db.On("GetPluginPolicy", mock.Anything, "policy-234").Return(policy, nil)

				// Setup ProposeTransactions
				signRequest := types.PluginKeysignRequest{
					PolicyID: "policy-234",
					KeysignRequest: types.KeysignRequest{
						PublicKey: "public-key-123",
						Messages:  []string{"message-hash-123"},
					},
					Transaction: "raw-tx-data",
				}
				plugin.On("ProposeTransactions", policy).Return([]types.PluginKeysignRequest{signRequest}, nil)

				// Setup auth token
				auth.On("GenerateToken").Return("jwt-token-123", nil)
			},
			wantErr:      true,
			errorMessage: "failed to parse policy UUID",
		},
		{
			name:    "Create Transaction sync fail",
			payload: types.PluginTriggerEvent{PolicyID: "policy-234"},
			mockSetup: func(db *database.MockDB, plugin *plugin.MockPlugin, auth *MockAuthService, queue *queueclient.MockQueueClient, inspector *MockInspector, syncer *syncer.MockSyncer) {
				// Setup trigger status updates
				db.On("UpdateTriggerStatus", mock.Anything, "policy-234", types.StatusTimeTriggerPending).Return(nil)
				db.On("UpdateTimeTriggerLastExecution", mock.Anything, "policy-234").Return(nil)

				// Setup GetPluginPolicy
				policy := types.PluginPolicy{
					ID:             "policy-234",
					PublicKeyEcdsa: "public-key-234",
					PluginType:     "test-plugin",
				}
				db.On("GetPluginPolicy", mock.Anything, "policy-234").Return(policy, nil)

				// Setup ProposeTransactions
				signRequest := types.PluginKeysignRequest{
					PolicyID: "f1674509-df78-4982-8a7f-29c37c4ebe1c",
					KeysignRequest: types.KeysignRequest{
						PublicKey: "public-key-123",
						Messages:  []string{"message-hash-123"},
					},
					Transaction: "raw-tx-data",
				}
				plugin.On("ProposeTransactions", policy).Return([]types.PluginKeysignRequest{signRequest}, nil)

				// Setup auth token
				auth.On("GenerateToken").Return("jwt-token-123", nil)

				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(false, errors.New("failed to execute transaction"))
			},
			wantErr:      true,
			errorMessage: "upsertAndSyncTransaction failed",
		},
		{
			name:    "Enqueue KeySign task fail",
			payload: types.PluginTriggerEvent{PolicyID: "f1674509-df78-4982-8a7f-29c37c4ebe1c"},
			mockSetup: func(db *database.MockDB, plugin *plugin.MockPlugin, auth *MockAuthService, queue *queueclient.MockQueueClient, inspector *MockInspector, syncer *syncer.MockSyncer) {
				// Setup trigger status updates
				db.On("UpdateTriggerStatus", mock.Anything, "f1674509-df78-4982-8a7f-29c37c4ebe1c", types.StatusTimeTriggerPending).Return(nil)
				db.On("UpdateTimeTriggerLastExecution", mock.Anything, "f1674509-df78-4982-8a7f-29c37c4ebe1c").Return(nil)

				// Setup GetPluginPolicy
				policy := types.PluginPolicy{
					ID:             "f1674509-df78-4982-8a7f-29c37c4ebe1c",
					PublicKeyEcdsa: "public-key-123",
					PluginType:     "test-plugin",
				}
				db.On("GetPluginPolicy", mock.Anything, "f1674509-df78-4982-8a7f-29c37c4ebe1c").Return(policy, nil)

				// Setup ProposeTransactions
				signRequest := types.PluginKeysignRequest{
					PolicyID: "f1674509-df78-4982-8a7f-29c37c4ebe1c",
					KeysignRequest: types.KeysignRequest{
						PublicKey: "public-key-123",
						Messages:  []string{"message-hash-123"},
					},
					Transaction: "raw-tx-data",
				}
				plugin.On("ProposeTransactions", policy).Return([]types.PluginKeysignRequest{signRequest}, nil)

				// Setup auth token
				auth.On("GenerateToken").Return("jwt-token-123", nil)

				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)
				txID := uuid.New()
				db.On("CreateTransactionHistoryTx", ctx, nil, mock.AnythingOfType("types.TransactionHistory")).
					Return(txID, nil)
				syncer.On("SyncTransaction", rsyncer.CreateAction, "jwt-token-123", mock.AnythingOfType("types.TransactionHistory")).
					Return(nil)

				// Setup task enqueue
				taskInfo := &asynq.TaskInfo{ID: "task-123"}
				queue.On("Enqueue", mock.AnythingOfType("*asynq.Task"), mock.Anything).Return(taskInfo, errors.New("failed to enqueue transaction"))
			},
			wantErr:      true,
			errorMessage: "failed to enqueue signing task",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create service with mocks
			worker, db, plugin, syncer, auth, inspector, queue := createTestWorkerService()

			// Setup HTTP server for verifier
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success"}`))
			}))
			defer server.Close()

			// Extract port from server.URL
			parts := bytes.Split([]byte(server.URL), []byte(":"))
			port := string(parts[len(parts)-1])
			portInt, _ := strconv.Atoi(port)
			worker.verifierPort = int64(portInt)

			// Configure mocks
			tc.mockSetup(db, plugin, auth, queue, inspector, syncer)

			// Create task payload
			payload, err := json.Marshal(tc.payload)
			require.NoError(t, err)
			task := asynq.NewTask("plugin:transaction", payload)

			// Execute the function
			err = worker.HandlePluginTransaction(ctx, task)

			// Check results
			if tc.wantErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
			}

			// Verify all mocks were called as expected
			db.AssertExpectations(t)
			plugin.AssertExpectations(t)
			auth.AssertExpectations(t)
			queue.AssertExpectations(t)
			inspector.AssertExpectations(t)
			syncer.AssertExpectations(t)
		})
	}
}

func TestInitiateTxSignWithVerifier(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		signRequest    types.PluginKeysignRequest
		metadata       map[string]interface{}
		newTx          types.TransactionHistory
		jwtToken       string
		serverStatus   int
		serverResponse string
		mockSetup      func(*database.MockDB, *syncer.MockSyncer)
		wantErr        bool
		errorMessage   string
	}{
		{
			name: "successful verifier interaction",
			signRequest: types.PluginKeysignRequest{
				PolicyID: "policy-123",
				KeysignRequest: types.KeysignRequest{
					PublicKey: "public-key-123",
				},
			},
			metadata:       map[string]interface{}{"timestamp": time.Now()},
			newTx:          types.TransactionHistory{ID: uuid.New(), Status: types.StatusPending},
			jwtToken:       "jwt-token-123",
			serverStatus:   http.StatusOK,
			serverResponse: `{"status":"success"}`,
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer) {
				// No mock setup needed for success case
			},
			wantErr: false,
		},
		{
			name: "verifier returns error",
			signRequest: types.PluginKeysignRequest{
				PolicyID: "policy-456",
				KeysignRequest: types.KeysignRequest{
					PublicKey: "public-key-456",
				},
			},
			metadata:       map[string]interface{}{"timestamp": time.Now()},
			newTx:          types.TransactionHistory{ID: uuid.New(), Status: types.StatusPending},
			jwtToken:       "jwt-token-456",
			serverStatus:   http.StatusBadRequest,
			serverResponse: `{"error":"invalid signature"}`,
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer) {
				// Setup transaction update for error case
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)
				db.On("UpdateTransactionStatusTx", mock.Anything, nil, mock.AnythingOfType("uuid.UUID"),
					types.StatusSigningFailed, mock.AnythingOfType("map[string]interface {}")).Return(nil)
				syncer.On("SyncTransaction", rsyncer.UpdateAction, "jwt-token-456", mock.AnythingOfType("types.TransactionHistory")).
					Return(nil)
			},
			wantErr:      true,
			errorMessage: "verifier responded with error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			worker, db, _, msyncer, _, _, _ := createTestWorkerService()

			tc.mockSetup(db, msyncer)

			var server *httptest.Server

			if tc.name != "connection to verifier fails" {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.serverStatus)
					w.Write([]byte(tc.serverResponse))
				}))
				defer server.Close()

				parts := bytes.Split([]byte(server.URL), []byte(":"))
				port := string(parts[len(parts)-1])
				portInt, _ := strconv.Atoi(port)
				worker.verifierPort = int64(portInt)
			} else {
				worker.verifierPort = 9999
			}

			err := worker.initiateTxSignWithVerifier(ctx, tc.signRequest, tc.metadata, tc.newTx, tc.jwtToken)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
			}

			db.AssertExpectations(t)
			msyncer.AssertExpectations(t)
		})
	}
}

func TestWaitForTaskResult(t *testing.T) {
	worker, _, _, _, _, _, _ := createTestWorkerService()

	tests := []struct {
		name           string
		taskID         string
		mockSetup      func(*MockInspector)
		expectedResult []byte
		wantErr        bool
		errorMessage   string
	}{
		{
			name:   "task completes within timeout",
			taskID: "task-123",
			mockSetup: func(inspector *MockInspector) {
				// Return a completed task with result data
				taskInfo := &asynq.TaskInfo{
					ID:     "task-123",
					State:  asynq.TaskStateCompleted,
					Result: []byte(`{"key1":"value1"}`),
				}
				inspector.On("GetTaskInfo", tasks.QUEUE_NAME, "task-123").Return(taskInfo, nil)
			},
			expectedResult: []byte(`{"key1":"value1"}`),
			wantErr:        false,
		},
		{
			name:   "task fails and gets archived",
			taskID: "task-456",
			mockSetup: func(inspector *MockInspector) {
				// First call returns pending
				pendingTask := &asynq.TaskInfo{
					ID:    "task-456",
					State: asynq.TaskStatePending,
				}
				// Second call returns archived (failed)
				archivedTask := &asynq.TaskInfo{
					ID:      "task-456",
					State:   asynq.TaskStateArchived,
					LastErr: "execution failed",
				}
				inspector.On("GetTaskInfo", tasks.QUEUE_NAME, "task-456").Return(pendingTask, nil).Once()
				inspector.On("GetTaskInfo", tasks.QUEUE_NAME, "task-456").Return(archivedTask, nil).Once()
			},
			wantErr:      true,
			errorMessage: "task archived",
		},
		{
			name:   "error retrieving task info",
			taskID: "task-789",
			mockSetup: func(inspector *MockInspector) {
				inspector.On("GetTaskInfo", tasks.QUEUE_NAME, "task-789").Return(nil, errors.New("task not found"))
			},
			wantErr:      true,
			errorMessage: "failed to get task info: task not found",
		},
		{
			name:   "unexpected task state",
			taskID: "task-999",
			mockSetup: func(inspector *MockInspector) {
				unexpectedTask := &asynq.TaskInfo{
					ID:    "task-999",
					State: asynq.TaskStateAggregating + 1,
				}
				inspector.On("GetTaskInfo", tasks.QUEUE_NAME, "task-999").Return(unexpectedTask, nil)
			},
			wantErr:      true,
			errorMessage: "unexpected task state",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockInspector := new(MockInspector)
			tc.mockSetup(mockInspector)
			worker.inspector = mockInspector

			result, err := worker.waitForTaskResult(tc.taskID, 3*time.Second)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}

			mockInspector.AssertExpectations(t)
		})
	}
}

func TestUpsertAndSyncTransaction(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name         string
		action       rsyncer.Action
		transaction  *types.TransactionHistory
		jwtToken     string
		mockSetup    func(*database.MockDB, *syncer.MockSyncer)
		wantErr      bool
		errorMessage string
	}{
		{
			name:   "successfully create transaction",
			action: rsyncer.CreateAction,
			transaction: &types.TransactionHistory{
				PolicyID: uuid.New(),
				TxBody:   "tx-data",
				TxHash:   "tx-hash",
				Status:   types.StatusPending,
				Metadata: map[string]interface{}{"key": "value"},
			},
			jwtToken: "jwt-token",
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer) {
				// Setup transaction
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				db.On("CreateTransactionHistoryTx", ctx, nil, mock.AnythingOfType("types.TransactionHistory")).
					Return(uuid.New(), nil)

				syncer.On("SyncTransaction", rsyncer.CreateAction, "jwt-token",
					mock.AnythingOfType("types.TransactionHistory")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "successfully update transaction",
			action: rsyncer.UpdateAction,
			transaction: &types.TransactionHistory{
				ID:       uuid.New(),
				PolicyID: uuid.New(),
				TxBody:   "tx-data",
				TxHash:   "tx-hash",
				Status:   types.StatusSigned,
				Metadata: map[string]interface{}{"key": "value"},
			},
			jwtToken: "jwt-token",
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer) {
				// Setup transaction
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				db.On("UpdateTransactionStatusTx", ctx, nil, mock.AnythingOfType("uuid.UUID"),
					types.StatusSigned, mock.AnythingOfType("map[string]interface {}")).Return(nil)

				syncer.On("SyncTransaction", rsyncer.UpdateAction, "jwt-token",
					mock.AnythingOfType("types.TransactionHistory")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "create transaction db error",
			action: rsyncer.CreateAction,
			transaction: &types.TransactionHistory{
				PolicyID: uuid.New(),
				TxBody:   "tx-data",
				TxHash:   "tx-hash",
				Status:   types.StatusPending,
				Metadata: map[string]interface{}{"key": "value"},
			},
			jwtToken: "jwt-token",
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer) {
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				db.On("CreateTransactionHistoryTx", ctx, nil, mock.AnythingOfType("types.TransactionHistory")).
					Return(uuid.New(), errors.New("db error"))
			},
			wantErr:      true,
			errorMessage: "failed to create transaction history",
		},
		{
			name:   "update transaction db error",
			action: rsyncer.UpdateAction,
			transaction: &types.TransactionHistory{
				ID:       uuid.New(),
				PolicyID: uuid.New(),
				TxBody:   "tx-data",
				TxHash:   "tx-hash",
				Status:   types.StatusSigned,
				Metadata: map[string]interface{}{"key": "value"},
			},
			jwtToken: "jwt-token",
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer) {
				// Setup transaction to fail at DB update
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				db.On("UpdateTransactionStatusTx", ctx, nil, mock.AnythingOfType("uuid.UUID"),
					types.StatusSigned, mock.AnythingOfType("map[string]interface {}")).Return(errors.New("db error"))
			},
			wantErr:      true,
			errorMessage: "failed to update transaction status",
		},
		{
			name:   "sync error",
			action: rsyncer.CreateAction,
			transaction: &types.TransactionHistory{
				PolicyID: uuid.New(),
				TxBody:   "tx-data",
				TxHash:   "tx-hash",
				Status:   types.StatusPending,
				Metadata: map[string]interface{}{"key": "value"},
			},
			jwtToken: "jwt-token",
			mockSetup: func(db *database.MockDB, syncer *syncer.MockSyncer) {
				// Setup transaction to fail at sync
				db.On("WithTransaction", ctx, mock.AnythingOfType("func(context.Context, pgx.Tx) error")).
					Return(true, nil)

				db.On("CreateTransactionHistoryTx", ctx, nil, mock.AnythingOfType("types.TransactionHistory")).
					Return(uuid.New(), nil)

				syncer.On("SyncTransaction", rsyncer.CreateAction, "jwt-token",
					mock.AnythingOfType("types.TransactionHistory")).Return(errors.New("sync error"))
			},
			wantErr:      true,
			errorMessage: "failed to sync transaction",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create service with mocks
			worker, db, _, syncer, _, _, _ := createTestWorkerService()

			// Configure mocks
			tc.mockSetup(db, syncer)

			// Execute the function
			err := worker.upsertAndSyncTransaction(ctx, tc.action, tc.transaction, tc.jwtToken)

			// Check results
			if tc.wantErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
			}

			// Verify all mocks were called as expected
			db.AssertExpectations(t)
			syncer.AssertExpectations(t)
		})
	}
}
