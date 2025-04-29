package postgres

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/vultisig/vultiserver-plugin/internal/types"
)

const PLUGIN_PRICINGS_TABLE = "plugin_pricings"

func (p *PostgresBackend) findPluginPricingById(ctx context.Context, id string) (*types.PluginPricing, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1 LIMIT 1;`, PLUGIN_PRICINGS_TABLE)

	rows, err := p.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	plugin, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[types.PluginPricing])
	if err != nil {
		return nil, err
	}

	return &plugin, nil
}

func (p *PostgresBackend) FindPluginPricingsBy(ctx context.Context, filters map[string]interface{}) ([]types.PluginPricing, error) {
	query := fmt.Sprintf(`SELECT * FROM %s`, PLUGIN_PRICINGS_TABLE)

	// apply filters, if any
	paramIndex := 0
	var args []any
	for key, value := range filters {
		if paramIndex == 0 {
			query += " WHERE "
		} else {
			query += " AND "
		}

		paramIndex++

		query += fmt.Sprintf("%s = $%s", key, strconv.Itoa(paramIndex))
		args = append(args, value)
	}

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err

	}

	collection, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.PluginPricing])
	if err != nil {
		return nil, err
	}

	return collection, nil
}

func (p *PostgresBackend) CreatePluginPricing(
	ctx context.Context,
	pluginPricingDto types.PluginPricingCreateDto,
) (*types.PluginPricing, error) {
	query := fmt.Sprintf(`INSERT INTO %s (
		public_key_ecdsa,
		public_key_eddsa,
		plugin_type,
		is_ecdsa,
		chain_code_hex,
		derive_path,
		signature,
		policy
	) VALUES (
		@PublicKeyEcdsa,
		@PublicKeyEddsa,
		@PluginType,
		@IsEcdsa,
		@ChainCodeHex,
		@DerivePath,
		@Signature,
		@Policy
	) RETURNING id;`, PLUGIN_PRICINGS_TABLE)
	args := pgx.NamedArgs{
		"PublicKeyEcdsa": pluginPricingDto.PublicKeyEcdsa,
		"PublicKeyEddsa": pluginPricingDto.PublicKeyEddsa,
		"PluginType":     pluginPricingDto.PluginType,
		"IsEcdsa":        pluginPricingDto.IsEcdsa,
		"ChainCodeHex":   pluginPricingDto.ChainCodeHex,
		"DerivePath":     pluginPricingDto.DerivePath,
		"Signature":      pluginPricingDto.Signature,
		"Policy":         pluginPricingDto.Policy,
	}

	var createdId string
	err := p.pool.QueryRow(ctx, query, args).Scan(&createdId)
	if err != nil {
		return nil, err
	}

	return p.findPluginPricingById(ctx, createdId)
}
