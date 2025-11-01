package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	//"github.com/gorilla/mux"
	//"go.mongodb.org/mongo-driver/bson"
	"github.com/Loboo34/collab-api/database"
	"github.com/Loboo34/collab-api/models"
	"github.com/Loboo34/collab-api/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Post Allowed", "")
		return
	}
	var project models.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Json", "")
		return
	}

	project.ID = primitive.NewObjectID()
	project.CreatedAt = time.Now()

	collection := database.DB.Collection("projects")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, project)
	if err != nil {
		utils.Logger.Warn("Failed to add project")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error failing to add project", "")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, "", map[string]string{"message": "Successfully added project"})
}
