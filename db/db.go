package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() {
	// var err error
	// dbPath := os.Getenv("DB_PATH")
	// if dbPath == "" {
	// 	dbPath = "./time_tracker.db"
	// }

	// DB, err = sql.Open("sqlite3", dbPath)
	// if err != nil {
	// 	log.Fatal("Failed to open database:", err)
	// }

	// if err := DB.Ping(); err != nil {
	// 	log.Fatal("Database ping failed:", err)
	// }
	// log.Println("Connected to SQLite at", dbPath)

	dataDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Error getting user config directory: %v", err)
	}

	appDataDir := filepath.Join(dataDir, "time-tracker")

	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		log.Fatalf("Error creating application data directory: %v", err)
	}

	dbPath := filepath.Join(appDataDir, "time_tracker.db")
	log.Printf("Opening database at: %s", dbPath)

	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Error connecting database: %v", err)
	}
	log.Println("Database connection established.")
}

func Migrate() {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		is_paused BOOLEAN DEFAULT FALSE,
		paused_duration INTEGER DEFAULT 0,
		duration INTEGER DEFAULT 0,
		last_resume_time DATETIME
	);
	`
	_, err := DB.Exec(schema)
	if err != nil {
		log.Fatal("Failed to run DB migration:", err)
	}
	log.Println("Database migrated successfully.")
}
