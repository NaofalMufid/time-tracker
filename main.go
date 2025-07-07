package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"task-time-tracker/db"
	"task-time-tracker/handlers"
	"task-time-tracker/utils"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed static/*
var embeddedFiles embed.FS

func initEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using default config")
	}
}

func main() {
	initEnv()
	db.Init()
	db.Migrate()

	r := mux.NewRouter()
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}).Methods("GET")

	r.HandleFunc("/tasks", handlers.CreateTaskHandler()).Methods("POST")
	r.HandleFunc("/tasks/{id}/pause", handlers.PauseTaskHandler()).Methods("POST")
	r.HandleFunc("/tasks/{id}/resume", handlers.ResumeTaskHandler()).Methods("POST")
	r.HandleFunc("/tasks/{id}/stop", handlers.StopTaskHandler()).Methods("POST")
	r.HandleFunc("/tasks/{id}", handlers.DeleteTaskHandler()).Methods("DELETE")
	r.HandleFunc("/tasks", handlers.ListTasksHandler()).Methods("GET")
	r.HandleFunc("/tasks/running", handlers.GetRunningTasksHandler()).Methods("GET")

	r.PathPrefix("/static/").Handler(http.FileServer(http.FS(embeddedFiles)))

	// serve the index.html for the root path
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		tmpl, err := template.ParseFS(embeddedFiles, "static/index.html")
		if err != nil {
			log.Printf("Error parsing index.html: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			log.Printf("Error executing index.html: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server running on http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
