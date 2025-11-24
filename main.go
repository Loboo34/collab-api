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

	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
})

	//handlers
	//auth
	r.HandleFunc("/auth/register", handlers.RegisterUser).Methods("POST")
	r.HandleFunc("/auth/login", handlers.LoginUser).Methods("POST")

	// teams
	r.HandleFunc("/team/create", middleware.CheckAuth(handlers.CreateTeam)).Methods("Post")
	r.HandleFunc("/team/invite", middleware.CheckAuth(middleware.CheckRole("Admin", handlers.InviteMember))).Methods("Post")
	r.HandleFunc("/invite/accept", middleware.CheckAuth(handlers.AcceptInvite)).Methods("Post")
	r.HandleFunc("/invite/Decline", middleware.CheckAuth(handlers.DeclineInvite)).Methods("Post")
	r.HandleFunc("/team/{teamId}/members", middleware.CheckAuth(handlers.GetTeamMembers)).Methods("Get")
	r.HandleFunc("/team/{teamId}/", middleware.CheckAuth(middleware.CheckRole("Admin", handlers.ChangeRole)))
	r.HandleFunc("/team/{teamId}/remove", middleware.CheckAuth(middleware.CheckRole("Admin", handlers.RemoveMember))).Methods("Delete")
	r.HandleFunc("/team/{teamId}", middleware.CheckAuth(middleware.CheckRole("Admin", handlers.DeleteTeam))).Methods("Delete")


	// project
	r.HandleFunc("/project/create/{teamId}", middleware.CheckAuth(middleware.CheckRole("Admin", handlers.CreateProject))).Methods("Post")
	r.HandleFunc("/project/{projectId}/update", middleware.CheckAuth(middleware.CheckRole("Admin",handlers.UpdateProject))).Methods("Put")
	r.HandleFunc("/team/{teamId}/projects", middleware.CheckAuth(handlers.GetProjects)).Methods("Get")
	r.HandleFunc("/project/{projectId}", middleware.CheckAuth(handlers.GetProject)).Methods("Get")
	r.HandleFunc("/project/{projectId}", middleware.CheckAuth(middleware.CheckRole("Admin", handlers.DeleteProject))).Methods("Delete")

	// tasks
	r.HandleFunc("/task/create", middleware.CheckAuth(handlers.CreateTask)).Methods("Post")
	r.HandleFunc("/task/{taskId}/update", middleware.CheckAuth(handlers.UpdateTask)).Methods("Put")
	r.HandleFunc("/task/{taskId}/assign", middleware.CheckAuth(handlers.AssignTo)).Methods("Post")
	r.HandleFunc("/task/{taskId}/status", middleware.CheckAuth(handlers.Status)).Methods("Put")
	r.HandleFunc("/project/{projectId}/tasks", middleware.CheckAuth(handlers.GetTasks)).Methods("Get")
	r.HandleFunc("/task/{taskId}", middleware.CheckAuth(handlers.GetTask)).Methods("Get")
	r.HandleFunc("/task/{taskId}", middleware.CheckAuth(handlers.DeleteTask)).Methods("Delete")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":"+port, r))
}
