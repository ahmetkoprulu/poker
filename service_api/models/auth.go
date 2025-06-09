package models

type LoginRequest struct {
	Provider   SocialNetwork `json:"provider"`
	Identifier string        `json:"identifier" binding:"required"` // email for Email provider, token for social providers
	Secret     string        `json:"secret"`                        // only required for Email provider
}

type LoginResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

type RegisterRequest struct {
	Provider   SocialNetwork `json:"provider" binding:"required"`
	Identifier string        `json:"identifier" binding:"required"` // email for Email provider, token for social providers
	Secret     string        `json:"secret"`                        // only required for Email provider
}

type RegisterResponse struct {
	User  UserPlayer `json:"user"`
	Token string     `json:"token"`
}
