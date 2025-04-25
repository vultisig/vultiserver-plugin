package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/vultisig/vultiserver-plugin/internal/types"
	"github.com/vultisig/vultiserver-plugin/storage"
)

var _ storage.PolicyRepository = (*PostgresBackend)(nil)

func (p *PostgresBackend) GetPluginPolicy(ctx context.Context, id string) (types.PluginPolicy, error) {
	if p.pool == nil {
		return types.PluginPolicy{}, fmt.Errorf("database pool is nil")
	}

	var policy types.PluginPolicy
	var policyJSON []byte

	query := `
        SELECT id, public_key_ecdsa, public_key_eddsa, plugin_version, policy_version, plugin_type, is_ecdsa, chain_code_hex, derive_path, active, progress, signature, policy
        FROM plugin_policies
        WHERE id = $1`

	err := p.pool.QueryRow(ctx, query, id).Scan(
		&policy.ID,
		&policy.PublicKeyEcdsa,
		&policy.PublicKeyEddsa,
		&policy.PluginVersion,
		&policy.PolicyVersion,
		&policy.PluginType,
		&policy.IsEcdsa,
		&policy.ChainCodeHex,
		&policy.DerivePath,
		&policy.Active,
		&policy.Progress,
		&policy.Signature,
		&policyJSON,
	)

	if err != nil {
		return types.PluginPolicy{}, fmt.Errorf("failed to get policy: %w", err)
	}
	policy.Policy = json.RawMessage(policyJSON)

	return policy, nil
}

func (p *PostgresBackend) GetAllPluginPolicies(ctx context.Context, publicKeyEcdsa string, pluginType string) ([]types.PluginPolicy, error) {
	if p.pool == nil {
		return []types.PluginPolicy{}, fmt.Errorf("database pool is nil")
	}

	query := `
  	SELECT id, public_key_ecdsa, public_key_eddsa, plugin_version, policy_version, plugin_type, is_ecdsa, chain_code_hex, derive_path, active, progress, signature, policy
		FROM plugin_policies
		WHERE public_key_ecdsa = $1
		AND plugin_type = $2`

	rows, err := p.pool.Query(ctx, query, publicKeyEcdsa, pluginType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var policies []types.PluginPolicy
	for rows.Next() {
		var policy types.PluginPolicy
		err := rows.Scan(
			&policy.ID,
			&policy.PublicKeyEcdsa,
			&policy.PublicKeyEddsa,
			&policy.PluginVersion,
			&policy.PolicyVersion,
			&policy.PluginType,
			&policy.IsEcdsa,
			&policy.ChainCodeHex,
			&policy.DerivePath,
			&policy.Active,
			&policy.Progress,
			&policy.Signature,
			&policy.Policy,
		)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

func (p *PostgresBackend) InsertPluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error) {
	policyJSON, err := json.Marshal(policy.Policy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal policy: %w", err)
	}

	query := `
  	INSERT INTO plugin_policies (
      id, public_key_ecdsa, public_key_eddsa, plugin_version, policy_version, plugin_type, is_ecdsa, chain_code_hex, derive_path, active, progress, signature, policy
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
    RETURNING id, public_key_ecdsa, public_key_eddsa, plugin_version, policy_version, plugin_type, is_ecdsa, chain_code_hex, derive_path, active, progress, signature, policy
	`

	var insertedPolicy types.PluginPolicy
	err = dbTx.QueryRow(ctx, query,
		&policy.ID,
		&policy.PublicKeyEcdsa,
		&policy.PublicKeyEddsa,
		&policy.PluginVersion,
		&policy.PolicyVersion,
		&policy.PluginType,
		&policy.IsEcdsa,
		&policy.ChainCodeHex,
		&policy.DerivePath,
		&policy.Active,
		&policy.Progress,
		&policy.Signature,
		policyJSON,
	).Scan(
		&insertedPolicy.ID,
		&insertedPolicy.PublicKeyEcdsa,
		&insertedPolicy.PublicKeyEddsa,
		&insertedPolicy.PluginVersion,
		&insertedPolicy.PolicyVersion,
		&insertedPolicy.PluginType,
		&insertedPolicy.IsEcdsa,
		&insertedPolicy.ChainCodeHex,
		&insertedPolicy.DerivePath,
		&insertedPolicy.Active,
		&insertedPolicy.Progress,
		&insertedPolicy.Signature,
		&insertedPolicy.Policy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert policy: %w", err)
	}

	return &insertedPolicy, nil
}

func (p *PostgresBackend) UpdatePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error) {
	policyJSON, err := json.Marshal(policy.Policy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal policy: %w", err)
	}

	// TODO: update other fields
	setClauses := []string{
		"public_key_ecdsa = $2",
		"public_key_eddsa = $3",
		"plugin_type = $4",
		"signature = $5",
		"active = $6",
		"policy = $7",
	}
	args := []interface{}{
		policy.ID,
		policy.PublicKeyEcdsa,
		policy.PublicKeyEddsa,
		policy.PluginType,
		policy.Signature,
		policy.Active,
		policyJSON,
	}
	returningFields := "id, public_key, plugin_id, plugin_version, policy_version, plugin_type, signature, active, progress, policy"

	if policy.Progress != "" {
		setClauses = append(setClauses, fmt.Sprintf("progress = $%d", len(args)+1))
		args = append(args, policy.Progress)
	}

	query := fmt.Sprintf(`
		UPDATE plugin_policies
		SET %s
		WHERE id = $1
		RETURNING %s
	`, strings.Join(setClauses, ", "), returningFields)

	var updatedPolicy types.PluginPolicy

	dest := []interface{}{
		&updatedPolicy.ID,
		&updatedPolicy.PublicKeyEcdsa,
		&updatedPolicy.PublicKeyEddsa,
		&updatedPolicy.PluginVersion,
		&updatedPolicy.PolicyVersion,
		&updatedPolicy.PluginType,
		&updatedPolicy.IsEcdsa,
		&updatedPolicy.ChainCodeHex,
		&updatedPolicy.DerivePath,
		&updatedPolicy.Active,
		&updatedPolicy.Signature,
		&updatedPolicy.Progress,
		&updatedPolicy.Policy,
	}

	err = dbTx.QueryRow(ctx, query, args...).Scan(dest...)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("policy not found with ID: %s", policy.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	return &updatedPolicy, nil
}

func (p *PostgresBackend) DeletePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, id string) error {
	_, err := dbTx.Exec(ctx, `
	DELETE FROM transaction_history
	WHERE policy_id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete transaction history: %w", err)
	}
	_, err = dbTx.Exec(ctx, `
	DELETE FROM time_triggers
	WHERE policy_id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete time triggers: %w", err)
	}
	_, err = dbTx.Exec(ctx, `
	DELETE FROM plugin_policies
	WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	return nil
}
