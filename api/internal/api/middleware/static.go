package middleware

import (
	"net/http"
	"strings"

	"github.com/ahmetkoprulu/rtrp/common/storage"
	"github.com/gin-gonic/gin"
)

// StaticFileMiddleware serves files from the storage directory
func StaticFileMiddleware() gin.HandlerFunc {
	fileServer := http.FileServer(http.Dir(storage.StorageRoot))

	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, storage.PublicPrefix) {
			// Remove the /static prefix before serving
			c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, storage.PublicPrefix)
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}
		c.Next()
	}
}
