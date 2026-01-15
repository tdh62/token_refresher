package database

import (
	"database/sql"
	"fmt"
	"jwt_refresher/models"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func InitDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &DB{db}, nil
}

func createTables(db *sql.DB) error {
	projectsTable := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		enabled BOOLEAN DEFAULT 1,

		refresh_url TEXT NOT NULL,
		refresh_method TEXT DEFAULT 'POST',
		refresh_headers TEXT,
		refresh_body_template TEXT,

		access_token_path TEXT NOT NULL,
		refresh_token_path TEXT NOT NULL,
		expires_in_path TEXT,

		custom_variables TEXT,

		current_access_token TEXT,
		current_refresh_token TEXT,
		token_expires_at DATETIME,

		refresh_before_seconds INTEGER DEFAULT 300,

		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_refresh_at DATETIME,
		last_refresh_status TEXT
	);`

	logsTable := `
	CREATE TABLE IF NOT EXISTS refresh_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		refresh_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		status TEXT NOT NULL,
		error_message TEXT,
		old_token_preview TEXT,
		new_token_preview TEXT,
		old_refresh_token_preview TEXT,
		new_refresh_token_preview TEXT,
		response_body TEXT,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(projectsTable); err != nil {
		return fmt.Errorf("failed to create projects table: %w", err)
	}

	if _, err := db.Exec(logsTable); err != nil {
		return fmt.Errorf("failed to create refresh_logs table: %w", err)
	}

	return nil
}

// Project CRUD operations

func (db *DB) CreateProject(p *models.Project) error {
	query := `
		INSERT INTO projects (
			name, description, enabled,
			refresh_url, refresh_method, refresh_headers, refresh_body_template,
			access_token_path, refresh_token_path, expires_in_path,
			custom_variables, current_refresh_token,
			refresh_before_seconds
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := db.Exec(query,
		p.Name, p.Description, p.Enabled,
		p.RefreshURL, p.RefreshMethod, p.RefreshHeaders, p.RefreshBodyTemplate,
		p.AccessTokenPath, p.RefreshTokenPath, p.ExpiresInPath,
		p.CustomVariables, p.CurrentRefreshToken,
		p.RefreshBeforeSeconds,
	)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	p.ID = id
	return nil
}

func (db *DB) GetProject(id int64) (*models.Project, error) {
	query := `
		SELECT id, name, description, enabled,
			refresh_url, refresh_method, refresh_headers, refresh_body_template,
			access_token_path, refresh_token_path, expires_in_path,
			custom_variables, current_access_token, current_refresh_token, token_expires_at,
			refresh_before_seconds,
			created_at, updated_at, last_refresh_at, last_refresh_status
		FROM projects WHERE id = ?
	`
	pdb := &models.ProjectDB{}
	err := db.QueryRow(query, id).Scan(
		&pdb.ID, &pdb.Name, &pdb.Description, &pdb.Enabled,
		&pdb.RefreshURL, &pdb.RefreshMethod, &pdb.RefreshHeaders, &pdb.RefreshBodyTemplate,
		&pdb.AccessTokenPath, &pdb.RefreshTokenPath, &pdb.ExpiresInPath,
		&pdb.CustomVariables, &pdb.CurrentAccessToken, &pdb.CurrentRefreshToken, &pdb.TokenExpiresAt,
		&pdb.RefreshBeforeSeconds,
		&pdb.CreatedAt, &pdb.UpdatedAt, &pdb.LastRefreshAt, &pdb.LastRefreshStatus,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return pdb.ToProject(), nil
}

func (db *DB) GetAllProjects() ([]*models.Project, error) {
	query := `
		SELECT id, name, description, enabled,
			refresh_url, refresh_method, refresh_headers, refresh_body_template,
			access_token_path, refresh_token_path, expires_in_path,
			custom_variables, current_access_token, current_refresh_token, token_expires_at,
			refresh_before_seconds,
			created_at, updated_at, last_refresh_at, last_refresh_status
		FROM projects ORDER BY created_at DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		pdb := &models.ProjectDB{}
		err := rows.Scan(
			&pdb.ID, &pdb.Name, &pdb.Description, &pdb.Enabled,
			&pdb.RefreshURL, &pdb.RefreshMethod, &pdb.RefreshHeaders, &pdb.RefreshBodyTemplate,
			&pdb.AccessTokenPath, &pdb.RefreshTokenPath, &pdb.ExpiresInPath,
			&pdb.CustomVariables, &pdb.CurrentAccessToken, &pdb.CurrentRefreshToken, &pdb.TokenExpiresAt,
			&pdb.RefreshBeforeSeconds,
			&pdb.CreatedAt, &pdb.UpdatedAt, &pdb.LastRefreshAt, &pdb.LastRefreshStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, pdb.ToProject())
	}
	return projects, nil
}

func (db *DB) GetEnabledProjects() ([]*models.Project, error) {
	query := `
		SELECT id, name, description, enabled,
			refresh_url, refresh_method, refresh_headers, refresh_body_template,
			access_token_path, refresh_token_path, expires_in_path,
			custom_variables, current_access_token, current_refresh_token, token_expires_at,
			refresh_before_seconds,
			created_at, updated_at, last_refresh_at, last_refresh_status
		FROM projects WHERE enabled = 1
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		pdb := &models.ProjectDB{}
		err := rows.Scan(
			&pdb.ID, &pdb.Name, &pdb.Description, &pdb.Enabled,
			&pdb.RefreshURL, &pdb.RefreshMethod, &pdb.RefreshHeaders, &pdb.RefreshBodyTemplate,
			&pdb.AccessTokenPath, &pdb.RefreshTokenPath, &pdb.ExpiresInPath,
			&pdb.CustomVariables, &pdb.CurrentAccessToken, &pdb.CurrentRefreshToken, &pdb.TokenExpiresAt,
			&pdb.RefreshBeforeSeconds,
			&pdb.CreatedAt, &pdb.UpdatedAt, &pdb.LastRefreshAt, &pdb.LastRefreshStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, pdb.ToProject())
	}
	return projects, nil
}

func (db *DB) UpdateProject(p *models.Project) error {
	query := `
		UPDATE projects SET
			name = ?, description = ?, enabled = ?,
			refresh_url = ?, refresh_method = ?, refresh_headers = ?, refresh_body_template = ?,
			access_token_path = ?, refresh_token_path = ?, expires_in_path = ?,
			custom_variables = ?, current_refresh_token = ?,
			refresh_before_seconds = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.Exec(query,
		p.Name, p.Description, p.Enabled,
		p.RefreshURL, p.RefreshMethod, p.RefreshHeaders, p.RefreshBodyTemplate,
		p.AccessTokenPath, p.RefreshTokenPath, p.ExpiresInPath,
		p.CustomVariables, p.CurrentRefreshToken,
		p.RefreshBeforeSeconds,
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

func (db *DB) UpdateProjectTokens(id int64, accessToken, refreshToken string, expiresAt time.Time, status string) error {
	query := `
		UPDATE projects SET
			current_access_token = ?,
			current_refresh_token = ?,
			token_expires_at = ?,
			last_refresh_at = CURRENT_TIMESTAMP,
			last_refresh_status = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.Exec(query, accessToken, refreshToken, expiresAt, status, id)
	if err != nil {
		return fmt.Errorf("failed to update project tokens: %w", err)
	}
	return nil
}

func (db *DB) UpdateProjectRefreshStatus(id int64, status string) error {
	query := `
		UPDATE projects SET
			last_refresh_at = CURRENT_TIMESTAMP,
			last_refresh_status = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update project refresh status: %w", err)
	}
	return nil
}

func (db *DB) DeleteProject(id int64) error {
	query := `DELETE FROM projects WHERE id = ?`
	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

func (db *DB) ToggleProject(id int64) error {
	query := `UPDATE projects SET enabled = NOT enabled, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to toggle project: %w", err)
	}
	return nil
}

// RefreshLog operations

func (db *DB) CreateRefreshLog(log *models.RefreshLog) error {
	query := `
		INSERT INTO refresh_logs (
			project_id, status, error_message,
			old_token_preview, new_token_preview,
			old_refresh_token_preview, new_refresh_token_preview,
			response_body
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := db.Exec(query,
		log.ProjectID, log.Status, log.ErrorMessage,
		log.OldTokenPreview, log.NewTokenPreview,
		log.OldRefreshTokenPreview, log.NewRefreshTokenPreview,
		log.ResponseBody,
	)
	if err != nil {
		return fmt.Errorf("failed to create refresh log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	log.ID = id
	return nil
}

func (db *DB) GetProjectLogs(projectID int64, limit int) ([]*models.RefreshLog, error) {
	query := `
		SELECT id, project_id, refresh_at, status, error_message,
			old_token_preview, new_token_preview,
			old_refresh_token_preview, new_refresh_token_preview,
			response_body
		FROM refresh_logs
		WHERE project_id = ?
		ORDER BY refresh_at DESC
		LIMIT ?
	`
	rows, err := db.Query(query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get project logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.RefreshLog
	for rows.Next() {
		log := &models.RefreshLog{}
		err := rows.Scan(
			&log.ID, &log.ProjectID, &log.RefreshAt, &log.Status, &log.ErrorMessage,
			&log.OldTokenPreview, &log.NewTokenPreview,
			&log.OldRefreshTokenPreview, &log.NewRefreshTokenPreview,
			&log.ResponseBody,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan refresh log: %w", err)
		}
		logs = append(logs, log)
	}
	return logs, nil
}
