package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/vultisig/vultiserver-plugin/internal/types"
)

const CATEGORIES_TABLE = "categories"

func (p *PostgresBackend) FindCategories(ctx context.Context) ([]types.Category, error) {
	query := fmt.Sprintf(`SELECT * FROM %s`, CATEGORIES_TABLE)

	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	categories, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Category])
	if err != nil {
		return nil, err
	}

	return categories, nil
}
