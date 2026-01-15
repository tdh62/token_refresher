package refresher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"jwt_refresher/database"
	"jwt_refresher/models"
	"log"
	"net/http"
	"strings"
	"time"
)

type Engine struct {
	db *database.DB
}

func NewEngine(db *database.DB) *Engine {
	return &Engine{db: db}
}

func (e *Engine) Refresh(project *models.Project) error {
	log.Printf("Starting refresh for project: %s (ID: %d)", project.Name, project.ID)

	oldTokenPreview := ""
	if len(project.CurrentAccessToken) > 10 {
		oldTokenPreview = project.CurrentAccessToken[:10]
	} else {
		oldTokenPreview = project.CurrentAccessToken
	}

	// 1. 构建HTTP请求体（替换模板变量）
	body, err := RenderTemplate(project.RefreshBodyTemplate, project)
	if err != nil {
		e.logRefreshError(project.ID, oldTokenPreview, "", fmt.Sprintf("Failed to render template: %v", err), "")
		return fmt.Errorf("failed to render template: %w", err)
	}

	// 2. 创建HTTP请求
	req, err := http.NewRequest(project.RefreshMethod, project.RefreshURL, bytes.NewBufferString(body))
	if err != nil {
		e.logRefreshError(project.ID, oldTokenPreview, "", fmt.Sprintf("Failed to create request: %v", err), "")
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 3. 设置请求头
	if project.RefreshHeaders != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(project.RefreshHeaders), &headers); err != nil {
			e.logRefreshError(project.ID, oldTokenPreview, "", fmt.Sprintf("Failed to parse headers: %v", err), "")
			return fmt.Errorf("failed to parse headers: %w", err)
		}
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// 4. 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		e.logRefreshError(project.ID, oldTokenPreview, "", fmt.Sprintf("Failed to send request: %v", err), "")
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 5. 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		e.logRefreshError(project.ID, oldTokenPreview, "", fmt.Sprintf("Failed to read response: %v", err), "")
		return fmt.Errorf("failed to read response: %w", err)
	}

	respBodyStr := string(respBody)

	// 6. 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		e.logRefreshError(project.ID, oldTokenPreview, "", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, respBodyStr), respBodyStr)
		return fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, respBodyStr)
	}

	// 7. 使用JSONPath提取token
	accessToken, err := ExtractToken(respBodyStr, project.AccessTokenPath)
	if err != nil {
		e.logRefreshError(project.ID, oldTokenPreview, "", fmt.Sprintf("Failed to extract access token: %v", err), respBodyStr)
		return fmt.Errorf("failed to extract access token: %w", err)
	}

	refreshToken, err := ExtractToken(respBodyStr, project.RefreshTokenPath)
	if err != nil {
		e.logRefreshError(project.ID, oldTokenPreview, "", fmt.Sprintf("Failed to extract refresh token: %v", err), respBodyStr)
		return fmt.Errorf("failed to extract refresh token: %w", err)
	}

	// 8. 提取过期时间（如果有）
	var expiresAt time.Time
	if project.ExpiresInPath != "" {
		expiresIn, err := ExtractExpiresIn(respBodyStr, project.ExpiresInPath)
		if err != nil {
			log.Printf("Warning: Failed to extract expires_in: %v", err)
			// 不是致命错误，继续执行
		} else {
			expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
		}
	}

	// 9. 更新数据库
	if err := e.db.UpdateProjectTokens(project.ID, accessToken, refreshToken, expiresAt, "success"); err != nil {
		e.logRefreshError(project.ID, oldTokenPreview, "", fmt.Sprintf("Failed to update database: %v", err), respBodyStr)
		return fmt.Errorf("failed to update database: %w", err)
	}

	// 10. 记录成功日志
	newTokenPreview := ""
	if len(accessToken) > 10 {
		newTokenPreview = accessToken[:10]
	} else {
		newTokenPreview = accessToken
	}

	oldRefreshTokenPreview := ""
	if len(project.CurrentRefreshToken) > 10 {
		oldRefreshTokenPreview = project.CurrentRefreshToken[:10]
	} else {
		oldRefreshTokenPreview = project.CurrentRefreshToken
	}

	newRefreshTokenPreview := ""
	if len(refreshToken) > 10 {
		newRefreshTokenPreview = refreshToken[:10]
	} else {
		newRefreshTokenPreview = refreshToken
	}

	logEntry := &models.RefreshLog{
		ProjectID:              project.ID,
		Status:                 "success",
		OldTokenPreview:        oldTokenPreview,
		NewTokenPreview:        newTokenPreview,
		OldRefreshTokenPreview: oldRefreshTokenPreview,
		NewRefreshTokenPreview: newRefreshTokenPreview,
		ResponseBody:           respBodyStr,
	}
	if err := e.db.CreateRefreshLog(logEntry); err != nil {
		log.Printf("Warning: Failed to create refresh log: %v", err)
	}

	log.Printf("Successfully refreshed tokens for project: %s (ID: %d)", project.Name, project.ID)
	return nil
}

func (e *Engine) logRefreshError(projectID int64, oldTokenPreview, newTokenPreview, errorMsg, responseBody string) {
	// 更新项目状态
	if err := e.db.UpdateProjectRefreshStatus(projectID, "failed"); err != nil {
		log.Printf("Warning: Failed to update project refresh status: %v", err)
	}

	// 记录错误日志
	logEntry := &models.RefreshLog{
		ProjectID:       projectID,
		Status:          "failed",
		ErrorMessage:    errorMsg,
		OldTokenPreview: oldTokenPreview,
		NewTokenPreview: newTokenPreview,
		ResponseBody:    responseBody,
	}
	if err := e.db.CreateRefreshLog(logEntry); err != nil {
		log.Printf("Warning: Failed to create refresh log: %v", err)
	}
}

func (e *Engine) ShouldRefresh(project *models.Project) bool {
	// 如果没有access token，需要刷新
	if project.CurrentAccessToken == "" {
		return true
	}

	// 如果没有设置过期时间，不刷新
	if !project.TokenExpiresAt.Valid {
		return false
	}

	// 计算距离过期还有多久
	timeUntilExpiry := time.Until(project.TokenExpiresAt.Time)

	// 如果已经过期或即将过期（在refresh_before_seconds秒内），需要刷新
	refreshBefore := time.Duration(project.RefreshBeforeSeconds) * time.Second
	return timeUntilExpiry <= refreshBefore
}

func getTokenPreview(token string) string {
	if len(token) > 10 {
		return token[:10] + "..."
	}
	return token
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func sanitizeForLog(s string) string {
	// 移除敏感信息
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	return truncateString(s, 500)
}
