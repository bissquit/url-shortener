package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr      string
	BaseURL         string
	FileStoragePath string
	DSN             string
	AuditFilePath   string
	AuditURL        string
	EnableHTTPS     bool
}

func GetDefaultConfig() *Config {
	return &Config{
		ServerAddr:      ":8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
		DSN:             "",
	}
}

func GetConfig() *Config {
	cfg := GetDefaultConfig()

	flag.StringVar(&cfg.ServerAddr, "a", cfg.ServerAddr,
		"server address in host:port format (default :8080)")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL,
		"base URL (default http://localhost:8080)")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath,
		"file storage path (default \"\")")
	flag.StringVar(&cfg.DSN, "d", cfg.DSN,
		"Database DSN (default \"\")")
	flag.StringVar(&cfg.AuditFilePath, "audit-file", cfg.AuditFilePath,
		"audit file path (default \"\")")
	flag.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL,
		"audit URL (default \"\")")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS,
		"enable HTTPS (default false)")
	flag.Parse()

	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		cfg.ServerAddr = envServerAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}
	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		cfg.DSN = envDSN
	}
	if envAuditFilePath := os.Getenv("AUDIT_FILE"); envAuditFilePath != "" {
		cfg.AuditFilePath = envAuditFilePath
	}
	if envAuditURL := os.Getenv("AUDIT_URL"); envAuditURL != "" {
		cfg.AuditURL = envAuditURL
	}
	if os.Getenv("ENABLE_HTTPS") != "" {
		cfg.EnableHTTPS = true
	}

	return cfg
}
