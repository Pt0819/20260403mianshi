package config

import (
	"time"
)

// Config 应用配置
type Config struct {
	// 扫描配置
	Timeout    time.Duration
	Workers    int
	MaxRetries int
	Verbose    bool

	// 输出配置
	OutputFormat string
	OutputFile   string

	// 网络配置
	NetworkInterface string
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Timeout:          5 * time.Second,
		Workers:          50,
		MaxRetries:       3,
		Verbose:          false,
		OutputFormat:     "yaml",
		OutputFile:       "",
		NetworkInterface: "",
	}
}

// NewConfig 创建新的配置
func NewConfig(timeout int, workers int, verbose bool, outputFile string) *Config {
	cfg := DefaultConfig()
	cfg.Timeout = time.Duration(timeout) * time.Second
	cfg.Workers = workers
	cfg.Verbose = verbose
	cfg.OutputFile = outputFile

	return cfg
}
