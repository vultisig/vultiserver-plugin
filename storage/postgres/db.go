package postgres

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
)

//go:embed migrations/*
var embeddedMigrations embed.FS

type PostgresBackend struct {
	pool *pgxpool.Pool
}

func NewPostgresBackend(readonly bool, dsn string) (*PostgresBackend, error) {
	logrus.Info("Connecting to database with DSN: ", dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	backend := &PostgresBackend{
		pool: pool,
	}

	if err := backend.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return backend, nil
}

func (d *PostgresBackend) Close() error {
	d.pool.Close()

	return nil
}

func (d *PostgresBackend) Migrate() error {
	logrus.Info("Starting database migration...")
	goose.SetBaseFS(embeddedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	db := stdlib.OpenDBFromPool(d.pool)
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to run goose up: %w", err)
	}
	logrus.Info("Database migration completed successfully")
	return nil
}

func (p *PostgresBackend) CreateTimeTrigger(trigger TimeTrigger) error {
	logrus.Info("Creating time trigger in database")
	if p.pool == nil {
		return fmt.Errorf("database pool is nil")
	}

	query := `
        INSERT INTO time_triggers 
        (policy_id, cron_expression, start_time, end_time, frequency) 
        VALUES ($1, $2, $3, $4, $5)`

	_, err := p.pool.Exec(context.Background(), query,
		trigger.PolicyID,
		trigger.CronExpression,
		trigger.StartTime,
		trigger.EndTime,
		trigger.Frequency)

	return err
}

func (p *PostgresBackend) GetPendingTriggers() ([]TimeTrigger, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("database pool is nil")
	}

	query := `
        SELECT policy_id, cron_expression, start_time, end_time, frequency, last_execution 
        FROM time_triggers 
        WHERE start_time <= NOW() 
        AND (end_time IS NULL OR end_time > NOW())`

	rows, err := p.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var triggers []TimeTrigger
	for rows.Next() {
		var t TimeTrigger
		err := rows.Scan(
			&t.PolicyID,
			&t.CronExpression,
			&t.StartTime,
			&t.EndTime,
			&t.Frequency,
			&t.LastExecution)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, t)
	}

	return triggers, nil
}

func (p *PostgresBackend) UpdateTriggerExecution(policyID string) error {
	if p.pool == nil {
		return fmt.Errorf("database pool is nil")
	}

	query := `
        UPDATE time_triggers 
        SET last_execution = NOW() 
        WHERE policy_id = $1`

	_, err := p.pool.Exec(context.Background(), query, policyID)
	return err
}

type TimeTrigger struct {
	PolicyID       string
	CronExpression string
	StartTime      time.Time
	EndTime        *time.Time
	Frequency      string
	LastExecution  *time.Time
}
