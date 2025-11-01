package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Loboo34/collab-api/database"
	"github.com/Loboo34/collab-api/models"
	"github.com/Loboo34/collab-api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", "")
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
		return
	}
	team.ID = primitive.NewObjectID()
	team.CreatedBy = userID
	team.Members = []string{userID}
	team.CreatedAt = time.Now()

	teamCollection := database.DB.Collection("teams")

	var members models.TeamMember
	members.TeamId = team.ID
	members.User = userID
	members.Role = "Admin"
	members.JoinedAt = time.Now()

	membersCollection := database.DB.Collection("team-members")

	//var user models.User
	//user.Teams = team.ID

	userCollection := database.DB.Collection("users")
	userObjID, _ := primitive.ObjectIDFromHex(userID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = teamCollection.InsertOne(ctx, team)
	if err != nil {
		utils.Logger.Warn("Failed to Create team")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error creating Team", "")
		return
	}

	_, err = membersCollection.InsertOne(ctx, members)
	if err != nil {
		utils.Logger.Warn("Failed to create team admin")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error saving admin", "")
	}

	_, err = userCollection.UpdateOne(ctx, bson.M{"_id": userObjID}, bson.M{"$addToSet": bson.M{"teams": team.ID}})

	utils.RespondWithJSON(w, http.StatusCreated, "Team created Successfully", map[string]interface{}{"team_id": team.ID.Hex(),
		"name": team.Name})
}

func InviteMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Post Allowed", "")
		return
	}

	tokenString := r.Header.Get("Authorizarion")
	if tokenString == ""{
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing token", "")
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	
	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", "")
	}

	userId, ok := claims["id"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "user ID not found", "")
		return
	}

	userRole, ok := claims["role"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User role not found", "")
		return
	}


	isAdmin, err := strings.EqualFold(userRole, "Admin")
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "User is not admin", "")
		return
	}

	var user models.User
	if err = json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "INvalid json format","")
		return
	}

	
}
