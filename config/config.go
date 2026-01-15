package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port     int    `yaml:"port"`
	DataDir  string `yaml:"data_dir"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	LogFile  string `yaml:"log_file"`

	// Computed fields (not in YAML)
	DBPath string `yaml:"-"`
}

func Load() (*Config, error) {
	// Default configuration
	cfg := &Config{
		Port:    3007,
		DataDir: "./data",
		LogFile: "app.log",
	}

	// Try to load from config.yaml
	if data, err := os.ReadFile("config.yaml"); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config.yaml: %w", err)
		}
		log.Println("Loaded configuration from config.yaml")
	} else {
		log.Println("config.yaml not found, using defaults and environment variables")
	}

	// Environment variables override config file
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}
	if dataDir := os.Getenv("DATA_DIR"); dataDir != "" {
		cfg.DataDir = dataDir
	}
	if username := os.Getenv("USERNAME"); username != "" {
		cfg.Username = username
	}
	if password := os.Getenv("PASSWORD"); password != "" {
		cfg.Password = password
	}
	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		cfg.LogFile = logFile
	}

	// Compute derived paths
	cfg.DBPath = filepath.Join(cfg.DataDir, "jwt_refresher.db")

	// Validate required fields
	if cfg.Username == "" || cfg.Password == "" {
		return nil, fmt.Errorf("username and password must be set via config file or environment variables")
	}

	return cfg, nil
}
