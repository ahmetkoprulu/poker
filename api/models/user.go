package models

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type Player struct {
	ID     string `json:"id"`
	UserID string `json:"-"`
	Chips  int64  `json:"chips"`
}

type UserPlayer struct {
	ID     string `json:"id"`
	Player Player `json:"player"`
}
