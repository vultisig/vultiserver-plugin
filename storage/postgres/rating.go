package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/vultisig/vultiserver-plugin/internal/types"
)

const PLUGIN_RATING_TABLE = "plugin_rating"

func (p *PostgresBackend) FindRatingByPluginId(ctx context.Context, dbTx pgx.Tx, pluginId string) ([]types.PluginRatingDto, error) {
	query := fmt.Sprintf(`
	SELECT *
    FROM %s
    WHERE plugin_id = $1`, PLUGIN_RATING_TABLE)

	rows, err := dbTx.Query(ctx, query, pluginId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []types.PluginRatingDto
	for rows.Next() {
		var review types.PluginRating
		err := rows.Scan(
			&review.PluginID,
			&review.Rating,
			&review.Count,
		)
		if err != nil {
			return nil, err
		}

		var reviewDto types.PluginRatingDto
		reviewDto.Rating = review.Rating
		reviewDto.Count = review.Count
		ratings = append(ratings, reviewDto)
	}

	return ratings, nil
}

func (p *PostgresBackend) CreateRatingForPlugin(ctx context.Context, dbTx pgx.Tx, pluginId string) error {
	ratingQuery := fmt.Sprintf(`INSERT INTO %s (plugin_id, rating, count)
	      VALUES ($1, 1, 0), ($1, 2, 0), ($1, 3, 0), ($1, 4, 0), ($1, 5, 0)`, PLUGIN_RATING_TABLE)

	_, err := dbTx.Exec(ctx, ratingQuery, pluginId)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresBackend) UpdateRatingForPlugin(ctx context.Context, dbTx pgx.Tx, pluginId string, reviewRating int) error {
	ratingQuery := fmt.Sprintf(`
	UPDATE %s
	SET count = count + 1
	WHERE plugin_id = $1 AND rating = $2`, PLUGIN_RATING_TABLE)

	ct, err := dbTx.Exec(ctx, ratingQuery, pluginId, reviewRating)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("%s row not found for plugin_id=%s rating=%d", PLUGIN_RATING_TABLE, pluginId, reviewRating)
	}

	return nil
}
