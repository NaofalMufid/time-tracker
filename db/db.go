package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() {
	var err error
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./time_tracker.db"
	}

	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}
	log.Println("Connected to SQLite at", dbPath)
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
