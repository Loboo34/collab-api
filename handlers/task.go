package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Loboo34/collab-api/database"
	"github.com/Loboo34/collab-api/models"
	"github.com/Loboo34/collab-api/utils"
)

func CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Post Allowed", "")
		return
	}

	tokenstring := r.Header.Get("Authorization")
	if tokenstring == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "missing Auth token", "")
		return
	}

	tokenstring = strings.TrimPrefix(tokenstring, "Bearer ")

	claims, err := utils.ValidateJWT(tokenstring)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token string", "")
		return
	}

	userId, ok := claims["id"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User id not found", "")
		return
	}

	var request struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		TeamID      string `json:"teamId"`
		ProjectID   string `json:"projectId"`
	}

	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid json format", "")
		return
	}

	teamID, err := primitive.ObjectIDFromHex(request.TeamID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid teamID", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var team models.Team
	teamCollection := database.DB.Collection("teams")
	err = teamCollection.FindOne(ctx, bson.M{"_id": teamID}).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Team Not Found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error Finding team", "")
		}
		return
	}

	projectID, err := primitive.ObjectIDFromHex(request.ProjectID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid teamID", "")
		return
	}

	var project models.Project
	projectCollection := database.DB.Collection("projects")

	err = projectCollection.FindOne(ctx, bson.M{"_id": projectID}).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Project Not Found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error Finding project", "")
		}
		return
	}

	task := models.Task{
		ID:          primitive.NewObjectID(),
		Title:       request.Title,
		Description: request.Description,
		Status:      "Pending",
		TeamId:      teamID,
		ProjectId:   projectID,
		CreatedBy:   userId,
		CreatedAt:   time.Now(),
	}

	taskCollection := database.DB.Collection("tasks")

	_, err = taskCollection.InsertOne(ctx, task)
	if err != nil {
		utils.Logger.Warn("Failed to Add Task")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error adding task", "")
		return
	}

	_, err = projectCollection.UpdateOne(ctx, bson.M{"_id": projectID}, bson.M{"$addToSet": bson.M{"projects": task.ID}})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error Adding task to project", "")
		return
	}

	utils.Logger.Info("Task created successfully")
	utils.RespondWithJSON(w, http.StatusCreated, "Task added successfully", map[string]interface{}{"user": userId})
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
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
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid Token string", "")
		return
	}

	userID, ok := claims["id"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User ID", "")
		return
	}

	role, ok := claims["role"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User role", "")
		return
	}

	vars := mux.Vars(r)
	taskIDStr := vars["taskId"]
	if taskIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing task Id", "")
		return
	}

	taskID, err := primitive.ObjectIDFromHex(taskIDStr)
	if err != nil {
		utils.Logger.Warn("Invalid id")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid id format", "")
		return
	}

	var updates struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err = json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid json", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	taskCollection := database.DB.Collection("tasks")
	var task models.Task
	err = taskCollection.FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Task not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding task", "")
		}
		return
	}

	update := bson.M{
		"$set": bson.M{
			"title":       updates.Title,
			"description": updates.Description,
		},
	}

	memberCollection := database.DB.Collection("team-members")
	var member models.TeamMember

	err = memberCollection.FindOne(ctx, bson.M{"user": userID, "teamId": task.TeamId}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Member not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding member", "")
		}
		return
	}

	if !strings.EqualFold(role, "Admin") && task.CreatedBy != userID {
		utils.RespondWithError(w, http.StatusForbidden, "User Not allowed to perform action", "")
		return
	}

	result, err := taskCollection.UpdateOne(ctx, bson.M{"_id": taskID}, update)
	if err != nil {
		utils.Logger.Warn("Failed to update task")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error updating Task", "")
		return
	}

	if result.MatchedCount == 0 {
		utils.Logger.Warn("Failed to find task")
		utils.RespondWithError(w, http.StatusNotFound, "Task not found", "")
		return
	}

	utils.Logger.Info("Task updated")
	utils.RespondWithJSON(w, http.StatusOK, "Update successful", map[string]interface{}{"taskID": task})
}

func AssignTo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Post Allowed", "")
		return
	}

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "Mising token", "")
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token", "")
		return
	}

	_, ok := claims["id"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User ID", "")
		return
	}

	vars := mux.Vars(r)
	taskIDStr := vars["taskId"]
	if taskIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing task ID", "")
		return
	}

	taskID, err := primitive.ObjectIDFromHex(taskIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid task id", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	taskCollection := database.DB.Collection("tasks")
	var task models.Task

	err = taskCollection.FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Task not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding task", "")
		}
		return
	}

	var body struct {
		AssignedTo string `json:"assignedTo"`
	}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid json format", "")
		return
	}

	memberCollection := database.DB.Collection("team-members")
	var member models.TeamMember

	err = memberCollection.FindOne(ctx, bson.M{"user": body.AssignedTo, "team": task.TeamId}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Member not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding team member", "")
		}
		return
	}

	assignedTOID, err := primitive.ObjectIDFromHex(body.AssignedTo)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid user ID", "")
		return
	}

	update := bson.M{"$set": bson.M{"assignedTo": assignedTOID}}
	result, err := database.DB.Collection("tasks").UpdateOne(ctx, bson.M{"_id": taskID}, update)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error assining task", "")
		return
	}

	if result.MatchedCount == 0 {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to find task", "")
		return
	}
}

func Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Put Allowed", "")
		return
	}

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing auth token", "")
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token string", "")
		return
	}

	userIDStr, ok := claims["id"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing user ID", "")
		return
	}

	vars := mux.Vars(r)
	taskIDStr := vars["taskId"]
	if taskIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing task ID", "")
		return
	}

	taskID, err := primitive.ObjectIDFromHex(taskIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid taskID", "")
		return
	}

	var body struct {
		Status string `json:"status"`
	}

	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Json Format", "")
		return
	}

	validStatuses := map[string]bool{"pending": true, "inProgress": true, "done": true}
	if !validStatuses[body.Status] {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid status. Must be: pending, inProgress, or done", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	taskCollection := database.DB.Collection("tasks")
	var task models.Task

	err = taskCollection.FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Task not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding task", "")
		}
		return
	}

	teamMemberCollection := database.DB.Collection("team-members")
	var member models.TeamMember

	err = teamMemberCollection.FindOne(ctx, bson.M{"user": userIDStr, "team": task.TeamId}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Member not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding member", "")
		}
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid user ID", "")
		return
	}
	if task.AssignedTo != userID {
		utils.RespondWithError(w, http.StatusBadRequest, "Task is not assigned to user", "")
		return
	}

	result, err := taskCollection.UpdateOne(ctx, bson.M{"_id": taskID}, bson.M{"$set": bson.M{"status": body.Status}})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error updating task status", "")
		return
	}

	if result.MatchedCount == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Error finding task", "")
		return
	}

	utils.Logger.Info("Task Status Updated Successfuly")
	utils.RespondWithJSON(w, http.StatusOK, "Status Update successfully", map[string]interface{}{
		"taskID": taskIDStr,
		"status": body.Status,
	})

}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Delete Allowed", "")
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
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid Auth token", "")
		return
	}

	userID, ok := claims["id"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User ID", "")
		return
	}

	role, ok := claims["role"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User role", "")
		return
	}

	vars := mux.Vars(r)
	taskIDStr := vars["taskId"]
	if taskIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing task ID", "")
		return
	}
	taskID, _ := primitive.ObjectIDFromHex(taskIDStr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	taskCollection := database.DB.Collection("tasks")
	var task models.Task

	err = taskCollection.FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Task not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding task", "")
		}
		return
	}

	if task.CreatedBy != userID && role != "Admin" {
		utils.RespondWithError(w, http.StatusForbidden, "Not Permited to perform action", "")
		return
	}

	result, err := taskCollection.DeleteOne(ctx, bson.M{"_id": taskID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error While deliting task", "")
		return
	}

	if result.DeletedCount == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Error finding task", "")
		return
	}

	utils.Logger.Info("Task deleted successfuly")
	utils.RespondWithJSON(w, http.StatusOK, "Task Deleted", "")
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not Allowed", "")
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
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token string", "")
		return
	}

	userID := claims["id"].(string)

	vars := mux.Vars(r)
	projectIDStr := vars["projectId"]

	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid project ID", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	memberCollection := database.DB.Collection("team-members")
	var member models.TeamMember

	err = memberCollection.FindOne(ctx, bson.M{"user": userID}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Member not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding member", "")
		}
		return
	}

	taskCollection := database.DB.Collection("tasks")
	cursor, err := taskCollection.Find(ctx, bson.M{"projectId": projectID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching tasks", "")
		return
	}

	defer cursor.Close(ctx)

	var tasks []models.Task
	for cursor.Next(ctx) {
		var task models.Task
		if err := cursor.Decode(&task); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error decoding task", "")
			return
		}
		tasks = append(tasks, task)
	}

	if err = cursor.Err(); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Cursor error", "")
		return
	}

	utils.Logger.Info("Fetched team projects successfully")
	utils.RespondWithJSON(w, http.StatusOK, "Projects retrieved", map[string]interface{}{
		"project_id": projectID.Hex(),
		"tasks":      tasks,
	})

}

func GetTask(w http.ResponseWriter, r *http.Request) {
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

	userID := claims["id"].(string)

	vars := mux.Vars(r)
	taskIDStr := vars["taskId"]
	if taskIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing Task ID", "")
		return
	}

	taskID, err := primitive.ObjectIDFromHex(taskIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid task ID", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	memberCollection := database.DB.Collection("team-members")
	var member models.TeamMember

	err = memberCollection.FindOne(ctx, bson.M{"user": userID}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "member not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding member", "")
		}
		return
	}

	if userID != member.User {
		utils.RespondWithError(w, http.StatusForbidden, "User must be part of team", "")
		return
	}

	taskCollection := database.DB.Collection("tasks")
	var task models.Task

	if member.TeamId != task.TeamId {
		utils.RespondWithError(w, http.StatusForbidden, "You are not a member of this task's team", "")
		return
	}

	err = taskCollection.FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusNotFound, "Task not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding task", "")
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Task fetched", map[string]interface{}{"task": task})

}
