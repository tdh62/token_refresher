package main

import (
	"embed"
	"fmt"
	"io"
	"jwt_refresher/api"
	"jwt_refresher/config"
	"jwt_refresher/database"
	"jwt_refresher/logger"
	"jwt_refresher/refresher"
	"jwt_refresher/scheduler"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

//go:embed web/static/*
var staticFiles embed.FS

func main() {
	log.Println("Starting JWT Token Refresher...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded: Port=%d, DataDir=%s", cfg.Port, cfg.DataDir)

	// Setup logging with rotation
	rotatingLogger, err := logger.SetupRotatingLogger(cfg.DataDir, cfg.LogFile)
	if err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}
	defer rotatingLogger.Close()
	log.Println("Logging configured successfully (with rotation: 10MB max, 5 backups)")

	// Create data directory if not exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}
	log.Printf("Data directory ready: %s", cfg.DataDir)

	// Migrate existing database if needed
	if err := migrateDatabase(cfg.DataDir); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化数据库
	db, err := database.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Printf("Database initialized: %s", cfg.DBPath)

	// 创建刷新引擎
	engine := refresher.NewEngine(db)
	log.Println("Refresh engine created")

	// 创建并启动调度器
	sched := scheduler.NewScheduler(db, engine)
	sched.Start()
	defer sched.Stop()

	// 设置Web服务
	router := api.SetupRouter(db, engine, staticFiles, cfg.Username, cfg.Password)

	// 启动Web服务
	log.Printf("Starting web server on port %d...", cfg.Port)
	log.Printf("Access the web interface at: http://localhost:%d", cfg.Port)
	log.Printf("Authentication required: username=%s", cfg.Username)

	// 优雅关闭
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Port)
		if err := router.Run(addr); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	sched.Stop()
	log.Println("Server stopped")
}

// migrateDatabase moves existing jwt_refresher.db to data directory
func migrateDatabase(dataDir string) error {
	oldPath := "jwt_refresher.db"
	newPath := filepath.Join(dataDir, "jwt_refresher.db")

	// Check if old database exists in root
	if _, err := os.Stat(oldPath); err == nil {
		// Check if new location already has database
		if _, err := os.Stat(newPath); err == nil {
			log.Printf("Database already exists at %s, skipping migration", newPath)
			log.Printf("WARNING: Old database at %s will not be used", oldPath)
			return nil
		}

		// Move database to new location
		log.Printf("Migrating database from %s to %s", oldPath, newPath)
		if err := os.Rename(oldPath, newPath); err != nil {
			// If rename fails (cross-device), try copy
			if err := copyFile(oldPath, newPath); err != nil {
				return fmt.Errorf("failed to migrate database: %w", err)
			}
			// Remove old file after successful copy
			os.Remove(oldPath)
		}
		log.Println("Database migration completed successfully")
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Sync()
}
