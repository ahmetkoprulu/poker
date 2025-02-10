package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/common/utils"
	"github.com/ahmetkoprulu/rtrp/internal/api"
	"github.com/ahmetkoprulu/rtrp/internal/config"
)

func main() {
	config := config.LoadEnvironment()
	fmt.Println(config)

	utils.InitLogger()
	defer utils.Logger.Sync()

	utils.SetJWTSecret(config.JWTSecret)
	err := data.LoadPostgres(config.DatabaseURL, config.DatabaseName)
	if err != nil {
		log.Fatalf("Failed to load Postgres: %v\n", err)
	}

	db, err := data.NewPgDbContext()
	if err != nil {
		utils.Logger.Fatal("Failed to connect to database", utils.Logger.String("error", err.Error()))
	}
	defer db.Close()

	// redis, err := cache.NewRedisCache(config.CacheURL, 0)
	// if err != nil {
	// 	utils.Logger.Fatal("Failed to connect to redis", utils.Logger.String("error", err.Error()))
	// }
	// defer redis.Close()

	server := api.NewServer(db)
	go func() {
		addr := fmt.Sprintf(":%s", os.Getenv("PORT"))
		if err := server.Start(addr); err != nil {
			utils.Logger.Fatal("Failed to start server", utils.Logger.String("error", err.Error()))
		}
	}()

	utils.Logger.Info("Server started successfully", utils.Logger.String("port", os.Getenv("PORT")))

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Logger.Info("Server is shutting down...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		utils.Logger.Fatal("Server forced to shutdown", utils.Logger.String("error", err.Error()))
	}

	utils.Logger.Info("Server exited gracefully")
}
