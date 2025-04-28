package storage

import (
	"context"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vultisig/vultiserver-plugin/internal/types"
)

type PoolProvider interface {
	Pool() *pgxpool.Pool
}

type Transactor interface {
	PoolProvider
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

type DatabaseStorage interface {
	Transactor
	PolicyRepository
	TimeTriggerRepository
	TransactionRepository
	UserRepository
	PricingRepository
	PluginRepository
	PluginPricingRepository
	CategoryRepository
	TagRepository
	ReviewRepository
	RatingRepository
	Close() error
}

type PolicyRepository interface {
	GetPluginPolicy(ctx context.Context, id string) (types.PluginPolicy, error)
	GetAllPluginPolicies(ctx context.Context, pluginType string, publicKeyEcdsa string, take int, skip int) (types.PluginPolicyPaginatedList, error)

	DeletePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, id string) error
	InsertPluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error)
	UpdatePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error)
}

type TimeTriggerRepository interface {
	CreateTimeTriggerTx(ctx context.Context, dbTx pgx.Tx, trigger types.TimeTrigger) error
	GetPendingTimeTriggers(ctx context.Context) ([]types.TimeTrigger, error)
	UpdateTimeTriggerLastExecution(ctx context.Context, policyID string) error
	UpdateTimeTriggerTx(ctx context.Context, policyID string, trigger types.TimeTrigger, dbTx pgx.Tx) error
	DeleteTimeTrigger(ctx context.Context, policyID string) error
	UpdateTriggerStatus(ctx context.Context, policyID string, status types.TimeTriggerStatus) error
	GetTriggerStatus(ctx context.Context, policyID string) (types.TimeTriggerStatus, error)
}

type TransactionRepository interface {
	CountTransactions(ctx context.Context, policyID uuid.UUID, status types.TransactionStatus, txType string) (int64, error)
	CreateTransactionHistoryTx(ctx context.Context, dbTx pgx.Tx, tx types.TransactionHistory) (uuid.UUID, error)
	UpdateTransactionStatusTx(ctx context.Context, dbTx pgx.Tx, txID uuid.UUID, status types.TransactionStatus, metadata map[string]interface{}) error
	CreateTransactionHistory(ctx context.Context, tx types.TransactionHistory) (uuid.UUID, error)
	UpdateTransactionStatus(ctx context.Context, txID uuid.UUID, status types.TransactionStatus, metadata map[string]interface{}) error
	GetTransactionHistory(ctx context.Context, policyID uuid.UUID, transactionType string, take int, skip int) (types.TransactionHistoryPaginatedList, error)
	GetTransactionByHash(ctx context.Context, txHash string) (*types.TransactionHistory, error)
}

type UserRepository interface {
	FindUserById(ctx context.Context, userId string) (*types.User, error)
	FindUserByName(ctx context.Context, username string) (*types.UserWithPassword, error)
}

type PricingRepository interface {
	FindPricingById(ctx context.Context, id string) (*types.Pricing, error)
	CreatePricing(ctx context.Context, pricingDto types.PricingCreateDto) (*types.Pricing, error)
	DeletePricingById(ctx context.Context, id string) error
}

type PluginRepository interface {
	FindPlugins(ctx context.Context, filters types.PluginFilters, skip int, take int, sort string) (types.PluginsPaginatedList, error)
	FindPluginById(ctx context.Context, dbTx pgx.Tx, id string) (*types.Plugin, error)
	FindPluginByType(ctx context.Context, pluginType string) (*types.PluginPlain, error)
	CreatePlugin(ctx context.Context, dbTx pgx.Tx, pluginDto types.PluginCreateDto) (string, error)
	UpdatePlugin(ctx context.Context, id string, updates types.PluginUpdateDto) (*types.Plugin, error)
	DeletePluginById(ctx context.Context, id string) error
	AttachTagToPlugin(ctx context.Context, pluginId string, tagId string) (*types.Plugin, error)
	DetachTagFromPlugin(ctx context.Context, pluginId string, tagId string) (*types.Plugin, error)

	Pool() *pgxpool.Pool
}

type PluginPricingRepository interface {
	FindPluginPricingsBy(ctx context.Context, filters map[string]interface{}) ([]types.PluginPricing, error)
	CreatePluginPricing(ctx context.Context, pluginPricingDto types.PluginPricingCreateDto) (*types.PluginPricing, error)
}

type CategoryRepository interface {
	FindCategories(ctx context.Context) ([]types.Category, error)
}

type TagRepository interface {
	FindTags(ctx context.Context) ([]types.Tag, error)
	FindTagById(ctx context.Context, id string) (*types.Tag, error)
	FindTagByName(ctx context.Context, name string) (*types.Tag, error)
	CreateTag(ctx context.Context, tagDto types.CreateTagDto) (*types.Tag, error)
}

type ReviewRepository interface {
	CreateReview(ctx context.Context, reviewDto types.ReviewCreateDto, pluginId string) (string, error)
	FindReviews(ctx context.Context, pluginId string, take int, skip int, sort string) (types.ReviewsDto, error)
	FindReviewById(ctx context.Context, db pgx.Tx, id string) (*types.ReviewDto, error)
}

type RatingRepository interface {
	FindRatingByPluginId(ctx context.Context, dbTx pgx.Tx, pluginId string) ([]types.PluginRatingDto, error)
	CreateRatingForPlugin(ctx context.Context, dbTx pgx.Tx, pluginId string) error
	UpdateRatingForPlugin(ctx context.Context, dbTx pgx.Tx, pluginId string, reviewRating int) error
}
