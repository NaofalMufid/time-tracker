package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"task-time-tracker/db"
	"task-time-tracker/models"
	"task-time-tracker/utils"
	"time"

	"github.com/gorilla/mux"
)

type PaginatedTasksResponse struct {
	Data       []models.Task     `json:"data"`
	Pagination utils.PageDetails `json:"pagination"`
}

func getActiveTaskID(excludeID int) (int, bool, error) {
	var id int
	query := `
		SELECT id FROM tasks
		WHERE is_paused = 0 AND end_time IS NULL`
	args := []interface{}{}

	if excludeID > 0 {
		query += " AND id != ?"
		args = append(args, excludeID)
	}

	err := db.DB.QueryRow(query, args...).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

func ListTasksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, pageSize := utils.GetParams(r)
		offset := (page - 1) * pageSize

		status := r.URL.Query().Get("status")
		sortBy := r.URL.Query().Get("sortBy")
		orderBy := r.URL.Query().Get("orderBy")

		var queryArgs []interface{}

		baseQuery := "FROM tasks"

		whereClause := ""
		switch status {
		case "paused":
			whereClause = "WHERE is_paused = 1"
		case "stopped":
			whereClause = "WHERE end_time IS NOT NULL"
		}

		sortableColumns := map[string]string{
			"title":      "title",
			"start_time": "start_time",
		}

		sortColumn, isValidColumn := sortableColumns[sortBy]
		if !isValidColumn {
			sortColumn = "start_time"
		}

		if strings.ToUpper(orderBy) != "ASC" {
			orderBy = "DESC"
		}

		orderClause := fmt.Sprintf("ORDER BY %s %s", sortColumn, orderBy)

		var totalRecords int64
		countQuery := fmt.Sprintf("SELECT COUNT(id) %s %s", baseQuery, whereClause)
		err := db.DB.QueryRow(countQuery, queryArgs...).Scan(&totalRecords)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to count task")
			return
		}

		dataQuery := fmt.Sprintf(`
			SELECT id, title, start_time, end_time, is_paused,
				paused_duration, duration, last_resume_time
			%s %s %s LIMIT ? OFFSET ?
		`, baseQuery, whereClause, orderClause)

		finalArgs := append(queryArgs, pageSize, offset)

		rows, err := db.DB.Query(dataQuery, finalArgs...)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch tasks")
			return
		}
		defer rows.Close()

		tasks := []models.Task{}
		for rows.Next() {
			var t models.Task
			var endTime, lastResume sql.NullTime

			err := rows.Scan(&t.ID, &t.Title, &t.StartTime, &endTime, &t.IsPaused,
				&t.PausedDuration, &t.Duration, &lastResume)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, "Scan error")
				return
			}
			if endTime.Valid {
				t.EndTime = &endTime.Time
			}
			if lastResume.Valid {
				t.LastResumeTime = &lastResume.Time
			}
			tasks = append(tasks, t)
		}

		pageDetail := utils.CalculateMetadata(totalRecords, page, pageSize)

		response := PaginatedTasksResponse{
			Data:       tasks,
			Pagination: pageDetail,
		}

		utils.WriteJSON(w, http.StatusOK, response)
	}
}

func GetRunningTasksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `SELECT * FROM tasks WHERE is_paused = 0 AND end_time IS NULL`
		rows, err := db.DB.Query(query)

		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch active task")
			return
		}

		defer rows.Close()

		task := []models.Task{}
		for rows.Next() {
			var t models.Task
			var endTime, lastResume sql.NullTime

			err := rows.Scan(&t.ID, &t.Title, &t.StartTime, &endTime, &t.IsPaused,
				&t.PausedDuration, &t.Duration, &lastResume)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, "Scan error")
				return
			}
			if endTime.Valid {
				t.EndTime = &endTime.Time
			}
			if lastResume.Valid {
				t.LastResumeTime = &lastResume.Time
			}
			task = append(task, t)
		}

		response := models.Task{}
		if len(task) > 0 {
			response = task[0]
		}

		utils.WriteJSON(w, http.StatusOK, response)
	}
}

func CreateTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Title == "" {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON or missing title")
			return
		}

		// Enforce only one active task
		if _, found, err := getActiveTaskID(0); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "DB error")
			return
		} else if found {
			utils.WriteError(w, http.StatusConflict, "Another task is already running")
			return
		}

		now := time.Now()
		result, err := db.DB.Exec(`
			INSERT INTO tasks (title, start_time, is_paused, paused_duration, duration, last_resume_time)
			VALUES (?, ?, 0, 0, 0, ?)`,
			input.Title, now, now,
		)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to create task")
			return
		}

		id, _ := result.LastInsertId()
		task := models.Task{
			ID:             int(id),
			Title:          input.Title,
			StartTime:      now,
			IsPaused:       false,
			PausedDuration: 0,
			Duration:       0,
			LastResumeTime: &now,
		}
		utils.WriteJSON(w, http.StatusCreated, task)
	}
}

func PauseTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := mux.Vars(r)["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid task ID")
			return
		}

		var isPaused bool
		var lastResume time.Time
		var pausedSoFar int

		row := db.DB.QueryRow(`
			SELECT is_paused, last_resume_time, paused_duration
			FROM tasks WHERE id = ? AND end_time IS NULL`, id)
		err = row.Scan(&isPaused, &lastResume, &pausedSoFar)
		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, "Task not found or already stopped")
			return
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "DB error")
			return
		}
		if isPaused {
			utils.WriteError(w, http.StatusConflict, "Task already paused")
			return
		}

		now := time.Now()
		delta := int(now.Sub(lastResume).Seconds())
		_, err = db.DB.Exec(`
			UPDATE tasks
			SET is_paused = 1,
			    paused_duration = ?,
			    duration = ?,
			    last_resume_time = NULL
			WHERE id = ?`,
			pausedSoFar+delta, pausedSoFar+delta, id,
		)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to pause task")
			return
		}

		utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "paused"})
	}
}

func ResumeTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := mux.Vars(r)["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid task ID")
			return
		}

		// Check if task exists and is paused
		var isPaused bool
		err = db.DB.QueryRow(`
			SELECT is_paused FROM tasks
			WHERE id = ? AND end_time IS NULL`, id).Scan(&isPaused)
		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, "Task not found or already stopped")
			return
		}
		if err != nil {
			fmt.Println(err.Error())
			utils.WriteError(w, http.StatusInternalServerError, "DB error")
			return
		}
		if !isPaused {
			utils.WriteError(w, http.StatusConflict, "Task is not paused")
			return
		}

		// Enforce only one active task
		if _, found, err := getActiveTaskID(id); err != nil {
			fmt.Println(err.Error())
			utils.WriteError(w, http.StatusInternalServerError, "DB error")
			return
		} else if found {
			utils.WriteError(w, http.StatusConflict, "Another task is already running")
			return
		}

		now := time.Now()
		_, err = db.DB.Exec(`
			UPDATE tasks
			SET is_paused = 0,
			    last_resume_time = ?
			WHERE id = ?`,
			now, id,
		)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to resume task")
			return
		}

		utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "resumed"})
	}
}

func StopTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := mux.Vars(r)["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid task ID")
			return
		}

		var isPaused bool
		var pausedDuration int
		var lastResume sql.NullTime

		row := db.DB.QueryRow(`
			SELECT is_paused, paused_duration, last_resume_time
			FROM tasks WHERE id = ? AND end_time IS NULL`, id)
		err = row.Scan(&isPaused, &pausedDuration, &lastResume)
		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, "Task not found or already stopped")
			return
		}
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "DB error")
			return
		}

		now := time.Now()
		totalDuration := pausedDuration
		if !isPaused && lastResume.Valid {
			delta := int(now.Sub(lastResume.Time).Seconds())
			totalDuration += delta
		}

		_, err = db.DB.Exec(`
			UPDATE tasks
			SET end_time = ?, is_paused = 0,
			    duration = ?, last_resume_time = NULL
			WHERE id = ?`,
			now, totalDuration, id,
		)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to stop task")
			return
		}

		utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "stopped", "duration": strconv.Itoa(totalDuration)})
	}
}

func DeleteTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := mux.Vars(r)["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid task ID")
			return
		}

		result, err := db.DB.Exec(`DELETE FROM tasks WHERE id = ?`, id)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to delete task")
			return
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			utils.WriteError(w, http.StatusNotFound, "Task not found")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
