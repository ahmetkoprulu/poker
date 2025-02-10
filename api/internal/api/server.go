package api

import (
	"context"
	"net/http"
	"time"

	"github.com/ahmetkoprulu/rtrp/common/data"

	// _ "github.com/ahmetkoprulu/bidi-menu/docs"
	"github.com/ahmetkoprulu/rtrp/internal/api/handlers"
	"github.com/ahmetkoprulu/rtrp/internal/api/middleware"
	"github.com/ahmetkoprulu/rtrp/internal/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	router        *gin.Engine
	httpServer    *http.Server
	authService   *services.AuthService
	playerService *services.PlayerService
	db            *data.PgDbContext
}

func NewServer(db *data.PgDbContext) *Server {
	authService := services.NewAuthService(db)
	playerService := services.NewPlayerService(db)
	server := &Server{
		router:        gin.Default(),
		authService:   authService,
		playerService: playerService,
		db:            db,
	}

	server.router.Use(middleware.RequestLogger())
	server.router.Use(middleware.CORSMiddleware())
	server.router.Use(middleware.ErrorMiddleware())
	server.router.Use(middleware.RateLimit(100, 200)) // 100 requests per second with burst of 200
	server.router.Use(middleware.StaticFileMiddleware())

	authHandler := handlers.NewAuthHandler(authService)
	playerHandler := handlers.NewPlayerHandler(playerService)
	healthHandler := handlers.NewHealthHandler()

	authMiddleware := middleware.AuthMiddleware()
	serverToServerAuthMiddleware := middleware.ServerToServerAuthMiddleware()

	healthHandler.RegisterRoutes(server.router.Group(""))

	v1 := server.router.Group("/api/v1")
	{
		authHandler.RegisterRoutes(v1)
		playerHandler.RegisterRoutes(v1, authMiddleware, serverToServerAuthMiddleware)

		// Protected routes
		// protected := v1.Group("", authMiddleware)
		// {

		// }
	}

	// Swagger documentation
	server.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return server
}

func (s *Server) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// return s.httpServer.ListenAndServeTLS("./ssl/cert.pem", "./ssl/key.pem")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
