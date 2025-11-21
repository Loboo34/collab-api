package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Loboo34/collab-api/utils"
)

func CheckRole(userRole string, next http.HandlerFunc) http.HandlerFunc{
	return  func (w http.ResponseWriter, r *http.Request)  {
	
		tokenString := r.Header.Get("Authorization")
		if tokenString == ""{
			utils.RespondWithError(w, http.StatusUnauthorized, "Missing Auth Token", "")
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		claims, err := utils.ValidateJWT(tokenString)
		if err != nil{
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid Auth token", "")
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Missing User Role", "")
			return
		}

		if !strings.EqualFold(role, userRole){
			utils.RespondWithError(w, http.StatusForbidden, "Not Permited to perform Action", "")
			return
		}

		next.ServeHTTP(w, r)

	}
}

func CheckAuth(next http.HandlerFunc) http.HandlerFunc{
	return func (w http.ResponseWriter, r *http.Request){
		tokenString := r.Header.Get("Authorization")
		if tokenString == ""{
			utils.RespondWithError(w, http.StatusUnauthorized, "Missing Auth token", "")
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid Auth Token", "")
			return
		}

		ctx := context.WithValue(r.Context(), "claims", claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}
}