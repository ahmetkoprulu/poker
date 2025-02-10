package middleware

import (
	"net/http"
	"strings"

	"github.com/ahmetkoprulu/rtrp/common/utils"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/gin-gonic/gin"
)

const (
	UserIDKey   = "userID"
	PlayerIDKey = "playerID"
)

const (
	Token = "1234"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		result := models.ApiResponse[any]{
			Success: false,
			Status:  models.StatusUnauthorized,
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			result.Message = "authorization header is required"
			c.JSON(http.StatusUnauthorized, result)
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			result.Message = "invalid authorization header format"
			c.JSON(http.StatusUnauthorized, result)
			c.Abort()
			return
		}

		token := tokenParts[1]
		claims, err := utils.ValidateJwTTokenWithClaims(token)
		if err != nil {
			result.Message = err.Error()
			c.JSON(http.StatusUnauthorized, result)
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(PlayerIDKey, claims.PlayerID)
		c.Next()
	}
}

func ServerToServerAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ApiResponse[any]{
				Success: false,
				Status:  models.StatusUnauthorized,
				Message: "authorization header is required",
			})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ApiResponse[any]{
				Success: false,
				Status:  models.StatusUnauthorized,
				Message: "invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]
		if token != Token {
			c.JSON(http.StatusUnauthorized, models.ApiResponse[any]{
				Success: false,
				Status:  models.StatusUnauthorized,
				Message: "invalid token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
