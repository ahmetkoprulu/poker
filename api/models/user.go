package models

import (
	"strconv"

	"golang.org/x/exp/rand"
)

type User struct {
	ID         string        `json:"id"`
	Provider   SocialNetwork `json:"provider"`
	Identifier string        `json:"identifier"`
	Password   string        `json:"-"`
	Profile    Profile       `json:"profile"`
	Player     *Player       `json:"player,omitempty"`
}

type Profile struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
	Phone string `json:"phone,omitempty"`
}

type Player struct {
	ID            string `json:"id"`
	Username      string `json:"username,omitempty"`
	ProfilePicURL string `json:"profile_pic_url,omitempty"`
	UserID        string `json:"-"`
	Chips         int64  `json:"chips"`
}

type UserPlayer struct {
	ID       string        `json:"id"`
	Provider SocialNetwork `json:"provider"`
	Player   Player        `json:"player"`
}

type SocialNetwork int

const (
	Guest SocialNetwork = iota
	Email
	Google
	Facebook
	Apple
)

// GetEmail returns the email from either the identifier (for email provider) or profile
func (u *User) GetEmail() string {
	if u.Provider == Email {
		return u.Identifier
	}

	return u.Profile.Email
}

// NewEmailUser creates a new user with email authentication
func NewEmailUser(email string, password string) *User {
	return &User{
		Provider:   Email,
		Identifier: email,
		Password:   password,
		Profile: Profile{
			Email: email,
		},
	}
}

func NewGuestUser(identifier string) *User {
	return &User{
		Provider:   Guest,
		Identifier: identifier,
	}
}

// NewSocialUser creates a new user with social authentication
func NewSocialUser(provider SocialNetwork, providerUserID string, email string, name string) *User {
	return &User{
		Provider:   provider,
		Identifier: providerUserID,
		Profile: Profile{
			Email: email,
			Name:  name,
		},
	}
}

func NewGuestPlayer(userID string) *Player {
	return &Player{
		ID:            userID,
		Username:      "Guest" + strconv.Itoa(rand.Intn(1000000)),
		ProfilePicURL: "avatar_0",
		Chips:         1000,
	}
}

func NewPlayer(userID string, username string, profilePicURL string, chips int64) *Player {
	if username == "" {
		username = GenerateUserName()
	}

	if profilePicURL == "" {
		profilePicURL = "avatar_" + strconv.Itoa(rand.Intn(10))
	}

	return &Player{
		UserID:        userID,
		Username:      username,
		ProfilePicURL: profilePicURL,
		Chips:         chips,
	}
}

func GenerateUserName() string {
	// select two random item from the list and combine them. username set should be meaningful objects and subjects
	usernameSet := []string{"Cat", "Dog", "Bird", "Fish", "Snake", "Lamp", "Table", "Chair", "Book", "Pen", "Pencil", "Computer", "Phone", "TV", "Car", "Bike", "House", "Tree", "Flower", "Sun", "Moon", "Star", "Cloud", "Rain", "Snow", "Fire", "Water", "Earth", "Wind", "Fire", "Water", "Earth", "Wind"}
	username := usernameSet[rand.Intn(len(usernameSet))] + usernameSet[rand.Intn(len(usernameSet))]
	return username
}
