package utils

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func ExtractToken(r *http.Request) string {
	tokenString := r.Header.Get("Authorization")

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	return tokenString
}

func GetClaims(r *http.Request) jwt.MapClaims {
	claims, _ := r.Context().Value("claims").(jwt.MapClaims)
	return claims
}

func GetUserID(r *http.Request) (string, error) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == ""{
		return "", errors.New("User ID not found")
	}
	return userID, nil
}

func GetUserRole(r *http.Request) (string, error) {
	role, ok := r.Context().Value("role").(string)
	if !ok || role == ""{
		return "", errors.New("User Role not found")
	}
	return role, nil
}
