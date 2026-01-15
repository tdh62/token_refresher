package models

import (
	"database/sql"
	"time"
)

type Project struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`

	// 刷新配置
	RefreshURL          string `json:"refresh_url"`
	RefreshMethod       string `json:"refresh_method"`
	RefreshHeaders      string `json:"refresh_headers"`
	RefreshBodyTemplate string `json:"refresh_body_template"`

	// Token提取规则 (JSONPath)
	AccessTokenPath  string `json:"access_token_path"`
	RefreshTokenPath string `json:"refresh_token_path"`
	ExpiresInPath    string `json:"expires_in_path"`

	// 自定义变量 (JSON格式: {"ClientId": "xxx", "ClientSecret": "yyy", ...})
	CustomVariables string `json:"custom_variables"`

	// 当前Token数据
	CurrentAccessToken  string       `json:"current_access_token"`
	CurrentRefreshToken string       `json:"current_refresh_token"`
	TokenExpiresAt      sql.NullTime `json:"token_expires_at"`

	// 刷新策略
	RefreshBeforeSeconds int `json:"refresh_before_seconds"`

	// 元数据
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	LastRefreshAt     sql.NullTime `json:"last_refresh_at"`
	LastRefreshStatus string       `json:"last_refresh_status"`
}

// 用于数据库扫描的辅助结构
type ProjectDB struct {
	ID          int64
	Name        string
	Description sql.NullString
	Enabled     bool

	RefreshURL          string
	RefreshMethod       string
	RefreshHeaders      sql.NullString
	RefreshBodyTemplate sql.NullString

	AccessTokenPath  string
	RefreshTokenPath string
	ExpiresInPath    sql.NullString

	CustomVariables sql.NullString

	CurrentAccessToken  sql.NullString
	CurrentRefreshToken sql.NullString
	TokenExpiresAt      sql.NullTime

	RefreshBeforeSeconds int

	CreatedAt         time.Time
	UpdatedAt         time.Time
	LastRefreshAt     sql.NullTime
	LastRefreshStatus sql.NullString
}

// 转换为Project
func (pdb *ProjectDB) ToProject() *Project {
	return &Project{
		ID:                   pdb.ID,
		Name:                 pdb.Name,
		Description:          pdb.Description.String,
		Enabled:              pdb.Enabled,
		RefreshURL:           pdb.RefreshURL,
		RefreshMethod:        pdb.RefreshMethod,
		RefreshHeaders:       pdb.RefreshHeaders.String,
		RefreshBodyTemplate:  pdb.RefreshBodyTemplate.String,
		AccessTokenPath:      pdb.AccessTokenPath,
		RefreshTokenPath:     pdb.RefreshTokenPath,
		ExpiresInPath:        pdb.ExpiresInPath.String,
		CustomVariables:      pdb.CustomVariables.String,
		CurrentAccessToken:   pdb.CurrentAccessToken.String,
		CurrentRefreshToken:  pdb.CurrentRefreshToken.String,
		TokenExpiresAt:       pdb.TokenExpiresAt,
		RefreshBeforeSeconds: pdb.RefreshBeforeSeconds,
		CreatedAt:            pdb.CreatedAt,
		UpdatedAt:            pdb.UpdatedAt,
		LastRefreshAt:        pdb.LastRefreshAt,
		LastRefreshStatus:    pdb.LastRefreshStatus.String,
	}
}
