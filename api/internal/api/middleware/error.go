package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError represents a custom application error
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

// ErrorMiddleware handles application errors and returns appropriate responses
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var appErr *AppError
			if errors.As(err, &appErr) {
				// Handle custom application error
				c.JSON(appErr.Code, gin.H{
					"error": appErr.Message,
				})
				return
			}

			// Handle other errors
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
	}
}

// NewAppError creates a new application error
func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Common application errors
var (
	ErrUnauthorized = NewAppError(http.StatusUnauthorized, "unauthorized")
	ErrForbidden    = NewAppError(http.StatusForbidden, "forbidden")
	ErrNotFound     = NewAppError(http.StatusNotFound, "resource not found")
	ErrBadRequest   = NewAppError(http.StatusBadRequest, "bad request")
)
