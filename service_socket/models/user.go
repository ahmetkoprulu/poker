package models

type User struct {
	ID         string        `json:"id"`
	Provider   SocialNetwork `json:"provider"`
	Identifier string        `json:"identifier"`
	Player     *Player       `json:"player,omitempty"`
}

type Player struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	ProfilePicURL string `json:"profile_pic_url"`
	UserID        string `json:"-"`
	Chips         int64  `json:"chips"`
}

type SocialNetwork int

const (
	Guest SocialNetwork = iota
	Email
	Google
	Facebook
	Apple
)
