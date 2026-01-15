package models

import (
	"time"
)

type RefreshLog struct {
	ID                    int64     `json:"id"`
	ProjectID             int64     `json:"project_id"`
	RefreshAt             time.Time `json:"refresh_at"`
	Status                string    `json:"status"` // success, failed
	ErrorMessage          string    `json:"error_message"`
	OldTokenPreview       string    `json:"old_token_preview"`
	NewTokenPreview       string    `json:"new_token_preview"`
	OldRefreshTokenPreview string   `json:"old_refresh_token_preview"`
	NewRefreshTokenPreview string   `json:"new_refresh_token_preview"`
	ResponseBody          string    `json:"response_body"`
}
