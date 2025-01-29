package scheduler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
)

type SchedulerService struct {
	db        storage.DatabaseStorage
	logger    *logrus.Logger
	client    *asynq.Client
	inspector *asynq.Inspector
	done      chan struct{}
}

func NewSchedulerService(db storage.DatabaseStorage, logger *logrus.Logger, client *asynq.Client, redisOpts asynq.RedisClientOpt) *SchedulerService {
	if db == nil {
		logger.Fatal("database connection is nil")
	}

	// create inspector using the same Redis configuration as the client
	inspector := asynq.NewInspector(redisOpts)

	return &SchedulerService{
		db:        db,
		logger:    logger,
		client:    client,
		inspector: inspector,
		done:      make(chan struct{}),
	}
}

func (s *SchedulerService) Start() {
	go s.run()
}

func (s *SchedulerService) Stop() {
	close(s.done)
}

func (s *SchedulerService) run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.logger.Info("Checking and enqueuing tasks")
			if err := s.checkAndEnqueueTasks(); err != nil {
				s.logger.Errorf("Failed to check and enqueue tasks: %v", err)
			}
		case <-s.done:
			return
		}
	}
}

func (s *SchedulerService) checkAndEnqueueTasks() error {
	triggers, err := s.db.GetPendingTriggers()
	if err != nil {
		return fmt.Errorf("failed to get pending triggers: %w", err)
	}

	for _, trigger := range triggers {
		// Parse cron expression
		schedule, err := cron.ParseStandard(trigger.CronExpression)
		if err != nil {
			s.logger.Errorf("Failed to parse cron expression: %v", err)
			s.db.DeleteTimeTrigger(trigger.PolicyID) //trigger is deleted if cron expression is invalid
			continue
		}

		// Check if it's time to execute
		var nextTime time.Time
		if trigger.LastExecution != nil {

			nextTime = schedule.Next(trigger.LastExecution.In(time.UTC))

			s.logger.WithFields(logrus.Fields{
				"current_time_utc": time.Now().UTC(),
				"next_time_utc":    nextTime.UTC(),
				"delay_duration":   nextTime.UTC().Sub(time.Now().UTC()),
				"status":           trigger.Status,
			}).Info("Next execution details")
		} else {
			nextTime = time.Now().UTC().Add(-1 * time.Minute)
			s.logger.WithFields(logrus.Fields{
				"current_time": time.Now(),
				"next_time":    nextTime,
			}).Info("New trigger details")
		}

		nextTime = nextTime.UTC()
		endTime := trigger.EndTime

		if endTime != nil && time.Now().UTC().After(*endTime) {
			s.logger.WithFields(logrus.Fields{
				"policy_id": trigger.PolicyID,
				"end_time":  *endTime,
			}).Info("Trigger end time reached")
			s.db.DeleteTimeTrigger(trigger.PolicyID)
			continue
		}

		triggerStatus, err := s.db.GetTriggerStatus(trigger.PolicyID)
		if err != nil {
			s.logger.Errorf("Failed to get trigger status: %v", err)
			continue
		}

		if time.Now().UTC().Before(nextTime) || triggerStatus == "Running" {
			continue
		}

		s.db.UpdateTriggerStatus(trigger.PolicyID, "Running")

		buf, err := json.Marshal(trigger)
		if err != nil {
			s.logger.Errorf("Failed to marshal trigger event: %v", err)
			continue
		}
		ti, err := s.client.Enqueue(
			asynq.NewTask(tasks.TypePluginTransaction, buf),
			asynq.MaxRetry(0),
			asynq.Timeout(5*time.Minute),
			asynq.Retention(10*time.Minute),
			asynq.Queue(tasks.QUEUE_NAME),
		)
		if err != nil {
			s.logger.Errorf("Failed to enqueue trigger task: %v", err)
			continue
		}

		s.logger.WithFields(logrus.Fields{
			"task_id":   ti.ID,
			"policy_id": trigger.PolicyID,
		}).Info("Enqueued trigger task")
	}

	return nil
}

func (s *SchedulerService) CreateTimeTrigger(policy types.PluginPolicy) error {
	if s.db == nil {
		return fmt.Errorf("database backend is nil")
	}

	s.logger.Info("Attempting to parse policy schedule")

	var policySchedule struct {
		Schedule struct {
			Frequency string     `json:"frequency"`
			StartTime time.Time  `json:"start_time"`
			EndTime   *time.Time `json:"end_time,omitempty"`
		} `json:"schedule"`
	}

	if err := json.Unmarshal(policy.Policy, &policySchedule); err != nil {
		return fmt.Errorf("failed to parse policy schedule: %w", err)
	}

	s.logger.Info("Frequency to cron")

	cronExpr := frequencyToCron(policySchedule.Schedule.Frequency, policySchedule.Schedule.StartTime)

	trigger := types.TimeTrigger{
		PolicyID:       policy.ID,
		CronExpression: cronExpr,
		StartTime:      policySchedule.Schedule.StartTime,
		EndTime:        policySchedule.Schedule.EndTime,
		Frequency:      policySchedule.Schedule.Frequency,
	}

	return s.db.CreateTimeTrigger(trigger)
}

func frequencyToCron(frequency string, startTime time.Time) string {
	switch frequency {
	case "5-minutely":
		return fmt.Sprintf("%d/5 * * * *", startTime.Minute()%5)
	case "hourly":
		return fmt.Sprintf("%d * * * *", startTime.Minute())
	case "daily":
		return fmt.Sprintf("%d %d * * *", startTime.Minute(), startTime.Hour())
	case "weekly":
		return fmt.Sprintf("%d %d * * %d", startTime.Minute(), startTime.Hour(), startTime.Weekday())
	case "monthly":
		return fmt.Sprintf("%d %d %d * *", startTime.Minute(), startTime.Hour(), startTime.Day())
	default:
		return ""
	}
}
