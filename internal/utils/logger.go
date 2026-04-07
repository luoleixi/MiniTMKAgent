package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logger     *Logger
	loggerOnce sync.Once
)

// Logger 日志管理器
type Logger struct {
	file   *os.File
	logger *log.Logger
	mu     sync.Mutex
}

// GetLogger 获取日志实例
func GetLogger() *Logger {
	loggerOnce.Do(func() {
		logger = newLogger()
	})
	return logger
}

func newLogger() *Logger {
	// 创建日志目录
	logDir := "logs"
	os.MkdirAll(logDir, 0755)

	// 日志文件名带日期
	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return &Logger{
			file:   os.Stderr,
			logger: log.New(os.Stderr, "", log.LstdFlags),
		}
	}

	return &Logger{
		file:   file,
		logger: log.New(file, "", log.LstdFlags|log.Lmicroseconds),
	}
}

// Debug 输出调试信息（只写入日志文件）
func (l *Logger) Debug(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Println(append([]interface{}{"[DEBUG]"}, v...)...)
}

// Debugf 格式化输出调试信息
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("[DEBUG] "+format, v...)
}

// Info 输出信息（控制台+日志）
func (l *Logger) Info(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprint(v...)
	fmt.Println(msg)
	l.logger.Println(append([]interface{}{"[INFO]"}, v...)...)
}

// Infof 格式化输出信息
func (l *Logger) Infof(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf(format, v...)
	fmt.Println(msg)
	l.logger.Printf("[INFO] "+format, v...)
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	if l.file != nil && l.file != os.Stderr {
		return l.file.Close()
	}
	return nil
}

// Warn 输出警告信息（控制台+日志）
func (l *Logger) Warn(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprint(v...)
	fmt.Println(msg)
	l.logger.Println(append([]interface{}{"[WARN]"}, v...)...)
}

// Warnf 格式化输出警告信息
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf(format, v...)
	fmt.Println(msg)
	l.logger.Printf("[WARN] "+format, v...)
}

// Error 输出错误信息（控制台+日志）
func (l *Logger) Error(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprint(v...)
	fmt.Fprintln(os.Stderr, msg)
	l.logger.Println(append([]interface{}{"[ERROR]"}, v...)...)
}

// Errorf 格式化输出错误信息
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf(format, v...)
	fmt.Fprintln(os.Stderr, msg)
	l.logger.Printf("[ERROR] "+format, v...)
}

// 包级别便捷函数

// Debug 输出调试信息（只写入日志文件）
func Debug(v ...interface{}) {
	GetLogger().Debug(v...)
}

// Debugf 格式化输出调试信息
func Debugf(format string, v ...interface{}) {
	GetLogger().Debugf(format, v...)
}

// Info 输出信息（控制台+日志）
func Info(v ...interface{}) {
	GetLogger().Info(v...)
}

// Infof 格式化输出信息
func Infof(format string, v ...interface{}) {
	GetLogger().Infof(format, v...)
}

// Warn 输出警告信息（控制台+日志）
func Warn(v ...interface{}) {
	GetLogger().Warn(v...)
}

// Warnf 格式化输出警告信息
func Warnf(format string, v ...interface{}) {
	GetLogger().Warnf(format, v...)
}

// Error 输出错误信息（控制台+日志）
func Error(v ...interface{}) {
	GetLogger().Error(v...)
}

// Errorf 格式化输出错误信息
func Errorf(format string, v ...interface{}) {
	GetLogger().Errorf(format, v...)
}
