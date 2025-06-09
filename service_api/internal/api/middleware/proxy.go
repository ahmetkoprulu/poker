package middleware

import (
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func ReverseProxy(frontendURL string) gin.HandlerFunc {
	target, err := url.Parse(frontendURL)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(c *gin.Context) {
		// Skip proxy for API routes and Swagger
		if strings.HasPrefix(c.Request.URL.Path, "/api") ||
			strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			c.Next()
			return
		}

		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
