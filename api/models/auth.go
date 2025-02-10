package models

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User  UserPlayer `json:"user"`
	Token string     `json:"token"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterResponse struct {
	User  UserPlayer `json:"user"`
	Token string     `json:"token"`
}
