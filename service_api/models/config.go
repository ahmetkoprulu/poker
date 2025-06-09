package models

type Config struct {
	DatabaseURL  string
	DatabaseName string
	MqURL        string
	CacheURL     string
	ElasticUrl   string
	JWTSecret    string
	ServiceName  string
	ServerPort   string
	SocialConfig SocialConfig
	EmailConfig  EmailConfig
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

type SocialConfig struct {
	FacebookClientID     string
	FacebookClientSecret string
	GoogleClientID       string
	GoogleClientSecret   string
}
