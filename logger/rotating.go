package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	maxLogSize = 10 * 1024 * 1024 // 10MB
	maxBackups = 5                 // 保留5个备份文件
)

type RotatingLogger struct {
	logPath    string
	file       *os.File
	size       int64
	mu         sync.Mutex
	multiWrite io.Writer
}

func NewRotatingLogger(logPath string) (*RotatingLogger, error) {
	rl := &RotatingLogger{
		logPath: logPath,
	}

	if err := rl.openFile(); err != nil {
		return nil, err
	}

	// 创建多输出writer（stdout + file）
	rl.multiWrite = io.MultiWriter(os.Stdout, rl.file)
	log.SetOutput(rl.multiWrite)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return rl, nil
}

func (rl *RotatingLogger) openFile() error {
	// 获取文件信息
	info, err := os.Stat(rl.logPath)
	if err == nil {
		rl.size = info.Size()
	}

	// 打开或创建日志文件
	file, err := os.OpenFile(rl.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	rl.file = file
	return nil
}

func (rl *RotatingLogger) Write(p []byte) (n int, err error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 检查是否需要轮转
	if rl.size+int64(len(p)) > maxLogSize {
		if err := rl.rotate(); err != nil {
			return 0, err
		}
	}

	// 写入数据
	n, err = rl.file.Write(p)
	if err != nil {
		return n, err
	}

	// 同时写入stdout
	os.Stdout.Write(p)

	rl.size += int64(n)
	return n, nil
}

func (rl *RotatingLogger) rotate() error {
	// 关闭当前文件
	if rl.file != nil {
		rl.file.Close()
	}

	// 轮转备份文件
	for i := maxBackups - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", rl.logPath, i)
		newPath := fmt.Sprintf("%s.%d", rl.logPath, i+1)
		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}

	// 重命名当前日志文件
	backupPath := fmt.Sprintf("%s.1", rl.logPath)
	if err := os.Rename(rl.logPath, backupPath); err != nil {
		// 如果重命名失败，尝试删除旧文件
		os.Remove(rl.logPath)
	}

	// 打开新文件
	if err := rl.openFile(); err != nil {
		return err
	}

	rl.size = 0
	log.Printf("Log file rotated: %s", rl.logPath)

	return nil
}

func (rl *RotatingLogger) Close() error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.file != nil {
		return rl.file.Close()
	}
	return nil
}

// SetupRotatingLogger 设置带轮转的日志记录器
func SetupRotatingLogger(dataDir, logFile string) (*RotatingLogger, error) {
	logPath := filepath.Join(dataDir, logFile)

	// 创建数据目录
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// 创建轮转日志记录器
	rl, err := NewRotatingLogger(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create rotating logger: %w", err)
	}

	log.SetOutput(rl)

	return rl, nil
}
