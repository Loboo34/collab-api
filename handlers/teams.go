package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"

	"net/http"
	"time"

	"github.com/Loboo34/collab-api/database"
	"github.com/Loboo34/collab-api/models"
	"github.com/Loboo34/collab-api/utils"
	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only POST Allowed", "")
		return
	}

	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "User ID not found", "")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid json format", "")
		return
	}

	teamCollection := database.DB.Collection("teams")

	team := models.Team{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		Description: req.Description,
		Members:     []string{userID},
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		Projects:    []string{},
	}

	membersCollection := database.DB.Collection("team-members")

	members := models.TeamMember{
		ID:       primitive.NewObjectID(),
		TeamId:   team.ID,
		User:     userID,
		Role:     "Admin",
		JoinedAt: time.Now(),
	}

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
		utils.RespondWithError(w, http.StatusInternalServerError, "Error saving Admin", "")
	}

	_, err = userCollection.UpdateOne(ctx, bson.M{"_id": userObjID}, bson.M{"$addToSet": bson.M{"teams": team.ID}})

	utils.Logger.Info("Team created successfully")
	utils.RespondWithJSON(w, http.StatusCreated, "Team created Successfully", map[string]interface{}{"team_id": team.ID.Hex(),
		"name": team.Name})
}

func UpdateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only PUT Allowed", "")
		return
	}

	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User ID", "")
		return
	}

	vars := mux.Vars(r)
	teamID := vars["teamId"]
	if teamID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing Team ID", "")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid json format", "")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teamCollection := database.DB.Collection("teams")
	var team models.Team

	err = teamCollection.FindOne(ctx, bson.M{"_id": teamID}).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Team not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding team", "")
		}
		return
	}

	memberCollection := database.DB.Collection("team-members")
	var member models.TeamMember

	err = memberCollection.FindOne(ctx, bson.M{"user": userID, "teamId": teamID}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Member not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding member", "")
		}
		return
	}

	update := bson.M{
		"$set": bson.M{
			"name":        req.Name,
			"description": req.Description,
		},
	}

	result, err := teamCollection.UpdateOne(ctx, bson.M{"_id": teamID}, update)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error updating Team", "")
		return
	}

	if result.MatchedCount == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Team not found", "")
		return
	}

	utils.Logger.Info("Team Updated")
	utils.RespondWithJSON(w, http.StatusOK, "Update Seccessful", map[string]interface{}{"team": update})

}

func InviteMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only POST Allowed", "")
		return
	}

	userId, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "User ID not found", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type InveteRequest struct {
		Email  string `json:"email"`
		TeamId string `json:"teamId"`
	}

	var req InveteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid json format", "")
		return
	}

	teamObjId, err := primitive.ObjectIDFromHex(req.TeamId)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Team ID", "")
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

	memberCollection := database.DB.Collection("team-members")
	var member models.TeamMember

	err = memberCollection.FindOne(ctx, bson.M{"user": userId, "teamId": teamObjId, "role": "Admin"}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Member not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding member", "")
		}
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

	var existingMember models.TeamMember
	err = memberCollection.FindOne(ctx, bson.M{"user": user.ID.Hex(), "teamId": teamObjId}).Decode(&existingMember)
	if err == nil {
		utils.RespondWithError(w, http.StatusConflict, "User Already exists in team", "")
		return
	}

	inviteCollection := database.DB.Collection("invites")
	var existingInvite models.Invite

	err = inviteCollection.FindOne(ctx, bson.M{"email": user.Email, "teamId": teamObjId, "status": "pending"}).Decode(&existingInvite)
	if err == nil {
		utils.RespondWithError(w, http.StatusConflict, "Invite already exists", "")
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

	_, err = inviteCollection.InsertOne(ctx, invite)
	if err != nil {
		utils.Logger.Warn("Failed to create invite")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error creating invite", "")
		return
	}

	inviteLink := "http://localhost:3000/invite/accept?token=" + inviteToken

	if err := utils.SendInviteEmail(user.Email, inviteLink); err != nil {
		utils.Logger.Warn("Failed to send email")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error sending email", "")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, "Invitation sent successfully", map[string]interface{}{
		"email":     user.Email,
		"team_id":   teamObjId.Hex(),
		"team_name": team.Name,
		"sent_by":   userId,
	})
}

func AcceptInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only POST Allowed", "")
		return
	}

	inviteToken := r.URL.Query().Get("token")
	if inviteToken == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing invite token", "")
		return
	}

	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User ID", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inviteCollection := database.DB.Collection("invites")
	var invite models.Invite
	err = inviteCollection.FindOne(ctx, bson.M{"token": inviteToken}).Decode(&invite)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Invalid or expired invite", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding invite", "")
		}
		return
	}

	if invite.Status != "pending" {
		utils.RespondWithError(w, http.StatusConflict, "Invite already processed", "")
		return
	}

	userCollection := database.DB.Collection("users")
	userObjID, _ := primitive.ObjectIDFromHex(userID)
	var user models.User
	err = userCollection.FindOne(ctx, bson.M{"_id": userObjID}).Decode(&user)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error finding user", "")
		return
	}

	if user.Email != invite.Email {
		utils.RespondWithError(w, http.StatusForbidden, "This invite is for a different user", "")
		return
	}

	membersCollection := database.DB.Collection("team-members")
	newMember := models.TeamMember{
		ID:       primitive.NewObjectID(),
		TeamId:   invite.TeamID,
		User:     userID,
		Role:     "Member",
		JoinedAt: time.Now(),
	}

	_, err = membersCollection.InsertOne(ctx, newMember)
	if err != nil {
		utils.Logger.Warn("Failed to Add user to team")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error adding member to team", "")
		return
	}

	_, err = userCollection.UpdateOne(
		ctx,
		bson.M{"_id": userObjID},
		bson.M{"$addToSet": bson.M{"teams": invite.TeamID}},
	)
	if err != nil {
		utils.Logger.Warn("Failed to update user teams")
	}

	teamCollection := database.DB.Collection("teams")
	_, err = teamCollection.UpdateOne(
		ctx,
		bson.M{"_id": invite.TeamID},
		bson.M{"$addToSet": bson.M{"members": userID}},
	)
	if err != nil {
		utils.Logger.Warn("Failed to update team members")
	}

	_, err = inviteCollection.UpdateOne(
		ctx,
		bson.M{"_id": invite.ID},
		bson.M{"$set": bson.M{"status": "accepted"}},
	)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error updating invite status", "")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Invite accepted successfully", map[string]interface{}{
		"team_id": invite.TeamID.Hex(),
		"user_id": userID,
	})
}

func DeclineInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only POST Allowed", "")
		return
	}

	inviteToken := r.URL.Query().Get("token")
	if inviteToken == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing invite token", "")
		return
	}

	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Missing User ID", "")
		return
	}

	id, _ := primitive.ObjectIDFromHex(userID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inviteCollection := database.DB.Collection("invites")
	var invite models.Invite

	err = inviteCollection.FindOne(ctx, bson.M{"token": inviteToken}).Decode(&invite)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Invalid or expired invite", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding invite", "")
		}
		return
	}

	if invite.Status != "pending" {
		utils.RespondWithError(w, http.StatusConflict, "Invite already processed", "")
		return
	}

	_, err = inviteCollection.UpdateOne(ctx, bson.M{"_id": invite.ID}, bson.M{"$set": bson.M{"status": "declined"}})
	if err != nil {
		utils.Logger.Warn("Failed to decline Invitation")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error declinign invitation", "")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Invite declined", map[string]interface{}{"user": id})
}

func GetTeamMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only GET Allowed", "")
		return
	}

	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User ID", "")
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	if teamIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing Team ID", "")
		return
	}

	teamID, err := primitive.ObjectIDFromHex(teamIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Team ID", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teamCollection := database.DB.Collection("teams")
	var team models.Team

	err = teamCollection.FindOne(ctx, bson.M{"_id": teamID}).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Team not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding Team", "")
		}
		return
	}

	membersCollection := database.DB.Collection("team-members")
	var userMember models.TeamMember

	err = membersCollection.FindOne(ctx, bson.M{"user": userID, "teamId": teamID}).Decode(&userMember)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Member not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding member ", "")
		}
		return
	}

	cursor, err := membersCollection.Find(ctx, bson.M{"teamId": teamID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching ", "")
		return
	}

	defer cursor.Close(ctx)
	var members []models.TeamMember
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

func ChangeRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only PUT Allowed", "")
		return
	}

	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User ID", "")
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	if teamIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing Team ID", "")
		return
	}

	teamID, err := primitive.ObjectIDFromHex(teamIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Team ID", "")
		return
	}

	var body struct {
		MemberID string `json:"memberId"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid json format", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teamCollection := database.DB.Collection("teams")
	var team models.Team
	err = teamCollection.FindOne(ctx, bson.M{"_id": teamID}).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Team not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding Team", "")
		}
		return
	}

	membersCollection := database.DB.Collection("team-members")
	var member models.TeamMember
	err = membersCollection.FindOne(ctx, bson.M{"user": body.MemberID, "teamId": teamID}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Member not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding Member", "")
		}
		return
	}

	result, err := membersCollection.UpdateOne(ctx, bson.M{"teamId": teamID, "user": body.MemberID}, bson.M{"$set": bson.M{"role": body.Role}})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Faile to update role", "")
		return
	}

	if result.MatchedCount == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Failed to find member", "")
		return
	}

	utils.Log(
		userID,
		teamIDStr,
		"",
		"",
		"Changed Role",
		userID+"Changed'"+body.MemberID+"' role to "+body.Role,
	)

	utils.Logger.Info("Changed users role successfully")
	utils.RespondWithJSON(w, http.StatusOK, "Role changed successfuly", map[string]interface{}{"user": member.ID, "role": body.Role})
}

func GetTeams(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only GET Allowed", "")
		return
	}

	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User ID", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teamsCollection := database.DB.Collection("teams")
	cursor, err := teamsCollection.Find(ctx, bson.M{"members": userID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error finding teams", "")
		return
	}
	defer cursor.Close(ctx)

	var teams []models.Team
	for cursor.Next(ctx) {
		var team models.Team
		if err := cursor.Decode(&team); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error decoding team", "")
			return
		}
		teams = append(teams, team)
	}

	if err = cursor.Err(); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Cursor error", "")
		return
	}

	if len(teams) == 0 {
		utils.RespondWithJSON(w, http.StatusOK, "No teams found", map[string]interface{}{
			"teams": []models.Team{},
			"count": 0,
		})
		return
	}

	utils.Logger.Info("Fetched user teams successfully")
	utils.RespondWithJSON(w, http.StatusOK, "Teams retrieved successfully", map[string]interface{}{
		"teams": teams,
		"count": len(teams),
	})
}

func RemoveMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Delete Allowed", "")
		return
	}

	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User ID", "")
		return
	}

	var request struct {
		User string `json:"user"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid json format", "")
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	if teamIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Error getting Team ID", "")
		return
	}

	teamID, err := primitive.ObjectIDFromHex(teamIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Team ID", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teamCollection := database.DB.Collection("teams")
	var team models.Team

	err = teamCollection.FindOne(ctx, bson.M{"_id": teamID}).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Team not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding Team", "")
		}
		return
	}

	membersCollection := database.DB.Collection("team-members")
	var member models.TeamMember

	err = membersCollection.FindOne(ctx, bson.M{"user": request.User, "teamId": teamID}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Member not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding Member", "")
		}
		return
	}
	result, err := membersCollection.DeleteOne(ctx, bson.M{"user": request.User, "teamId": teamID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error Removing user", "")
		return
	}

	if result.DeletedCount == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Member not found", "")
		return
	}

	_, err = teamCollection.UpdateOne(
		ctx,
		bson.M{"_id": teamID},
		bson.M{"$pull": bson.M{"members": request.User}},
	)
	if err != nil {
		utils.Logger.Warn("Failed to update team members array")

	}

	userCollection := database.DB.Collection("users")
	userObjID, _ := primitive.ObjectIDFromHex(request.User)
	_, err = userCollection.UpdateOne(
		ctx,
		bson.M{"_id": userObjID},
		bson.M{"$pull": bson.M{"teams": teamID}},
	)
	if err != nil {
		utils.Logger.Warn("Failed to update user's teams array")

	}

	utils.Log(
		userID,
		teamIDStr,
		"",
		"",
		"Removed Member",
		userID+"Removed '"+request.User+"' from team",
	)

	utils.Logger.Info("User successfully removed from team")
	utils.RespondWithJSON(w, http.StatusOK, "Member removed successfully", map[string]interface{}{
		"team_id":   teamID.Hex(),
		"member_id": request.User,
	})
}

func DeleteTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Delete Allowed", "")
		return
	}

	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User ID", "")
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	if teamIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing Team ID", "")
		return
	}

	teamID, err := primitive.ObjectIDFromHex(teamIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Team ID", "")
		return
	}

	var team models.Team
	teamCollection := database.DB.Collection("teams")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = teamCollection.FindOne(ctx, bson.M{"_id": teamID}).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Team not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding Team", "")
		}
		return
	}

	result, err := teamCollection.DeleteOne(ctx, bson.M{"_id": teamID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete Team", "")
		return
	}

	if result.DeletedCount == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Team Not found", "")
		return
	}

	membersCollection := database.DB.Collection("team-members")
	_, err = membersCollection.DeleteMany(ctx, bson.M{"teamId": teamID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failled to dalete team members", "")
		return
	}

	inviteCollection := database.DB.Collection("invites")
	_, err = inviteCollection.DeleteMany(ctx, bson.M{"teamId": teamID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "failed to delete invitations", "")
		return
	}

	utils.Logger.Info("Deleted Team")
	utils.RespondWithJSON(w, http.StatusOK, "Team successfuly deleted", map[string]interface{}{"Team deleted by": userID, "team": team})
}
