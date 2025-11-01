package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Loboo34/collab-api/database"
	"github.com/Loboo34/collab-api/models"
	"github.com/Loboo34/collab-api/utils"
)

func AddTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Post Allowed", "")
		return
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Json format", "")
		return
	}

	task.ID = primitive.NewObjectID()
	task.CreatedAt = time.Now()

	collection := database.DB.Collection("tasks")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, task)
	if err != nil {
		utils.Logger.Warn("Failed to Add Task")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error adding task", "")
		return
	}
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

	utils.RespondWithJSON(w, http.StatusCreated,"", map[string]string{"message": "Update successful"})
}
