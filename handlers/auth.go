package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Loboo34/collab-api/database"
	"github.com/Loboo34/collab-api/models"
	"github.com/Loboo34/collab-api/utils"

	"go.mongodb.org/mongo-driver/bson"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "ONly Post Allowed", "")
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.Logger.Warn("Invalid SJon")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON format", "")
		return
	}

	hashed, err := utils.HashPassword(user.Password)
	if err != nil {
		utils.Logger.Warn("Failed to hash Passwordf")
		utils.RespondWithError(w, http.StatusBadRequest, "Error Hashing password", "")
		return
	}
	user.Password = hashed

	collection := database.DB.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		utils.Logger.Warn("Failed to Register User")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error while regestering new user", "")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated,"Regestration successfull", map[string]string{"message": "User registered Successfuly"})

}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "On;y Post Allowed", "")
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON format", "")
		return
	}

	collection := database.DB.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var exists models.User
	err := collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&exists)
	if err != nil {
		utils.Logger.Warn("User Not found")
		utils.RespondWithError(w, http.StatusInternalServerError, "Invalid credentials", "")
		return
	}

	if !utils.ComparePassword(user.Password, exists.Password){
		utils.Logger.Warn("Incorrect password")
		utils.RespondWithError(w, http.StatusInternalServerError, "Invalid Credentials", "")
		return
	}

	memberCollection := database.DB.Collection("team-members")
	var member models.TeamMember

	role := ""
	err = memberCollection.FindOne(ctx, bson.M{"user": exists.ID.Hex()}).Decode(&member)
	if err == nil {
		role = member.Role
	}

token, err := utils.GenerateJWT(exists.ID.Hex(),exists.Email, role)
if err != nil{
	utils.RespondWithError(w, http.StatusInternalServerError, "Failed to login", "")
	return
}

utils.RespondWithJSON(w, http.StatusOK, "Login Successfull", map[string]string{"token": token})


}
