package models

type Config struct {
	MqURL        string
	DatabaseURL  string
	DatabaseName string
	CacheURL     string
	JWTSecret    string
	ServiceName  string
	ServerPort   string
	BaseUrl      string
	ApiUrl       string
}
