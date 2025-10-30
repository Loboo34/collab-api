package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/Loboo34/collab-api/handlers"
	"github.com/Loboo34/collab-api/database"
	"github.com/Loboo34/collab-api/middleware"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load env")
	}
}

func main() {
	r := mux.NewRouter()

	r.Use(middleware.Cors())

	db := database.ConnectDB()
	fmt.Println("DbName:", db.Name())

	//handlers
	//auth
	r.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	r.HandleFunc("/login", handlers.LoginUser).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":"+port, r))
}
