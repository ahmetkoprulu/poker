package models

type Config struct {
	DatabaseURL  string
	DatabaseName string
	MqURL        string
	CacheURL     string
	JWTSecret    string
	ServiceName  string
	ServerPort   string
	BaseUrl      string
}
