package utils

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecret []byte

type Claims struct {
	UserID   string `json:"userId"`
	PlayerID string `json:"playerId"`
	jwt.StandardClaims
}

// Valid implements the jwt.Claims interface
func (c Claims) Valid() error {
	return c.StandardClaims.Valid()
}

// SetJWTSecret sets the secret used to sign JWT tokens.
// This function should be called during the initialization of the application.
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

// GenerateJWTToken generates a JWT token given a user ID. The token is signed with the
// secret set by SetJWTSecret and will contain the user ID as a claim. The token does not
// contain an expiration time, so it should be used carefully.
func GenerateJWTToken(userId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		// "exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(jwtSecret)
}

// GenerateJWTTokenWithClaims generates a JWT token given a Claims object. The token is signed with
// the secret set by SetJWTSecret and will contain the claims as part of the token.
func GenerateJWTTokenWithClaims(claims Claims) (string, error) {
	claims.StandardClaims = jwt.StandardClaims{
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
		IssuedAt:  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWTToken validates a JWT token. It returns the user ID
// contained in the token, or an error if the token is invalid.
func ValidateJWTToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId := claims["userId"].(string)
		if userId == "" {
			return "", jwt.ErrInvalidKey
		}
		return userId, nil
	}

	return "", jwt.ErrSignatureInvalid
}

// ValidateJwTTokenWithClaims validates a JWT token. It returns the Claims object
// contained in the token, or an error if the token is invalid.
func ValidateJwTTokenWithClaims(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func GetBearerToken(token string) string {
	return token[7:]
}
