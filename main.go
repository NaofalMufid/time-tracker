package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"task-time-tracker/db"
	"task-time-tracker/handlers"
	"task-time-tracker/utils"
	"time"

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
	defer db.DB.Close()

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
	appURL := fmt.Sprintf("http://localhost:%s", port)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server runnin on %s", appURL)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	go func() {
		time.Sleep(500 * time.Millisecond)
		if err := openbrowser(appURL); err != nil {
			log.Printf("Failed to open browser: %v", err)
		}
	}()

	<-stop

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server gracefully stopped.")

	// log.Println("Server running on http://localhost:" + port)
	// log.Fatal(http.ListenAndServe(":"+port, r))
}

func openbrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	default:
		return fmt.Errorf("platform is not support for opening browser: %s", runtime.GOOS)
	}

	command := exec.Command(cmd, args...)
	log.Printf("Tyring open browser with command: %s %v", cmd, args)
	if err := command.Start(); err != nil {
		return fmt.Errorf("failed starting browser command")
	}
	return nil
}
