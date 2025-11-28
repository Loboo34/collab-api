package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Loboo34/collab-api/utils"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func CheckAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		token, err := utils.ExtractToken(r)
		if err != nil {
			utils.Logger.Warn("Auth token Not found: " + err.Error())
			utils.RespondWithError(w, http.StatusUnauthorized, "Missing Auth token", "")
			return
		}

		utils.Logger.Info("Token: " + token[:min(50, len(token))] + "...")

		claims, err := utils.ValidateJWT(token)
		if err != nil {
			utils.Logger.Warn("JWT validation failed" + err.Error())
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid Auth Token", "")
			return
		}
		userID, ok := claims["id"].(string)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid User ID", "")
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid User Role", "")
			return
		}

		ctx := context.WithValue(r.Context(), "claims", claims)
		ctx = context.WithValue(ctx, "userID", userID)
		ctx = context.WithValue(ctx, "role", role)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func CheckRole(userRole string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		role := r.Context().Value("role")
		if role == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Missing User Role", "")
			return
		}

		if !strings.EqualFold(role.(string), userRole) {
			utils.RespondWithError(w, http.StatusForbidden, "Not Permited to perform Action", "")
			return
		}

		next.ServeHTTP(w, r)

	}
}
