package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vultisig/vultiserver-plugin/internal/types"
	"github.com/vultisig/vultiserver-plugin/test/mocks/database"
	"github.com/vultisig/vultiserver-plugin/test/mocks/queueclient"
)

// MockClock is a mock implementation of the Clock interface
type MockClock struct {
	mock.Mock
}

func (m *MockClock) Now() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func (m *MockClock) NewTicker(d time.Duration) *time.Ticker {
	args := m.Called(d)
	return args.Get(0).(*time.Ticker)
}

func TestFrequencyToCron(t *testing.T) {
	now := time.Date(2023, 1, 1, 14, 30, 0, 0, time.UTC)

	testCases := []struct {
		name      string
		frequency string
		startTime time.Time
		interval  int
		expected  string
	}{
		{
			name:      "minutely with interval 1",
			frequency: "minutely",
			startTime: now,
			interval:  1,
			expected:  "*/1 * * * *",
		},
		{
			name:      "minutely with interval 5",
			frequency: "minutely",
			startTime: now,
			interval:  5,
			expected:  "*/5 * * * *",
		},
		{
			name:      "hourly with interval 1",
			frequency: "hourly",
			startTime: now,
			interval:  1,
			expected:  "30 * * * *",
		},
		{
			name:      "hourly with interval 5",
			frequency: "hourly",
			startTime: now,
			interval:  5,
			expected:  "30 */5 * * *",
		},
		{
			name:      "daily with interval 1",
			frequency: "daily",
			startTime: now,
			interval:  1,
			expected:  "30 14 * * *",
		},
		{
			name:      "weekly with interval 1",
			frequency: "weekly",
			startTime: now,
			interval:  1,
			expected:  "30 14 * * 0",
		},
		{
			name:      "monthly with interval 1",
			frequency: "monthly",
			startTime: now,
			interval:  1,
			expected:  "30 14 1 * *",
		},
		{
			name:      "unknown frequency",
			frequency: "unknown",
			startTime: now,
			interval:  1,
			expected:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FrequencyToCron(tc.frequency, tc.startTime, tc.interval)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestIntervalScheduleNext(t *testing.T) {
	baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)

	testCases := []struct {
		name        string
		schedule    *IntervalSchedule
		currentTime time.Time
		expected    time.Time
	}{
		{
			name: "daily with interval 1",
			schedule: &IntervalSchedule{
				Frequency: "daily",
				Interval:  1,
				StartTime: baseTime,
				Hour:      10,
				Minute:    0,
				Location:  time.UTC,
			},
			currentTime: time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			expected:    time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "daily with interval 1 - same day after time",
			schedule: &IntervalSchedule{
				Frequency: "daily",
				Interval:  1,
				StartTime: baseTime,
				Hour:      10,
				Minute:    0,
				Location:  time.UTC,
			},
			currentTime: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC),
			expected:    time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "daily with interval 2 - next day is valid",
			schedule: &IntervalSchedule{
				Frequency: "daily",
				Interval:  2,
				StartTime: baseTime,
				Hour:      10,
				Minute:    0,
				Location:  time.UTC,
			},
			currentTime: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC),
			expected:    time.Date(2023, 1, 3, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "weekly with interval 1 - next week",
			schedule: &IntervalSchedule{
				Frequency: "weekly",
				Interval:  1,
				StartTime: baseTime,
				Hour:      10,
				Minute:    0,
				Weekday:   time.Sunday,
				Location:  time.UTC,
			},
			currentTime: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC),
			expected:    time.Date(2023, 1, 8, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "weekly with interval 2 - skip a week",
			schedule: &IntervalSchedule{
				Frequency: "weekly",
				Interval:  2,
				StartTime: baseTime,
				Hour:      10,
				Minute:    0,
				Weekday:   time.Sunday,
				Location:  time.UTC,
			},
			currentTime: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC),
			expected:    time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "monthly with interval 1 - next month",
			schedule: &IntervalSchedule{
				Frequency: "monthly",
				Interval:  1,
				StartTime: baseTime,
				Day:       1,
				Hour:      10,
				Minute:    0,
				Location:  time.UTC,
			},
			currentTime: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC),
			expected:    time.Date(2023, 2, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "monthly with interval 2 - skip a month",
			schedule: &IntervalSchedule{
				Frequency: "monthly",
				Interval:  2,
				StartTime: baseTime,
				Day:       1,
				Hour:      10,
				Minute:    0,
				Location:  time.UTC,
			},
			currentTime: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC),
			expected:    time.Date(2023, 3, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "unsupported frequency",
			schedule: &IntervalSchedule{
				Frequency: "unsupported",
				Interval:  1,
				StartTime: baseTime,
				Hour:      10,
				Minute:    0,
				Location:  time.UTC,
			},
			currentTime: time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			expected:    time.Time{}, // Zero time for unsupported frequency
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.schedule.Next(tc.currentTime)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestNewIntervalSchedule(t *testing.T) {
	startTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)

	testCases := []struct {
		name      string
		frequency string
		startTime time.Time
		interval  int
		expectErr bool
	}{
		{
			name:      "valid daily schedule",
			frequency: "daily",
			interval:  1,
			expectErr: false,
		},
		{
			name:      "valid weekly schedule",
			frequency: "weekly",
			interval:  2,
			expectErr: false,
		},
		{
			name:      "valid monthly schedule",
			frequency: "monthly",
			interval:  3,
			expectErr: false,
		},
		{
			name:      "invalid interval zero",
			frequency: "daily",
			startTime: startTime,
			interval:  0,
			expectErr: true,
		},
		{
			name:      "invalid interval negative",
			frequency: "daily",
			startTime: startTime,
			interval:  -1,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			schedule, err := NewIntervalSchedule(tc.frequency, tc.startTime, tc.interval)
			if tc.expectErr {
				require.Error(t, err)
				require.Nil(t, schedule)
			} else {
				require.NoError(t, err)
				require.NotNil(t, schedule)
				require.Equal(t, tc.frequency, schedule.Frequency)
				require.Equal(t, tc.interval, schedule.Interval)
				require.Equal(t, tc.startTime, schedule.StartTime)
				require.Equal(t, tc.startTime.Minute(), schedule.Minute)
				require.Equal(t, tc.startTime.Hour(), schedule.Hour)
				require.Equal(t, tc.startTime.Day(), schedule.Day)
				require.Equal(t, tc.startTime.Weekday(), schedule.Weekday)
				require.Equal(t, tc.startTime.Location(), schedule.Location)
			}
		})
	}
}

func TestCreateSchedule(t *testing.T) {
	startTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)

	testCases := []struct {
		name         string
		cronExpr     string
		frequency    string
		startTime    time.Time
		interval     int
		wantErr      bool
		expectedType string
	}{
		{
			name:         "standard cron expression",
			cronExpr:     "0 10 * * *",
			frequency:    "custom",
			startTime:    startTime,
			interval:     1,
			wantErr:      false,
			expectedType: "cron.SpecSchedule",
		},
		{
			name:         "daily with interval > 1",
			cronExpr:     "0 10 * * *",
			frequency:    "daily",
			startTime:    startTime,
			interval:     2,
			wantErr:      false,
			expectedType: "*scheduler.IntervalSchedule",
		},
		{
			name:         "weekly with interval > 1",
			cronExpr:     "0 10 * * 0",
			frequency:    "weekly",
			startTime:    startTime,
			interval:     2,
			wantErr:      false,
			expectedType: "*scheduler.IntervalSchedule",
		},
		{
			name:         "monthly with interval > 1",
			cronExpr:     "0 10 1 * *",
			frequency:    "monthly",
			startTime:    startTime,
			interval:     2,
			wantErr:      false,
			expectedType: "*scheduler.IntervalSchedule",
		},
		{
			name:         "invalid cron expression",
			cronExpr:     "invalid cron",
			frequency:    "custom",
			startTime:    startTime,
			interval:     1,
			wantErr:      true,
			expectedType: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schedule, err := CreateSchedule(tc.cronExpr, tc.frequency, tc.startTime, tc.interval)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, schedule)
			} else {
				require.NoError(t, err)
				require.NotNil(t, schedule)
				require.Contains(t, fmt.Sprintf("%T", schedule), tc.expectedType)
			}
		})
	}
}

func TestProcessTrigger(t *testing.T) {
	baseTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	ctx := context.Background()

	testCases := []struct {
		name         string
		trigger      types.TimeTrigger
		mockSetup    func(*database.MockDB, *MockClock, *queueclient.MockQueueClient)
		wantErr      bool
		errorMessage string
	}{
		{
			name: "invalid schedule",
			trigger: types.TimeTrigger{
				PolicyID:       "policy1",
				CronExpression: "invalid cron",
				StartTime:      baseTime,
				Frequency:      "custom",
				Interval:       1,
			},
			mockSetup: func(db *database.MockDB, clock *MockClock, client *queueclient.MockQueueClient) {
				db.On("DeleteTimeTrigger", ctx, "policy1").Return(nil)
			},
			wantErr:      true,
			errorMessage: "invalid schedule",
		},
		{
			name: "end time reached",
			trigger: types.TimeTrigger{
				PolicyID:       "policy1",
				CronExpression: "0 10 * * *",
				StartTime:      baseTime,
				EndTime:        timePtr(baseTime.Add(-24 * time.Hour)),
				Frequency:      "daily",
				Interval:       1,
			},
			mockSetup: func(db *database.MockDB, clock *MockClock, client *queueclient.MockQueueClient) {
				clock.On("Now").Return(baseTime)
				db.On("DeleteTimeTrigger", ctx, "policy1").Return(nil)
			},
			wantErr:      true,
			errorMessage: ErrEndTimeReached.Error(),
		},
		{
			name: "trigger next time not reached",
			trigger: types.TimeTrigger{
				PolicyID:       "policy1",
				CronExpression: "0 12 * * *",
				StartTime:      baseTime,
				Frequency:      "daily",
				Interval:       1,
			},
			mockSetup: func(db *database.MockDB, clock *MockClock, client *queueclient.MockQueueClient) {
				clock.On("Now").Return(baseTime.Add(-1 * time.Hour))
				db.On("GetTriggerStatus", ctx, "policy1").Return(types.StatusTimeTriggerPending, nil)
			},
			wantErr:      true,
			errorMessage: ErrTriggerNotReady.Error(),
		},
		{
			name: "trigger already in RUNNING status",
			trigger: types.TimeTrigger{
				PolicyID:       "policy1",
				CronExpression: "0 10 * * *",
				StartTime:      baseTime,
				Frequency:      "daily",
				Interval:       1,
			},
			mockSetup: func(db *database.MockDB, clock *MockClock, client *queueclient.MockQueueClient) {
				clock.On("Now").Return(baseTime)
				db.On("GetTriggerStatus", ctx, "policy1").Return(types.StatusTimeTriggerRunning, nil)
			},
			wantErr:      true,
			errorMessage: ErrTriggerNotReady.Error(),
		},
		{
			name: "time trigger enqueue success",
			trigger: types.TimeTrigger{
				PolicyID:       "policy1",
				CronExpression: "0 10 * * *",
				StartTime:      baseTime,
				LastExecution:  timePtr(baseTime.Add(-24 * time.Hour)),
				Frequency:      "daily",
				Interval:       1,
			},
			mockSetup: func(db *database.MockDB, clock *MockClock, client *queueclient.MockQueueClient) {
				clock.On("Now").Return(baseTime)
				db.On("GetTriggerStatus", ctx, "policy1").Return(types.StatusTimeTriggerPending, nil)
				db.On("UpdateTriggerStatus", ctx, "policy1", types.StatusTimeTriggerRunning).Return(nil)

				client.On("Enqueue", mock.Anything, mock.Anything).Return(&asynq.TaskInfo{ID: "task123"}, nil)
			},
			wantErr: false,
		},

		{
			name: "time trigger enqueue fail",
			trigger: types.TimeTrigger{
				PolicyID:       "policy1",
				CronExpression: "0 10 * * *",
				StartTime:      baseTime,
				Frequency:      "daily",
				Interval:       1,
			},
			mockSetup: func(db *database.MockDB, clock *MockClock, client *queueclient.MockQueueClient) {
				clock.On("Now").Return(baseTime)
				db.On("GetTriggerStatus", ctx, "policy1").Return(types.StatusTimeTriggerPending, nil)
				db.On("UpdateTriggerStatus", ctx, "policy1", types.StatusTimeTriggerRunning).Return(nil)

				client.On("Enqueue", mock.Anything, mock.Anything).Return(&asynq.TaskInfo{}, fmt.Errorf("enqueue failed"))
			},
			wantErr:      true,
			errorMessage: "failed to enqueue trigger task",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockDB := new(database.MockDB)
			mockClock := new(MockClock)
			mockClient := new(queueclient.MockQueueClient)

			tc.mockSetup(mockDB, mockClock, mockClient)

			scheduler := SchedulerService{
				db:         mockDB,
				logger:     logrus.StandardLogger(),
				clock:      mockClock,
				taskClient: mockClient,
				inspector:  nil,
				done:       make(chan struct{}),
			}

			// Call method under test
			err := scheduler.processTrigger(ctx, tc.trigger)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
			mockClock.AssertExpectations(t)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestCheckAndEnqueueTasks(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)

	testCases := []struct {
		name         string
		mockSetup    func(db *database.MockDB, clock *MockClock, client *queueclient.MockQueueClient)
		wantErr      bool
		errorMessage string
	}{
		{
			name: "error fetching pending triggers",
			mockSetup: func(db *database.MockDB, clock *MockClock, client *queueclient.MockQueueClient) {
				db.On("GetPendingTimeTriggers", mock.Anything).Return([]types.TimeTrigger{}, fmt.Errorf("database error"))
			},
			wantErr:      true,
			errorMessage: "failed to get pending triggers",
		},
		{
			name: "no triggers found",
			mockSetup: func(db *database.MockDB, clock *MockClock, client *queueclient.MockQueueClient) {
				db.On("GetPendingTimeTriggers", mock.Anything).Return([]types.TimeTrigger{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error processing one trigger but continue with other",
			mockSetup: func(db *database.MockDB, clock *MockClock, client *queueclient.MockQueueClient) {
				triggers := []types.TimeTrigger{
					{
						PolicyID:       "policy1",
						CronExpression: "invalid cron",
						StartTime:      baseTime,
						Frequency:      "daily",
						Interval:       1,
					},
					{
						PolicyID:       "policy2",
						CronExpression: "0 10 * * *",
						StartTime:      baseTime.Add(-24 * time.Hour),
						LastExecution:  timePtr(baseTime.Add(-48 * time.Hour)),
						Frequency:      "daily",
						Interval:       1,
					},
				}

				db.On("GetPendingTimeTriggers", mock.Anything).Return(triggers, nil)

				// First trigger fails for invalid cron
				db.On("DeleteTimeTrigger", mock.Anything, "policy1").Return(nil)

				// Second trigger success
				clock.On("Now").Return(baseTime)
				db.On("GetTriggerStatus", mock.Anything, "policy2").Return(types.StatusTimeTriggerPending, nil)
				db.On("UpdateTriggerStatus", mock.Anything, "policy2", types.StatusTimeTriggerRunning).Return(nil)
				client.On("Enqueue", mock.Anything, mock.Anything).Return(&asynq.TaskInfo{ID: "task123"}, nil)
			},
			wantErr: false,
		},
		{
			name: "all triggers processed successfully",
			mockSetup: func(db *database.MockDB, clock *MockClock, client *queueclient.MockQueueClient) {

				triggers := []types.TimeTrigger{
					{
						PolicyID:       "policy1",
						CronExpression: "0 9 * * *", // Valid cron
						StartTime:      baseTime.Add(-24 * time.Hour),
						LastExecution:  timePtr(baseTime.Add(-48 * time.Hour)),
						Frequency:      "daily",
						Interval:       1,
					},
					{
						PolicyID:       "policy2",
						CronExpression: "0 8 * * *", // Valid cron
						StartTime:      baseTime.Add(-24 * time.Hour),
						LastExecution:  timePtr(baseTime.Add(-48 * time.Hour)),
						Frequency:      "daily",
						Interval:       1,
					},
				}

				db.On("GetPendingTimeTriggers", mock.Anything).Return(triggers, nil)

				//Both triggers success
				clock.On("Now").Return(baseTime).Times(2)
				db.On("GetTriggerStatus", mock.Anything, "policy1").Return(types.StatusTimeTriggerPending, nil)
				db.On("GetTriggerStatus", mock.Anything, "policy2").Return(types.StatusTimeTriggerPending, nil)
				db.On("UpdateTriggerStatus", mock.Anything, "policy1", types.StatusTimeTriggerRunning).Return(nil)
				db.On("UpdateTriggerStatus", mock.Anything, "policy2", types.StatusTimeTriggerRunning).Return(nil)
				client.On("Enqueue", mock.Anything, mock.Anything).Return(&asynq.TaskInfo{ID: "task1"}, nil).Once()
				client.On("Enqueue", mock.Anything, mock.Anything).Return(&asynq.TaskInfo{ID: "task2"}, nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(database.MockDB)
			mockClock := new(MockClock)
			mockClient := new(queueclient.MockQueueClient)

			tc.mockSetup(mockDB, mockClock, mockClient)

			scheduler := SchedulerService{
				db:         mockDB,
				logger:     logrus.StandardLogger(),
				clock:      mockClock,
				taskClient: mockClient,
				inspector:  nil,
				done:       make(chan struct{}),
			}

			err := scheduler.checkAndEnqueueTasks()

			if tc.wantErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
			mockClock.AssertExpectations(t)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetTriggerFromPolicy(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)

	testCases := []struct {
		name            string
		policy          types.PluginPolicy
		mockTime        time.Time
		wantErr         bool
		errorMessage    string
		validateTrigger func(*testing.T, *types.TimeTrigger)
	}{
		{
			name: "valid daily policy",
			policy: types.PluginPolicy{
				ID: "policy1",
				Policy: json.RawMessage(`{
					"schedule": {
						"frequency": "daily",
						"start_time": "2023-01-01T10:00:00Z",
						"interval": "1"
					}
				}`),
			},
			mockTime: baseTime,
			wantErr:  false,
			validateTrigger: func(t *testing.T, trigger *types.TimeTrigger) {
				lastExec := (*time.Time)(nil)
				require.Equal(t, "policy1", trigger.PolicyID)
				require.Equal(t, "daily", trigger.Frequency)
				require.Equal(t, 1, trigger.Interval)
				require.Equal(t, "0 10 * * *", trigger.CronExpression) // Based on frequencyToCron
				require.Equal(t, baseTime, trigger.StartTime)
				require.Equal(t, lastExec, trigger.LastExecution)
				require.Equal(t, types.StatusTimeTriggerPending, trigger.Status)
			},
		},
		{
			name: "valid weekly policy with end time",
			policy: types.PluginPolicy{
				ID: "policy2",
				Policy: json.RawMessage(`{
					"schedule": {
						"frequency": "weekly",
						"start_time": "2023-01-01T10:00:00Z",
						"interval": "2",
						"end_time": "2023-06-01T10:00:00Z"
					}
				}`),
			},
			mockTime: baseTime,
			wantErr:  false,
			validateTrigger: func(t *testing.T, trigger *types.TimeTrigger) {
				require.Equal(t, "policy2", trigger.PolicyID)
				require.Equal(t, "weekly", trigger.Frequency)
				require.Equal(t, 2, trigger.Interval)
				require.Equal(t, baseTime, trigger.StartTime)
				require.NotNil(t, trigger.EndTime)
				endTime := time.Date(2023, 6, 1, 10, 0, 0, 0, time.UTC)
				require.Equal(t, endTime, *trigger.EndTime)
			},
		},
		{
			name: "invalid policy json",
			policy: types.PluginPolicy{
				ID:     "policy3",
				Policy: json.RawMessage(`{invalid json`),
			},
			mockTime:     baseTime,
			wantErr:      true,
			errorMessage: "failed to parse policy schedule",
		},
		{
			name: "missing schedule fields",
			policy: types.PluginPolicy{
				ID: "policy4",
				Policy: json.RawMessage(`{
					"other_field": "value"
				}`),
			},
			mockTime:     baseTime,
			wantErr:      true,
			errorMessage: "failed to parse interval",
		},
		{
			name: "invalid interval",
			policy: types.PluginPolicy{
				ID: "policy5",
				Policy: json.RawMessage(`{
					"schedule": {
						"frequency": "daily",
						"start_time": "2023-01-01T10:00:00Z",
						"interval": "not_a_number"
					}
				}`),
			},
			mockTime:     baseTime,
			wantErr:      true,
			errorMessage: "failed to parse interval",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockClock := new(MockClock)

			if !tc.wantErr {
				mockClock.On("Now").Return(tc.mockTime)
			}

			scheduler := SchedulerService{
				db:         nil,
				logger:     logrus.StandardLogger(),
				clock:      mockClock,
				taskClient: nil,
				inspector:  nil,
			}

			trigger, err := scheduler.GetTriggerFromPolicy(tc.policy)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errorMessage != "" {
					require.Contains(t, err.Error(), tc.errorMessage)
				}
				require.Nil(t, trigger)
			} else {
				require.NoError(t, err)
				require.NotNil(t, trigger)
				if tc.validateTrigger != nil {
					tc.validateTrigger(t, trigger)
				}
			}

			mockClock.AssertExpectations(t)
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
