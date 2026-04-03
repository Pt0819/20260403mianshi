package utils

import (
	"fmt"
	"os"
	"time"
)

// ProgressBar 简单的进度条
type ProgressBar struct {
	total     int
	current   int
	startTime time.Time
}

// NewProgressBar 创建新的进度条
func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		total:     total,
		current:   0,
		startTime: time.Now(),
	}
}

// Increment 增加进度
func (p *ProgressBar) Increment() {
	p.current++
}

// Print 打印进度
func (p *ProgressBar) Print() {
	percent := float64(p.current) / float64(p.total) * 100
	elapsed := time.Since(p.startTime)

	fmt.Printf("\rProgress: %d/%d (%.1f%%) - Elapsed: %s",
		p.current, p.total, percent, elapsed.Round(time.Second))

	if p.current >= p.total {
		fmt.Println()
	}
}

// Logger 简单的日志记录器
type Logger struct {
	verbose bool
}

// NewLogger 创建新的日志记录器
func NewLogger(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

// Info 打印信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

// Debug 打印调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.verbose {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// Error 打印错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
}

// Warn 打印警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}