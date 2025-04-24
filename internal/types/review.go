package types

import "time"

type Review struct {
	ID        string    `json:"id" validate:"required"`
	Address   string    `json:"address" validate:"required"`
	Rating    int       `json:"rating" validate:"required"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	PluginId  string    `json:"plugin_id" validate:"required"`
}

type ReviewDto struct {
	ID        string            `json:"id" validate:"required"`
	Address   string            `json:"address" validate:"required"`
	Rating    int               `json:"rating" validate:"required"`
	Comment   string            `json:"comment,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	PluginId  string            `json:"plugin_id" validate:"required"`
	Ratings   []PluginRatingDto `json:"ratings,omitempty"`
}

type ReviewCreateDto struct {
	Address string `json:"address" validate:"required"`
	Rating  int    `json:"rating" validate:"required"`
	Comment string `json:"comment,omitempty"`
}

type ReviewsDto struct {
	Reviews    []Review `json:"reviews"`
	TotalCount int      `json:"total_count"`
}

type PluginRating struct {
	PluginID string `json:"plugin_id"`
	Rating   int    `json:"rating"`
	Count    int    `json:"count"`
}

type PluginRatingDto struct {
	Rating int `json:"rating"`
	Count  int `json:"count"`
}
