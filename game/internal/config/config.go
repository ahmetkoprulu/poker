package config

import (
	"os"

	"github.com/ahmetkoprulu/rtrp/game/models"
	"github.com/joho/godotenv"
)

func LoadEnvironment() *models.Config {
	err := godotenv.Load()
	if err != nil {
		return nil
	}

	return &models.Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		DatabaseName: os.Getenv("DATABASE_NAME"),
		MqURL:        os.Getenv("MQ_URL"),
		CacheURL:     os.Getenv("CACHE_URL"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		ServiceName:  os.Getenv("SERVICE_NAME"),
		ServerPort:   os.Getenv("PORT"),
		BaseUrl:      os.Getenv("BASE_URL"),
	}
}
