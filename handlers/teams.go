package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Loboo34/collab-api/database"
	"github.com/Loboo34/collab-api/models"
	"github.com/Loboo34/collab-api/utils"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jwtKey = os.Getenv("JWT_SECRET")

func CreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Post Allowed", "")
		return
	}

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing Auth token", "")
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", "")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User id not found", "")
		return
	}

	userID, ok := claims["id"].(string)
	if !ok {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	var team models.Team
	if err = json.NewDecoder(r.Body).Decode(&team); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid json", "")
	}
	team.ID = primitive.NewObjectID()
	team.CreatedBy = userID

	var admin models.TeamMember
	admin.User = userID
	admin.Role = "Admin"

	collection := database.DB.Collection("teams")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, team)
	if err != nil {
		utils.Logger.Warn("Failed to Create team")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error creating Team", "")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "Created Team successfully"})
}
