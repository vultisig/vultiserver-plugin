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
        SELECT id, public_key, is_ecdsa, chain_code_hex, derive_path, plugin_id, plugin_version, policy_version, plugin_type, signature, active, policy, progress
        FROM plugin_policies 
        WHERE id = $1`

	err := p.pool.QueryRow(ctx, query, id).Scan(
		&policy.ID,
		&policy.PublicKey,
		&policy.IsEcdsa,
		&policy.ChainCodeHex,
		&policy.DerivePath,
		&policy.PluginID,
		&policy.PluginVersion,
		&policy.PolicyVersion,
		&policy.PluginType,
		&policy.Signature,
		&policy.Active,
		&policyJSON,
		&policy.Progress,
	)

	if err != nil {
		return types.PluginPolicy{}, fmt.Errorf("failed to get policy: %w", err)
	}
	policy.Policy = json.RawMessage(policyJSON)

	return policy, nil
}

func (p *PostgresBackend) GetAllPluginPolicies(ctx context.Context, publicKey string, pluginType string, take int, skip int) (types.PluginPolicyPaginatedList, error) {
	if p.pool == nil {
		return types.PluginPolicyPaginatedList{}, fmt.Errorf("database pool is nil")
	}

	query := `
  	SELECT id, public_key, is_ecdsa, chain_code_hex, derive_path, plugin_id, plugin_version, policy_version, plugin_type, signature, active, policy, progress, COUNT(*) OVER() AS total_count
		FROM plugin_policies
		WHERE public_key = $1
		AND plugin_type = $2
		LIMIT $3 OFFSET $4`

	rows, err := p.pool.Query(ctx, query, publicKey, pluginType, take, skip)
	if err != nil {
		return types.PluginPolicyPaginatedList{}, err
	}
	defer rows.Close()
	var policies []types.PluginPolicy
	var totalCount int
	for rows.Next() {
		var policy types.PluginPolicy
		err := rows.Scan(
			&policy.ID,
			&policy.PublicKey,
			&policy.IsEcdsa,
			&policy.ChainCodeHex,
			&policy.DerivePath,
			&policy.PluginID,
			&policy.PluginVersion,
			&policy.PolicyVersion,
			&policy.PluginType,
			&policy.Signature,
			&policy.Active,
			&policy.Policy,
			&policy.Progress,
			&totalCount,
		)
		if err != nil {
			return types.PluginPolicyPaginatedList{}, err
		}
		policies = append(policies, policy)
	}

	dto := types.PluginPolicyPaginatedList{
		Policies:   policies,
		TotalCount: totalCount,
	}

	return dto, nil
}

func (p *PostgresBackend) InsertPluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error) {
	policyJSON, err := json.Marshal(policy.Policy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal policy: %w", err)
	}

	query := `
  	INSERT INTO plugin_policies (
      id, public_key, is_ecdsa, chain_code_hex, derive_path, plugin_id, plugin_version, policy_version, plugin_type, signature, active, policy
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    RETURNING id, public_key, is_ecdsa, chain_code_hex, derive_path, plugin_id, plugin_version, policy_version, plugin_type, signature, active, policy, progress
	`

	var insertedPolicy types.PluginPolicy
	err = dbTx.QueryRow(ctx, query,
		policy.ID,
		policy.PublicKey,
		policy.IsEcdsa,
		policy.ChainCodeHex,
		policy.DerivePath,
		policy.PluginID,
		policy.PluginVersion,
		policy.PolicyVersion,
		policy.PluginType,
		policy.Signature,
		policy.Active,
		policyJSON,
	).Scan(
		&insertedPolicy.ID,
		&insertedPolicy.PublicKey,
		&insertedPolicy.IsEcdsa,
		&insertedPolicy.ChainCodeHex,
		&insertedPolicy.DerivePath,
		&insertedPolicy.PluginID,
		&insertedPolicy.PluginVersion,
		&insertedPolicy.PolicyVersion,
		&insertedPolicy.PluginType,
		&insertedPolicy.Signature,
		&insertedPolicy.Active,
		&insertedPolicy.Policy,
		&insertedPolicy.Progress,
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
		"public_key = $2",
		"plugin_type = $3",
		"signature = $4",
		"active = $5",
		"policy = $6",
	}
	args := []interface{}{
		policy.ID,
		policy.PublicKey,
		policy.PluginType,
		policy.Signature,
		policy.Active,
		policyJSON,
	}
	returningFields := "id, public_key, plugin_id, plugin_version, policy_version, plugin_type, signature, active, policy, progress"

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
		&updatedPolicy.PublicKey,
		&updatedPolicy.PluginID,
		&updatedPolicy.PluginVersion,
		&updatedPolicy.PolicyVersion,
		&updatedPolicy.PluginType,
		&updatedPolicy.Signature,
		&updatedPolicy.Active,
		&updatedPolicy.Policy,
		&updatedPolicy.Progress,
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
