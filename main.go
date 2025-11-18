package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/Loboo34/collab-api/database"
	"github.com/Loboo34/collab-api/handlers"
	"github.com/Loboo34/collab-api/middleware"
	"github.com/Loboo34/collab-api/utils"
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
	utils.InitLogger()

	if err := utils.InitJWT(); err != nil {
		log.Fatal("Failed to initialize JWT:", err)
	}

	//handlers
	//auth
	r.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	r.HandleFunc("/login", handlers.LoginUser).Methods("POST")

	// teams
	r.HandleFunc("/team/create", handlers.CreateTeam).Methods("Post")
	r.HandleFunc("/team/invite", handlers.InviteMember).Methods("Post")
	r.HandleFunc("/invite/accept", handlers.AcceptInvite).Methods("Post")
	// r.HandleFunc("/invite/Decline", handlers.DeclineInvite).Methods("Post")
	r.HandleFunc("/team/{teamId}/members", handlers.GetTeamMembers).Methods("Get")
	r.HandleFunc("/team/{teamId}", handlers.DeleteTeam).Methods("Delete")
	r.HandleFunc("/team/{teamId}/remove", handlers.RemoveMember).Methods("Delete")

	// project
	r.HandleFunc("/project/create/{teamId}", handlers.CreateProject).Methods("Post")
	r.HandleFunc("/project/{projectId}/update", handlers.UpdateProject).Methods("Put")
	r.HandleFunc("/team/{teamId}/projects", handlers.GetProjects).Methods("Get")
	r.HandleFunc("/project/{projectId}", handlers.GetProject).Methods("Get")
	r.HandleFunc("/project/{projectId}", handlers.DeleteProject).Methods("Delete")

	// tasks
	r.HandleFunc("/create/task", handlers.CreateTask).Methods("Post")
	r.HandleFunc("/task/{taskId}/update", handlers.UpdateTask).Methods("Put")
	r.HandleFunc("/task/{taskId}/assign", handlers.AssignTo).Methods("Post")
	r.HandleFunc("/task/{taskId}/status", handlers.Status).Methods("Put")
	r.HandleFunc("/project/{projectId}/tasks", handlers.GetTasks).Methods("Get")
	r.HandleFunc("/task/{taskId}", handlers.GetTask).Methods("Get")
	r.HandleFunc("/task/{taskId}/delete", handlers.DeleteTask).Methods("Delete")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":"+port, r))
}
