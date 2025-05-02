package types

import (
	"encoding/json"
	"time"
)

type Plugin struct {
	ID             string            `json:"id" validate:"required"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	Type           string            `json:"type" validate:"required"`
	Title          string            `json:"title" validate:"required"`
	Description    string            `json:"description" validate:"required"`
	Metadata       json.RawMessage   `json:"metadata" validate:"required"`
	ServerEndpoint string            `json:"server_endpoint" validate:"required"`
	PricingID      string            `json:"pricing_id" validate:"required"`
	CategoryID     string            `json:"category_id" validate:"required"`
	Tags           []Tag             `json:"tags"`
	Ratings        []PluginRatingDto `json:"ratings,omitempty"`
}

// plugin model with no relations
type PluginPlain struct {
	ID             string          `json:"id" validate:"required"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	Type           string          `json:"type" validate:"required"`
	Title          string          `json:"title" validate:"required"`
	Description    string          `json:"description" validate:"required"`
	Metadata       json.RawMessage `json:"metadata" validate:"required"`
	ServerEndpoint string          `json:"server_endpoint" validate:"required"`
	PricingID      string          `json:"pricing_id" validate:"required"`
	CategoryID     string          `json:"category_id" validate:"required"`
}

type PluginFilters struct {
	Term       *string `json:"term"`
	TagID      *string `json:"tag_id"`
	CategoryID *string `json:"category_id"`
}

type PluginsPaginatedList struct {
	Plugins    []Plugin `json:"plugins"`
	TotalCount int      `json:"total_count"`
}

type PluginCreateDto struct {
	Type           string          `json:"type" validate:"required"`
	Title          string          `json:"title" validate:"required"`
	Description    string          `json:"description" validate:"required"`
	Metadata       json.RawMessage `json:"metadata" validate:"required"`
	ServerEndpoint string          `json:"server_endpoint" validate:"required"`
	PricingID      string          `json:"pricing_id" validate:"required"`
	CategoryID     string          `json:"category_id" validate:"required"`
}

// using references on struct fields allows us to process partially field DTOs
type PluginUpdateDto struct {
	Title          *string          `json:"title"`
	Description    *string          `json:"description"`
	Metadata       *json.RawMessage `json:"metadata"`
	ServerEndpoint *string          `json:"server_endpoint"`
	PricingID      *string          `json:"pricing_id"`
	CategoryID     string           `json:"category_id" validate:"required"`
}
