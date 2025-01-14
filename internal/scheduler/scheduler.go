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

func NewSchedulerService(db storage.DatabaseStorage, logger *logrus.Logger, client *asynq.Client) *SchedulerService {
	if db == nil {
		logger.Fatal("database connection is nil")
	}

	// create inspector using the same Redis configuration as the client
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr: "localhost:6379", // redis configuration
		DB:   1,                // redis DB number
	})

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
	ticker := time.NewTicker(5 * time.Second)
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
			s.logger.Errorf("Failed to parse cron expression: %v", err) //todo : remove trigger if cron expression is invalid
			continue
		}

		// Check if it's time to execute
		var nextTime time.Time
		if trigger.LastExecution != nil {
			s.logger.Info("Last execution is not nil, using last execution")
			s.logger.WithFields(logrus.Fields{
				"last execution": trigger.LastExecution,
			}).Info("Last execution")
			nextTime = schedule.Next(*trigger.LastExecution)
		} else {
			s.logger.Info("Last execution is nil, using default window")
			nextTime = schedule.Next(time.Now().Add(-24 * time.Hour)) //todo : change window to a better duration
		}

		s.logger.WithFields(logrus.Fields{
			"next time": nextTime,
		}).Info("Next time")

		s.logger.WithFields(logrus.Fields{
			"time": time.Now().UTC(),
		}).Info("Current time")

		s.logger.WithFields(logrus.Fields{
			"next time greater than current time": time.Now().UTC().After(nextTime),
		}).Info("Next time greater than current time")

		nextTime = nextTime.UTC()

		s.logger.WithFields(logrus.Fields{
			"next time": nextTime,
		}).Info("Next time")

		s.logger.WithFields(logrus.Fields{
			"time": time.Now().UTC(),
		}).Info("Current time")

		s.logger.WithFields(logrus.Fields{
			"next time greater than current time": time.Now().UTC().After(nextTime),
		}).Info("Next time greater than current time")

		/*s.logger.WithFields(logrus.Fields{
			"policy_id":    trigger.PolicyID,
			"last_exec":    trigger.LastExecution,
			"current_time": time.Now().UTC(),
			"next_time":    nextTime,
		}).Info("Processing trigger")*/

		if time.Now().UTC().After(nextTime) {
			s.logger.WithFields(logrus.Fields{
				"policy_id":    trigger.PolicyID,
				"last_exec":    trigger.LastExecution,
				"current_time": time.Now().UTC(),
				"next_time":    nextTime,
			}).Info("Inside if statement")

			triggerEvent := types.PluginTriggerEvent{
				PolicyID: trigger.PolicyID,
			}

			buf, err := json.Marshal(triggerEvent)
			if err != nil {
				s.logger.Errorf("Failed to marshal trigger event: %v", err)
				continue
			}
			ti, err := s.client.Enqueue(
				asynq.NewTask(tasks.TypePluginTransaction, buf),
				asynq.MaxRetry(-1),
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
		return "*/5 * * * *"
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
