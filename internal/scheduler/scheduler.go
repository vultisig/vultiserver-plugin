package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultiserver-plugin/internal/tasks"
	"github.com/vultisig/vultiserver-plugin/internal/types"
	"github.com/vultisig/vultiserver-plugin/storage"
)

const (
	secondsInDay  = 24 * 60 * 60
	secondsInWeek = 7 * 24 * 60 * 60
)

var (
	ErrTriggerNotReady = errors.New("trigger is not ready")
	ErrEndTimeReached  = errors.New("trigger end time reached")
)

type Clock interface {
	Now() time.Time
	NewTicker(d time.Duration) *time.Ticker
}

type RealClock struct{}

func (c *RealClock) Now() time.Time {
	return time.Now().UTC()
}
func (c *RealClock) NewTicker(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}

// Task client abstracts asynq client operations
type TaskClient interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type AsynqTaskClient struct {
	client *asynq.Client
}

func (a *AsynqTaskClient) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return a.client.Enqueue(task, opts...)
}

type SchedulerService struct {
	db         storage.TimeTriggerRepository
	logger     *logrus.Logger
	taskClient TaskClient
	inspector  *asynq.Inspector
	clock      Clock
	done       chan struct{}
}

func NewSchedulerService(db storage.DatabaseStorage, logger *logrus.Logger, client *asynq.Client, redisOpts asynq.RedisClientOpt) *SchedulerService {
	if db == nil {
		logger.Fatal("database connection is nil")
	}

	// create inspector using the same Redis configuration as the client
	inspector := asynq.NewInspector(redisOpts)

	return &SchedulerService{
		db:         db,
		logger:     logger,
		taskClient: &AsynqTaskClient{client},
		clock:      &RealClock{},
		inspector:  inspector,
		done:       make(chan struct{}),
	}
}

func (s *SchedulerService) Start() {
	go s.run()
}

func (s *SchedulerService) Stop() {
	close(s.done)
}

func (s *SchedulerService) run() {
	ticker := s.clock.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.checkAndEnqueueTasks(); err != nil {
				s.logger.Errorf("Failed to check and enqueue tasks: %v", err)
			}
		case <-s.done:
			return
		}
	}
}

func (s *SchedulerService) checkAndEnqueueTasks() error {
	ctx := context.Background()
	triggers, err := s.db.GetPendingTimeTriggers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending triggers: %w", err)
	}
	s.logger.Infof("Found %d active triggers: %+v: ", len(triggers), triggers)

	for _, trigger := range triggers {
		if err := s.processTrigger(ctx, trigger); err != nil {
			if !errors.Is(err, ErrTriggerNotReady) {
				s.logger.Errorf("Failed to process trigger: %v", err)
			}
			continue
		}
	}

	return nil
}

func (s *SchedulerService) processTrigger(ctx context.Context, trigger types.TimeTrigger) error {
	s.logger.WithFields(logrus.Fields{
		"policy_id": trigger.PolicyID,
		"last_exec": trigger.LastExecution,
	}).Info("Processing trigger")

	// Parse cron expression
	schedule, err := CreateSchedule(trigger.CronExpression, trigger.Frequency, trigger.StartTime, trigger.Interval)
	if err != nil {
		s.logger.Errorf("Failed to create schedule: %v", err)
		err := s.db.DeleteTimeTrigger(ctx, trigger.PolicyID)
		if err != nil {
			return fmt.Errorf("failed to delete time trigger: %w", err)
		}
		return fmt.Errorf("invalid schedule: %w", err)
	}

	// Check if the end time has been readched
	if endTime := trigger.EndTime; endTime != nil && s.clock.Now().After(*endTime) {
		// TODO: Check if this end time was ever set anywhere.
		s.logger.WithFields(logrus.Fields{
			"policy_id": trigger.PolicyID,
			"end_time":  *endTime,
		}).Info("Trigger end time reached")
		err := s.db.DeleteTimeTrigger(ctx, trigger.PolicyID)
		if err != nil {
			return fmt.Errorf("failed to delete time trigger: %w", err)
		}
		return ErrEndTimeReached //TODO: Think if we need an error to say about end time has been reached.
	}

	// Check if it's time to execute
	var nextTime time.Time
	if trigger.LastExecution != nil {
		nextTime = schedule.Next(*trigger.LastExecution)
	} else {
		nextTime = trigger.StartTime
	}

	nextTime = nextTime.UTC()

	triggerStatus, err := s.db.GetTriggerStatus(ctx, trigger.PolicyID)
	if err != nil {
		return fmt.Errorf("failed to get trigger status: %w", err)
	}

	if s.clock.Now().Before(nextTime) || triggerStatus == types.StatusTimeTriggerRunning {
		s.logger.WithFields(logrus.Fields{
			"policy_id": trigger.PolicyID,
			"next_time": nextTime,
			"state":     triggerStatus,
		}).Info("Trigger have not reached next time or it's in running status")
		return ErrTriggerNotReady // TODO: Think if we need a special error in this case to say that the trigger is not ready to execute.
	}

	return s.enqueueTriggerTask(ctx, trigger)
}

func (s *SchedulerService) enqueueTriggerTask(ctx context.Context, trigger types.TimeTrigger) error {
	if err := s.db.UpdateTriggerStatus(ctx, trigger.PolicyID, types.StatusTimeTriggerRunning); err != nil {
		return fmt.Errorf("failed to update trigger status: %w", err)
	}

	buf, err := json.Marshal(trigger)
	if err != nil {
		return fmt.Errorf("failed to marshal trigger event: %w", err)
	}

	ti, err := s.taskClient.Enqueue(
		asynq.NewTask(tasks.TypePluginTransaction, buf),
		asynq.MaxRetry(0),
		asynq.Timeout(5*time.Minute),
		asynq.Retention(10*time.Minute),
		asynq.Queue(tasks.QUEUE_NAME),
	)

	if err != nil {
		return fmt.Errorf("failed to enqueue trigger task: %w", err)

	}

	s.logger.WithFields(logrus.Fields{
		"task_id":   ti.ID,
		"policy_id": trigger.PolicyID,
	}).Info("Enqueued trigger task")
	return nil
}

func (s *SchedulerService) CreateTimeTrigger(ctx context.Context, policy types.PluginPolicy, dbTx pgx.Tx) error {
	if s.db == nil {
		return fmt.Errorf("database backend is nil")
	}

	trigger, err := s.GetTriggerFromPolicy(policy)
	if err != nil {
		return fmt.Errorf("failed to get trigger from policy: %w", err)
	}

	return s.db.CreateTimeTriggerTx(ctx, dbTx, *trigger)
}

func (s *SchedulerService) GetTriggerFromPolicy(policy types.PluginPolicy) (*types.TimeTrigger, error) {
	var policySchedule struct {
		Schedule struct {
			Frequency string     `json:"frequency"`
			StartTime time.Time  `json:"start_time"`
			Interval  string     `json:"interval"`
			EndTime   *time.Time `json:"end_time,omitempty"`
		} `json:"schedule"`
	}

	if err := json.Unmarshal(policy.Policy, &policySchedule); err != nil {
		return nil, fmt.Errorf("failed to parse policy schedule: %w", err)
	}

	interval, err := strconv.Atoi(policySchedule.Schedule.Interval)
	if err != nil {
		return nil, fmt.Errorf("failed to parse interval: %w", err)
	}

	cronExpr := FrequencyToCron(policySchedule.Schedule.Frequency, policySchedule.Schedule.StartTime, interval)
	trigger := types.TimeTrigger{
		PolicyID:       policy.ID,
		CronExpression: cronExpr,
		StartTime:      s.clock.Now(),
		EndTime:        policySchedule.Schedule.EndTime,
		Frequency:      policySchedule.Schedule.Frequency,
		Interval:       interval,
		Status:         types.StatusTimeTriggerPending,
	}

	return &trigger, nil
}

func CreateSchedule(cronExpr, frequency string, startTime time.Time, interval int) (cron.Schedule, error) {
	// Use our custom schedule implementation for intervals > 1 and when frequency is daily, weekly, monthly
	if interval > 1 && (frequency == "daily" || frequency == "weekly" || frequency == "monthly") {
		return NewIntervalSchedule(frequency, startTime, interval)
	}

	// For standard cron
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cron expression: %w", err)
	}

	return schedule, nil
}

func FrequencyToCron(frequency string, startTime time.Time, interval int) string {
	switch frequency {
	case "minutely":
		return fmt.Sprintf("*/%d * * * *", interval)
	case "hourly":
		if interval == 1 {
			return fmt.Sprintf("%d * * * *", startTime.Minute())
		}
		return fmt.Sprintf("%d */%d * * *", startTime.Minute(), interval)
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

type IntervalSchedule struct {
	Frequency string
	Interval  int
	StartTime time.Time
	Minute    int
	Hour      int
	Day       int
	Weekday   time.Weekday
	Location  *time.Location
}

func NewIntervalSchedule(frequency string, startTime time.Time, interval int) (*IntervalSchedule, error) {
	if interval < 1 {
		return nil, fmt.Errorf("interval must be at least 1")
	}

	return &IntervalSchedule{
		Frequency: frequency,
		Interval:  interval,
		StartTime: startTime,
		Minute:    startTime.Minute(),
		Hour:      startTime.Hour(),
		Day:       startTime.Day(),
		Weekday:   startTime.Weekday(),
		Location:  startTime.Location(),
	}, nil
}

func (s *IntervalSchedule) Next(t time.Time) time.Time {
	t = t.In(s.Location)

	switch s.Frequency {
	case "daily":
		return s.nextDaily(t)
	case "weekly":
		return s.nextWeekly(t)
	case "monthly":
		return s.nextMonthly(t)
	default:
		return time.Time{}
	}
}

func (s *IntervalSchedule) nextDaily(t time.Time) time.Time {
	// Create candidate time with the correct hour and minute on the current day
	candidate := time.Date(t.Year(), t.Month(), t.Day(), s.Hour, s.Minute, 0, 0, s.Location)

	// If the candidate is in the past, move to the next day
	if !candidate.After(t) {
		candidate = candidate.AddDate(0, 0, 1)
	}

	// Calculate the absolute number of days from the epoch for both start time and candidate
	// This ensures proper alignment regardless of month boundaries
	startDays := int(s.StartTime.Unix() / secondsInDay)
	candidateDays := int(candidate.Unix() / secondsInDay)

	// Calculate how many days past the start time
	daysPastStart := candidateDays - startDays

	// If we're already on a valid day, return the candidate
	if daysPastStart >= 0 && daysPastStart%s.Interval == 0 {
		return candidate
	}

	// Otherwise, calculate days to add to reach the next valid day
	daysToAdd := s.Interval - (daysPastStart % s.Interval)
	if daysPastStart < 0 {
		// Special handling if we're before the start time
		daysToAdd = -daysPastStart
	}

	return candidate.AddDate(0, 0, daysToAdd)
}

// nextWeekly calculates the next execution for weekly intervals > 1
func (s *IntervalSchedule) nextWeekly(t time.Time) time.Time {
	// First find the next occurrence of the correct weekday
	daysUntilWeekday := int(s.Weekday - t.Weekday())
	if daysUntilWeekday <= 0 {
		daysUntilWeekday += 7
	}

	// Create the candidate time with the correct weekday, hour, and minute
	candidate := time.Date(
		t.Year(), t.Month(), t.Day()+daysUntilWeekday,
		s.Hour, s.Minute, 0, 0, s.Location,
	)

	// If the candidate is in the past, move to the next week
	if !candidate.After(t) {
		candidate = candidate.AddDate(0, 0, 7)
	}

	// Calculate absolute number of weeks from epoch for proper alignment
	// Using Monday as the start of the week for consistent calculations
	startWeeks := int(timeToMondayMidnight(s.StartTime).Unix() / secondsInWeek)
	candidateWeeks := int(timeToMondayMidnight(candidate).Unix() / secondsInWeek)

	// Calculate how many weeks past the start time
	weeksPastStart := candidateWeeks - startWeeks

	// If we're already on a valid week, return the candidate
	if weeksPastStart >= 0 && weeksPastStart%s.Interval == 0 {
		return candidate
	}

	// Otherwise, calculate weeks to add to reach the next valid week
	weeksToAdd := s.Interval - (weeksPastStart % s.Interval)
	if weeksPastStart < 0 {
		// Special handling if we're before the start time
		weeksToAdd = -weeksPastStart
	}

	return candidate.AddDate(0, 0, 7*weeksToAdd)
}

func (s *IntervalSchedule) nextMonthly(t time.Time) time.Time {
	// Always start from at least the schedule's start time
	if t.Before(s.StartTime) {
		t = s.StartTime
	}

	// Calculate total months since the epoch (or any fixed reference point)
	startMonths := s.StartTime.Year()*12 + int(s.StartTime.Month()) - 1
	currentMonths := t.Year()*12 + int(t.Month()) - 1

	// Calculate how many intervals have passed since start
	intervalsPassed := (currentMonths - startMonths) / s.Interval

	// Calculate the last interval month
	lastIntervalMonth := startMonths + intervalsPassed*s.Interval

	// Calculate the next interval month
	nextIntervalMonth := lastIntervalMonth

	// If we're already past the day/time in the current interval month,
	// or if we're exactly at the current interval month but before the start date,
	// move to the next interval
	if currentMonths > lastIntervalMonth ||
		(currentMonths == lastIntervalMonth &&
			(t.Day() > s.Day || (t.Day() == s.Day && (t.Hour() > s.Hour || (t.Hour() == s.Hour && t.Minute() >= s.Minute))))) {
		nextIntervalMonth = lastIntervalMonth + s.Interval
	}

	// Convert back to year and month
	nextYear := nextIntervalMonth / 12
	nextMonth := time.Month(nextIntervalMonth%12 + 1)

	// Create the candidate time
	candidate := time.Date(nextYear, nextMonth, s.Day, s.Hour, s.Minute, 0, 0, s.Location)

	// Handle months with fewer days than our target day
	if candidate.Day() != s.Day {
		// We got bumped to the next month due to day overflow, go back to last day of previous month
		candidate = time.Date(nextYear, nextMonth, 0, s.Hour, s.Minute, 0, 0, s.Location)
	}

	return candidate
}

func timeToMondayMidnight(t time.Time) time.Time {
	daysFromMonday := int(t.Weekday())
	if daysFromMonday == 0 { // Sunday
		daysFromMonday = 6
	} else {
		daysFromMonday--
	}

	return time.Date(t.Year(), t.Month(), t.Day()-daysFromMonday, 0, 0, 0, 0, t.Location())
}
