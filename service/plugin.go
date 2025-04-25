package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultiserver-plugin/internal/types"
)

type Plugin interface {
	CreatePluginWithRating(ctx context.Context, pluginDto types.PluginCreateDto) (*types.Plugin, error)
	GetPluginWithRating(ctx context.Context, pluginId string) (*types.Plugin, error)
	CreatePluginReviewWithRating(ctx context.Context, reviewDto types.ReviewCreateDto, pluginId string) (*types.ReviewDto, error)
}

type PluginServiceStorage interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error

	FindRatingByPluginId(ctx context.Context, dbTx pgx.Tx, pluginId string) ([]types.PluginRatingDto, error)
	CreateRatingForPlugin(ctx context.Context, dbTx pgx.Tx, pluginId string) error
	UpdateRatingForPlugin(ctx context.Context, dbTx pgx.Tx, pluginId string, reviewRating int) error

	CreateReview(ctx context.Context, reviewDto types.ReviewCreateDto, pluginId string) (string, error)
	FindReviews(ctx context.Context, pluginId string, take int, skip int, sort string) (types.ReviewsDto, error)
	FindReviewById(ctx context.Context, db pgx.Tx, id string) (*types.ReviewDto, error)

	CreatePlugin(ctx context.Context, dbTx pgx.Tx, pluginDto types.PluginCreateDto) (string, error)
	FindPluginById(ctx context.Context, dbTx pgx.Tx, id string) (*types.Plugin, error)
}

type PluginService struct {
	db     PluginServiceStorage
	logger *logrus.Logger
}

func NewPluginService(db PluginServiceStorage, logger *logrus.Logger) (*PluginService, error) {
	if db == nil {
		return nil, fmt.Errorf("database storage cannot be nil")
	}
	return &PluginService{
		db:     db,
		logger: logger,
	}, nil
}

func (s *PluginService) CreatePluginWithRating(ctx context.Context, pluginDto types.PluginCreateDto) (*types.Plugin, error) {
	var plugin *types.Plugin
	err := s.db.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error

		// Insert plugin
		newPluginId, err := s.db.CreatePlugin(ctx, tx, pluginDto)
		if err != nil {
			return fmt.Errorf("failed to insert plugin: %w", err)
		}

		// Insert rating
		err = s.db.CreateRatingForPlugin(ctx, tx, newPluginId)
		if err != nil {
			return fmt.Errorf("failed to insert rating: %w", err)
		}

		// Find plugin
		plugin, err = s.db.FindPluginById(ctx, tx, newPluginId)
		if err != nil {
			return fmt.Errorf("failed to get plugin: %w", err)
		}

		// Find rating
		rating, err := s.db.FindRatingByPluginId(ctx, tx, newPluginId)
		if err != nil {
			return fmt.Errorf("failed to get rating: %w", err)
		}

		plugin.Ratings = rating

		return nil
	})

	if err != nil {
		return nil, err
	}
	return plugin, nil
}

func (s *PluginService) GetPluginWithRating(ctx context.Context, pluginId string) (*types.Plugin, error) {
	var plugin *types.Plugin
	err := s.db.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error

		// Find plugin
		plugin, err = s.db.FindPluginById(ctx, tx, pluginId)

		if err != nil {
			return fmt.Errorf("failed to get plugin: %w", err)
		}

		// Find rating
		var rating []types.PluginRatingDto
		rating, err = s.db.FindRatingByPluginId(ctx, tx, pluginId)
		if err != nil {
			return fmt.Errorf("failed to get rating: %w", err)
		}

		plugin.Ratings = rating

		return nil
	})
	if err != nil {
		return nil, err
	}
	return plugin, nil

}

func (s *PluginService) CreatePluginReviewWithRating(ctx context.Context, reviewDto types.ReviewCreateDto, pluginId string) (*types.ReviewDto, error) {
	var review *types.ReviewDto
	err := s.db.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		// Insert review
		reviewId, err := s.db.CreateReview(ctx, reviewDto, pluginId)
		if err != nil {
			return fmt.Errorf("failed to create review: %w", err)
		}

		// Update rating
		err = s.db.UpdateRatingForPlugin(ctx, tx, pluginId, reviewDto.Rating)
		if err != nil {
			return fmt.Errorf("failed to update rating: %w", err)
		}

		// Find review
		review, err = s.db.FindReviewById(ctx, tx, reviewId)
		if err != nil {
			return fmt.Errorf("failed to get plugin: %w", err)
		}

		// Find rating
		rating, err := s.db.FindRatingByPluginId(ctx, tx, pluginId)
		if err != nil {
			return fmt.Errorf("failed to get rating: %w", err)
		}

		review.Ratings = rating

		return nil
	})
	if err != nil {
		return nil, err
	}
	return review, nil

}
