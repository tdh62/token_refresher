package api

import (
	"embed"
	"io/fs"
	"jwt_refresher/database"
	"jwt_refresher/refresher"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(db *database.DB, engine *refresher.Engine, staticFiles embed.FS, username, password string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Create auth middleware
	authMiddleware := BasicAuthMiddleware(username, password)

	// API handlers
	projectHandler := NewProjectHandler(db, engine)
	tokenHandler := NewTokenHandler(db)

	// Protected API routes
	api := r.Group("/api")
	api.Use(authMiddleware) // Apply auth to all API routes
	{
		// 项目管理
		api.GET("/projects", projectHandler.GetAllProjects)
		api.GET("/projects/:id", projectHandler.GetProject)
		api.POST("/projects", projectHandler.CreateProject)
		api.PUT("/projects/:id", projectHandler.UpdateProject)
		api.DELETE("/projects/:id", projectHandler.DeleteProject)
		api.POST("/projects/:id/toggle", projectHandler.ToggleProject)
		api.POST("/projects/:id/refresh", projectHandler.RefreshProject)

		// Token查询
		api.GET("/projects/:id/token", tokenHandler.GetToken)
		api.GET("/projects/:id/logs", tokenHandler.GetLogs)
	}

	// Protected static files and web interface
	staticFS, err := fs.Sub(staticFiles, "web/static")
	if err == nil {
		protected := r.Group("/")
		protected.Use(authMiddleware)
		{
			protected.StaticFS("/static", http.FS(staticFS))
			protected.GET("/", func(c *gin.Context) {
				c.Redirect(http.StatusMovedPermanently, "/static/index.html")
			})
		}
	}

	return r
}
