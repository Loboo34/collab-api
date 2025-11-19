package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	//"github.com/gorilla/mux"
	//"go.mongodb.org/mongo-driver/bson"
	"github.com/Loboo34/collab-api/database"
	"github.com/Loboo34/collab-api/models"
	"github.com/Loboo34/collab-api/utils"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Post Allowed", "")
		return
	}

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing Auth Token", "")
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claim, err := utils.ValidateJWT(tokenString)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", "")
		return
	}

	userID, ok := claim["id"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User ID not found", "")
		return
	}

	userRole, ok := claim["role"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User role not found", "")
		return
	}

	if !strings.EqualFold(userRole, "Admin") {
		utils.RespondWithError(w, http.StatusUnauthorized, "User is not Admin", "")
		return
	}

	var request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		//TeamId      string `json:"teamId"`
		//CreatedBy   string `json:"createdBy"`
	}

	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invlaid Json", "")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	if teamIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing team id", "")
		return
	}

	teamID, err := primitive.ObjectIDFromHex(teamIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Team id", "")
		return
	}
	var team models.Team

	teamCollection := database.DB.Collection("teams")
	err = teamCollection.FindOne(ctx, bson.M{"_id": teamID}).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Team not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding Team", "")
		}
		return
	}

	project := models.Project{
		ID:          primitive.NewObjectID(),
		Name:        request.Name,
		Description: request.Description,
		TeamId:      teamID,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		Tasks:       []string{},
	}

	projectCollection := database.DB.Collection("projects")
	_, err = projectCollection.InsertOne(ctx, project)
	if err != nil {
		utils.Logger.Warn("Failed to create project")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error creating project", "")
		return
	}

	_, err = teamCollection.UpdateOne(ctx, bson.M{"_id": teamID}, bson.M{"$addToSet": bson.M{"projects": project.ID}})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error Adding project to teams", "")
		return
	}

	utils.Logger.Info("Project Created Successfuly")
	utils.RespondWithJSON(w, http.StatusCreated, "Project created successful", map[string]string{"projectID": project.ID.Hex(), "name": project.Name})
}

func UpdateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only PUT Allowed", "")
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
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid Token", "")
		return
	}

	userID, ok := claims["id"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing user ID", "")
		return
	}

	role, ok := claims["role"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing Role", "")
		return
	}

	vars := mux.Vars(r)

	projectIDStr := vars["projectId"]
	if projectIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing project id", "")
		return
	}

	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Project id", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	projectCollection := database.DB.Collection("projects")
	var project models.Project

	err = projectCollection.FindOne(ctx, bson.M{"_id": projectID}).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Project Not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding project", "")
		}
		return
	}

	if !strings.EqualFold(role, "Admin") && userID != project.CreatedBy {
		utils.RespondWithError(w, http.StatusForbidden, "User Not authorized to perform action", "")
		return
	}

	var updates struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err = json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid json format", "")
		return
	}

	update := bson.M{
		"$set": bson.M{
			"name":        updates.Name,
			"description": updates.Description,
		},
	}

	result, err := projectCollection.UpdateOne(ctx, bson.M{"_id": projectID}, update)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error updating Project", "")
		return
	}

	if result.MatchedCount == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Error finding project", "")
		return
	}

	utils.Logger.Info("Project updated successfuly")
	utils.RespondWithJSON(w, http.StatusOK, "Project updated", map[string]interface{}{"projectID": project.ID, "name": updates.Name})
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Delete Allowed", "")
		return
	}

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing Auth Token", "")
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Ivalid Token", "")
		return
	}

	userID, ok := claims["id"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing id", "")
		return
	}

	role, ok := claims["role"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User Role", "")
		return
	}

	if !strings.EqualFold(role, "Admin") {
		utils.RespondWithError(w, http.StatusForbidden, "User is not admin", "")
		return
	}

	vars := mux.Vars(r)
	projectIDStr := vars["projectId"]
	if projectIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "mising project id", "")
		return
	}

	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid project id", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	projectCollection := database.DB.Collection("projects")
	var project models.Project

	err = projectCollection.FindOne(ctx, bson.M{"_id": projectID}).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Project Not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding project", "")
		}
		return
	}

	if project.CreatedBy != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Only Project creator can Perform action", "")
		return
	}

	result, err := projectCollection.DeleteOne(ctx, bson.M{"_id": projectID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error while deleting project", "")
		return
	}

	if result.DeletedCount == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Project not Found", "")
		return
	}

	taskCollection := database.DB.Collection("tasks")
	_, err = taskCollection.DeleteMany(ctx, bson.M{"projectId": projectID})
	if err != nil {
		utils.Logger.Warn("Failed to delete project tasks")
		return

	}

	teamCollection := database.DB.Collection("teams")
	_, err = teamCollection.UpdateOne(
		ctx,
		bson.M{"_id": project.TeamId},
		bson.M{"$pull": bson.M{"projects": projectID}},
	)
	if err != nil {
		utils.Logger.Warn("Failed to update team's projects array")
		return
	}

	utils.Logger.Info("Project deleted successfuly")
	utils.RespondWithError(w, http.StatusOK, "Project deleted", map[string]interface{}{"Project": projectID, "user": userID})
}

func GetProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Get Allowe", "")
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
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid toke", "")
		return
	}

	userID, ok := claims["id"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing user Id", "")
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	if teamIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing team id", "")
		return
	}

	teamID, err := primitive.ObjectIDFromHex(teamIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Team id", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

	teamCollection := database.DB.Collection("teams")
	var team models.Team

	err = teamCollection.FindOne(ctx, bson.M{"_id": teamID}).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Team Not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding team", "")
		}
		return
	}

	projectCollection := database.DB.Collection("projects")
	cursor, err := projectCollection.Find(ctx, bson.M{"teamId": teamID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching projets", "")
		return
	}
	defer cursor.Close(ctx)
	var projects []models.Project
	for cursor.Next(ctx) {
		var project models.Project
		if err := cursor.Decode(&project); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error decoding project", "")
			return
		}
		projects = append(projects, project)
	}

	if err = cursor.Err(); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Cursor error", "")
		return
	}

	utils.Logger.Info("Fetched team projects successfully")
	utils.RespondWithJSON(w, http.StatusOK, "Projects retrieved", map[string]interface{}{
		"team_id":  teamID.Hex(),
		"projects": projects,
		"count":    len(projects),
	})

}

func GetProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only get Allowed", "")
		return
	}

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing Auth Token", "")
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid Token string", "")
		return
	}

	userID, ok := claims["id"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing user ID", "")
		return
	}

	vars := mux.Vars(r)
	projectIDStr := vars["projectId"]
	if projectIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing Task ID", "")
		return
	}

	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid task ID", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	projectCollection := database.DB.Collection("projects")
	var project models.Project

	err = projectCollection.FindOne(ctx, bson.M{"_id": projectID}).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Project not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding project", "")
		}
		return
	}

	memberCollection := database.DB.Collection("team-members")
	var member models.TeamMember

	err = memberCollection.FindOne(ctx, bson.M{"user": userID, "teamId": project.TeamId}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Member not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding member", "")
		}
		return
	}

	taskCollection := database.DB.Collection("tasks")
	var task models.Task

	err = taskCollection.FindOne(ctx, bson.M{"_id": projectID}).Decode(&task)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching task", "")
		return
	}

	utils.Logger.Info("Project fetched successfullyt")
	utils.RespondWithJSON(w, http.StatusOK, "Task fetched", map[string]interface{}{"project": project})

}
