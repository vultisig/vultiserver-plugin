package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/internal/types"
)

func (p *PostgresBackend) CreateTimeTrigger(trigger types.TimeTrigger) error {
	logrus.Info("Creating time trigger in database")
	if p.pool == nil {
		return fmt.Errorf("database pool is nil")
	}

	query := `
        INSERT INTO time_triggers 
        (policy_id, cron_expression, start_time, end_time, frequency, status) 
        VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := p.pool.Exec(context.Background(), query,
		trigger.PolicyID,
		trigger.CronExpression,
		trigger.StartTime,
		trigger.EndTime,
		trigger.Frequency,
		trigger.Status)

	return err
}

func (p *PostgresBackend) GetPendingTriggers() ([]types.TimeTrigger, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("database pool is nil")
	}

	query := `
        SELECT policy_id, cron_expression, start_time, end_time, frequency, last_execution, status 
        FROM time_triggers 
        WHERE start_time <= NOW() 
        AND (end_time IS NULL OR end_time > NOW())`

	rows, err := p.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var triggers []types.TimeTrigger
	for rows.Next() {
		var t types.TimeTrigger
		err := rows.Scan(
			&t.PolicyID,
			&t.CronExpression,
			&t.StartTime,
			&t.EndTime,
			&t.Frequency,
			&t.LastExecution,
			&t.Status,
		)
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
        SET last_execution = $2
        WHERE policy_id = $1`

	_, err := p.pool.Exec(context.Background(), query, policyID, time.Now().UTC())
	return err
}

func (p *PostgresBackend) UpdateTriggerStatus(policyID string, status string) error {
	if p.pool == nil {
		return fmt.Errorf("database pool is nil")
	}

	query := `
        UPDATE time_triggers 
        SET status = $2
        WHERE policy_id = $1`

	_, err := p.pool.Exec(context.Background(), query, policyID, status)
	return err
}

func (p *PostgresBackend) GetTriggerStatus(policyID string) (string, error) {
	if p.pool == nil {
		return "", fmt.Errorf("database pool is nil")
	}

	query := `
        SELECT status 
        FROM time_triggers 
        WHERE policy_id = $1`

	var status string
	err := p.pool.QueryRow(context.Background(), query, policyID).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("trigger not found for policy_id: %s", policyID)
		}
		return "", err
	}

	return status, nil
}
