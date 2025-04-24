package types

type Category struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}
