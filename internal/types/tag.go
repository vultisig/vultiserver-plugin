package types

type CreateTagDto struct {
	Name  string `json:"name" validate:"required"`
	Color string `json:"color" validate:"required"`
}

type Tag struct {
	ID    string `json:"id" validate:"required"`
	Name  string `json:"name" validate:"required"`
	Color string `json:"color" validate:"required"`
}
