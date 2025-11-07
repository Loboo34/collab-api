package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Loboo34/collab-api/database"
	"github.com/Loboo34/collab-api/models"
	"github.com/Loboo34/collab-api/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing token", "")
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", "")
		return
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

	if !strings.EqualFold(userRole, "Admin") {
		utils.RespondWithError(w, http.StatusUnauthorized, "User is not admin", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type InvetmentRequest struct {
		Email  string `json:"email"`
		TeamId string `json:"teamId"`
	}

	var req InvetmentRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid json format", "")
		return
	}

	userCollection := database.DB.Collection("users")
	var user models.User
	err = userCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "User not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding user", "")
		}
		return
	}

	teamObjId, err := primitive.ObjectIDFromHex(req.TeamId)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid team id", "")
		return
	}

	teamCollection := database.DB.Collection("teams")
	var team models.Team
	err = teamCollection.FindOne(ctx, bson.M{"_id": teamObjId}).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Team not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding team", "")
		}
		return
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error generating token", "")
		return
	}
	inviteToken := hex.EncodeToString(tokenBytes)

	invite := models.Invite{
		ID:        primitive.NewObjectID(),
		TeamID:    teamObjId,
		Email:     user.Email,
		Token:     inviteToken,
		Status:    "pending",
		SentBy:    userId,
		CreatedAt: time.Now(),
	}

	inviteCollection := database.DB.Collection("invites")
	_, err = inviteCollection.InsertOne(ctx, invite)
	if err != nil {
		utils.Logger.Warn("Failed to create invite")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error creating invite", "")
	}

	inviteLink := "http://localhost:3000/invite/accept?token=" + inviteToken

	if err := utils.SendInviteEmail(user.Email, inviteLink); err != nil {
		utils.Logger.Warn("Failed to send email")
		utils.RespondWithError(w, http.StatusInternalServerError, "Erreor sending email", "")
	}

	utils.RespondWithJSON(w, http.StatusCreated, "Invitation sent successfully", map[string]interface{}{"user": user.ID})
}

func AcceptInvite(w http.ResponseWriter, r *http.Request) {
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
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid JWT", "")
		return
	}

	userID, ok := claims["id"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusNotFound, "Missing Id", "")
		return
	}

	id, _ := primitive.ObjectIDFromHex(userID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inviteCollection := database.DB.Collection("invites")
	var invite models.Invite
	err = inviteCollection.FindOne(ctx, bson.M{"token": invite.Token}).Decode(&invite)

	membersCollection := database.DB.Collection("team-members")
	_, err = membersCollection.UpdateOne(ctx, bson.M{"_id": invite.TeamID}, bson.M{"$addToSet": bson.M{"teamMembers": id}})
	if err != nil {
		utils.Logger.Warn("Failed to Add user to team")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error adding member to team", "")
		return
	}

	_, err = inviteCollection.UpdateOne(ctx, bson.M{"_id": invite.ID}, bson.M{"$set": bson.M{"status": "accepted"}})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error changing invite status", "")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Invite Accepted", map[string]interface{}{"user": id})
}

func DeclineInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Post Allowe", "")
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
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token String", "")
		return
	}

	userID, ok := claims["id"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusNotFound, "Id Not found", "")
		return
	}

	id, _ := primitive.ObjectIDFromHex(userID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inviteCollectioon := database.DB.Collection("invites")
	var invite models.Invite

	_, err = inviteCollectioon.UpdateOne(ctx, bson.M{"_id": invite.ID}, bson.M{"$set": bson.M{"status": "declined"}})
	if err != nil {
		utils.Logger.Warn("Failed to decline Invitation")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error declinign invitation", "")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Invite declined", map[string]interface{}{"user": id})
}

func GetTeamMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Get Allowed", "")
		return
	}

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		utils.RespondWithError(w, http.StatusNotFound, "Missing Auth token", "")
		return
	}

	teamCollection := database.DB.Collection("teams")
	var team models.Team

	membersCollection := database.DB.Collection("team-members")
	var members []models.TeamMember

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := teamCollection.FindOne(ctx, bson.M{"_id": team.ID}).Decode(&team)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Error getting team id", "")
		return
	}

	cursor, err := membersCollection.Find(ctx, bson.M{})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching ", "")
		return
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var member models.TeamMember
		if err := cursor.Decode(&member); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching ", "")
			return
		}
		members = append(members, member)
	}

	if err = cursor.Err(); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Cursor error", "")
		return
	}

	utils.Logger.Info("Fetched All team members")
	utils.RespondWithJSON(w, http.StatusOK, "", map[string]interface{}{"members": members})
}

func DeleteTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Delete Allowed", "")
		return
	}

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		utils.RespondWithError(w, http.StatusNotFound, "Missing Auth token", "")
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
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing Id", "")
		return
	}

	role, ok := claims["role"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing role", "")
		return
	}

	if !strings.EqualFold(role, "Admin") {
		utils.RespondWithError(w, http.StatusUnauthorized, "Only Admin can Perform Action", "")
		return
	}

	var team models.Team
	teamCollection := database.DB.Collection("teams")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//teamID, err := primitive.ObjectIDFromHex(team.ID)

	result, err := teamCollection.DeleteOne(ctx, bson.M{"_id": team.ID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete Team", "")
		return
	}

	if result.DeletedCount == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Team Not found", "")
		return
	}

	utils.Logger.Info("Deleted Team")
	utils.RespondWithJSON(w, http.StatusOK, "Team successfuly deleted", map[string]interface{}{"Team deleted by": userID, "team": team})
}
