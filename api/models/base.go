package models

const (
	StatusSuccess      = 200
	StatusError        = 500
	StatusUnauthorized = 401
)

type ApiResponse[T any] struct {
	Success bool   `json:"success"`
	Status  int    `json:"status"`
	Data    T      `json:"data"`
	Message string `json:"message"`
}
