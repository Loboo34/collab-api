package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Loboo34/collab-api/database"
	"github.com/Loboo34/collab-api/models"
	"github.com/Loboo34/collab-api/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "ONly Post Allowed", "")
		return
	}

	var req struct {
		FullName string `json:"fullname"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	fmt.Println(req)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Logger.Warn("Invalid SJon")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON format", "")
		return
	}

	hashedPass, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Logger.Warn("Failed to hash Passwordf")
		utils.RespondWithError(w, http.StatusBadRequest, "Error Hashing password", "")
		return
	}

	newUser := models.User{
		ID:        primitive.NewObjectID(),
		FullName:  req.FullName,
		Email:     req.Email,
		Password:  hashedPass,
		CreatedAt: time.Now(),
		Teams:     []primitive.ObjectID{},
	}
fmt.Println(newUser)
	collection := database.DB.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, newUser)
	if err != nil {
		utils.Logger.Warn("Failed to Register User")
		utils.RespondWithError(w, http.StatusInternalServerError, "Error while regestering new user", "")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, "Regestration successfull", map[string]interface{}{"message": "User registered Successfuly"})

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

	if !utils.ComparePassword(user.Password, exists.Password) {
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

	token, err := utils.GenerateJWT(exists.ID.Hex(), exists.Email, role)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to login", "")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Login Successfull", map[string]string{"token": token})

}


func Profile(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Only Get Allowed", "")
		return
	}

	userID, err := utils.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Missing User ID", "",)
		return
	}

	userCollection := database.DB.Collection("users")
	var user models.User


	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()


	err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments{
			utils.RespondWithError(w, http.StatusNotFound, "User not found", "")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error finding user", "")
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "User fetched", map[string]interface{}{"user": user})
}