package api

import (
	"time"

	"github.com/ahmetkoprulu/rtrp/game/internal/config"
)

type ApiService struct {
	config        ApiServiceConfig
	clientFactory *Factory
	AuthService   *AuthService
}

func NewApiService() *ApiService {
	config := config.GetConfig()
	apiConfig := ApiServiceConfig{
		BaseURL:              config.ApiUrl,
		Timeout:              10 * time.Second,
		RetryCount:           3,
		RetryDelay:           500 * time.Millisecond,
		EnableRequestLogging: true,
	}

	factory := NewFactory()
	service := &ApiService{
		config:        apiConfig,
		clientFactory: factory,
	}
	service.AuthService = NewAuthService(service, "/auth")

	return service
}

func (s *ApiService) getClient(serviceName string, relativePath string) *ApiClient {
	clientConfig := DefaultConfig()
	clientConfig.BaseURL = s.config.BaseURL + relativePath
	clientConfig.Timeout = s.config.Timeout
	client := s.clientFactory.GetOrCreate(serviceName, clientConfig)

	// if s.config.RetryCount > 0 {
	// 	client.SetHeader("X-Service-Name", serviceName) // Just a way to identify the service

	// 	// Apply retry middleware
	// 	client.UseMiddleware(RetryMiddleware(s.config.RetryCount, s.config.RetryDelay))

	// 	// Apply logging middleware if enabled
	// 	if s.config.EnableRequestLogging {
	// 		client.UseMiddleware(LoggingMiddleware())
	// 	}
	// }

	return client
}

type ApiServiceConfig struct {
	BaseURL              string
	Timeout              time.Duration
	RetryCount           int
	RetryDelay           time.Duration
	EnableRequestLogging bool
}
