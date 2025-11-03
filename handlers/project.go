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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()


	var team models.Team
	//teamID, err := primitive.ObjectIDFromHex(team.ID)
	

	teamCollection := database.DB.Collection("teams")
	err = teamCollection.FindOne(ctx, bson.M{"_id": team.ID}).Decode(&team)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Team not found", "")
		return
	}

	var project models.Project
	if err = json.NewDecoder(r.Body).Decode(&project); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invlaid Json", "")
		return
	}

	project.ID = primitive.NewObjectID()
	project.TeamId = team.ID
	project.CreatedBy = userID
	project.CreatedAt = time.Now()
	projectCollection := database.DB.Collection("projects")

	_,err = projectCollection.InsertOne(ctx, project)
	if err != nil {
		utils.Logger.Warn("Failed to create project")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error creating project", "")
		return
	}



	

	utils.RespondWithJSON(w, http.StatusCreated, "", map[string]string{"message": "Successfully added project"})
}
