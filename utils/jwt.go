package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey []byte

func InitJWT() error {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return errors.New("JWT_SECRET not set")
	}
	jwtKey = []byte(secret)
	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)

	return string(bytes), err
}

func ComparePassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateJWT(userID, email, role string) (string, error) {
	if len(jwtKey) == 0 {
		return "", errors.New("JWT key not initialized")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    userID,
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(jwtKey)
}

func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	if len(jwtKey) == 0 {
		return nil, errors.New("JWT key not initialized")
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}
