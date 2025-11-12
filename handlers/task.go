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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var team models.Team
	teamCollection := database.DB.Collection("teams")
	err = teamCollection.FindOne(ctx, bson.M{"_id": team.ID}).Decode(&team)

	var project models.Project
	projectCollection := database.DB.Collection("projects")

	err = projectCollection.FindOne(ctx, bson.M{"_id": project.ID}).Decode(&project)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Error fetching project", "")
		return
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Json format", "")
		return
	}

	task.ID = primitive.NewObjectID()
	task.CreatedAt = time.Now()
	task.TeamId = team.ID
	task.ProjectId = project.ID

	taskCollection := database.DB.Collection("tasks")

	_, err = taskCollection.InsertOne(ctx, task)
	if err != nil {
		utils.Logger.Warn("Failed to Add Task")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error adding task", "")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, "Task added successfully", map[string]interface{}{"user": userId})
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only PUT Allowed", "")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing Id", "")
		return
	}

	objectId, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.Logger.Warn("Invalid id")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid id format", "")
		return
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.Logger.Warn("Invalid Json format")
		utils.RespondWithError(w, http.StatusBadRequest, "INvalid json", "")
		return
	}

	collection := database.DB.Collection("task")

	update := bson.M{
		"$set": bson.M{
			"title":       task.Title,
			"description": task.Description,
			"assignedTo":  task.AssignedTo,
			"updated":     time.Now(),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectId}, update)
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

	utils.RespondWithJSON(w, http.StatusCreated, "", map[string]string{"message": "Update successful"})
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

	taskIDStr := r.URL.Query().Get("taskId")
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
		utils.RespondWithError(w, http.StatusNotFound, "Error finding task", "")
		return
	}

	var body struct {
		AssignedTo string `json:"assignedto"`
	}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid json format", "")
		return
	}

	assignedTo, _ := primitive.ObjectIDFromHex(body.AssignedTo)

	memberCollection := database.DB.Collection("team-members")
	var member models.TeamMember

	err = memberCollection.FindOne(ctx, bson.M{"user": assignedTo}).Decode(&member)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Error finding member", "")
		return
	}

	update := bson.M{"set": bson.M{"assignedTo": assignedTo}}
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

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid user ID", "")
		return
	}

	taskIDStr := r.URL.Query().Get("taskId")
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
		TaskID string `json:"taskID"`
		Status string `json:"status"`
	}

	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Json Format", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teamMemberCollection := database.DB.Collection("team-members")
	var member models.TeamMember

	err = teamMemberCollection.FindOne(ctx, bson.M{"user": userID}).Decode(&member)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Error finding member", "")
		return
	}

	taskCollection := database.DB.Collection("tasks")
	var task models.Task

	err = taskCollection.FindOne(ctx, bson.M{"taskId": body.TaskID}).Decode(&task)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Error finding task", "")
	}
	err = taskCollection.FindOne(ctx, bson.M{"assignedTo": userID}).Decode(&task)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Error finding task", "")
	}

	if task.AssignedTo != userID {
		utils.RespondWithError(w, http.StatusBadRequest, "Task is not assigned to user", "")
		return
	}

	result, err := taskCollection.UpdateOne(ctx, bson.M{"taskId": taskID}, bson.M{"$set": bson.M{"status": body.Status}})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error updating task status", "")
		return
	}

	if result.MatchedCount == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Error finding task", "")
		return
	}

	utils.Logger.Info("Task Status Updated Successfuly")
	utils.RespondWithJSON(w, http.StatusOK, "Update successful", map[string]interface{}{
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

	tokenString = strings.TrimPrefix(tokenString, "Bearer")

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

	taskIDStr := r.URL.Query().Get("taskId")
	if taskIDStr == ""{
		utils.RespondWithError(w, http.StatusBadRequest, "Missing task ID", "")
		return
	}
	taskID,_ := primitive.ObjectIDFromHex(taskIDStr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	taskCollection := database.DB.Collection("tasks")
	var task models.Task

	err = taskCollection.FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err != nil{
		utils.RespondWithError(w, http.StatusBadRequest, "Missing task", "")
		return
	}

	if task.CreatedBy != userID && role != "Admin"{
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

func GetTasks(w http.ResponseWriter, r *http.Request){
if r.Method != http.MethodGet {
	utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not Allowed", "")
	return
}

tokenString := r.Header.Get("Authorization")
if tokenString == ""{
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

projectIDStr := r.URL.Query().Get("projectId")

projectID, err := primitive.ObjectIDFromHex(projectIDStr)
if err != nil {
	utils.RespondWithError(w, http.StatusBadRequest, "Invalid project ID", "")
	return
}

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

memberCollection := database.DB.Collection("team-member")
var member models.TeamMember

err = memberCollection.FindOne(ctx, bson.M{"user": userID}).Decode(&member)
if err != nil {
	utils.RespondWithError(w, http.StatusBadRequest, "User is not on team", "")
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
for cursor.Next(ctx){
	var task models.Task
	if err := cursor.Decode(&task); err != nil{
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
        "project_id":  projectID.Hex(),
        "tasks": tasks,
    })

}


func getTask(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only get Allowed", "")
		return
	}

	tokenString := r.Header.Get("Authorization")
	if tokenString == ""{
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing Auth Token", "")
		return
	}

	tokenString = strings.TrimPrefix(tokenString, "Beare ")

	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid Token string", "")
		return
	}

	userID := claims["id"].(string)

	
}