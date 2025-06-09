package config

import (
	"os"

	"github.com/ahmetkoprulu/rtrp/consumers/models"
	"github.com/joho/godotenv"
)

var config *models.Config

func LoadEnvironment() *models.Config {
	err := godotenv.Load()
	if err != nil {
		return nil
	}

	config = &models.Config{
		MqURL:        os.Getenv("MQ_URL"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		DatabaseName: os.Getenv("DATABASE_NAME"),
		CacheURL:     os.Getenv("CACHE_URL"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		ServiceName:  os.Getenv("SERVICE_NAME"),
		ServerPort:   os.Getenv("PORT"),
		BaseUrl:      os.Getenv("BASE_URL"),
		ApiUrl:       os.Getenv("API_URL"),
	}

	return config
}

func GetConfig() *models.Config {
	if config == nil {
		LoadEnvironment()
	}

	return config
}
