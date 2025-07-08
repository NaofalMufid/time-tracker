package models

import "time"

type Task struct {
	ID             int        `json:"id"`
	Title          string     `json:"title"`
	Detail         string     `json:"detail"`
	StartTime      time.Time  `json:"start_time"`
	EndTime        *time.Time `json:"end_time,omitempty"`
	IsPaused       bool       `json:"is_paused"`
	PausedDuration int        `json:"paused_duration"` // in seconds
	Duration       int        `json:"duration"`        // in seconds
	LastResumeTime *time.Time `json:"last_resume_time,omitempty"`
}
