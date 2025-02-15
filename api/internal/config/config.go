package config

import (
	"os"

	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/joho/godotenv"
)

var config *models.Config

func GetConfig() *models.Config {
	if config == nil {
		config = LoadEnvironment()
	}
	return config
}

func LoadEnvironment() *models.Config {
	godotenv.Load()

	config = &models.Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		DatabaseName: os.Getenv("DATABASE_NAME"),
		MqURL:        os.Getenv("MQ_URL"),
		CacheURL:     os.Getenv("CACHE_URL"),
		ElasticUrl:   os.Getenv("ELASTIC_URL"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		ServiceName:  os.Getenv("SERVICE_NAME"),
		ServerPort:   os.Getenv("PORT"),
		SocialConfig: models.SocialConfig{
			FacebookClientID:     os.Getenv("FACEBOOK_CLIENT_ID"),
			FacebookClientSecret: os.Getenv("FACEBOOK_CLIENT_SECRET"),
			GoogleClientID:       os.Getenv("GOOGLE_CLIENT_ID"),
			GoogleClientSecret:   os.Getenv("GOOGLE_CLIENT_SECRET"),
		},
		EmailConfig: models.EmailConfig{
			SMTPHost:     os.Getenv("SMTP_HOST"),
			SMTPPort:     os.Getenv("SMTP_PORT"),
			SMTPUsername: os.Getenv("SMTP_USERNAME"),
			SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		},
	}

	return config
}
