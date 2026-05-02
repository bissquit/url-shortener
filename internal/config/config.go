package config

import (
	"encoding/json"
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
	TrustedSubnet   string
}

type jsonConfig struct {
	ServerAddr      string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DSN             string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
	TrustedSubnet   string `json:"trusted_subnet"`
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

	var configPath string
	flag.StringVar(&configPath, "c", "", "path to JSON config file")
	flag.StringVar(&configPath, "config", "", "path to JSON config file")
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
	flag.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet,
		"trusted subnet in CIDR format (default \"\")")
	flag.Parse()

	// get path to config file
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		configPath = envConfig
	}

	// JSON — low priority
	if configPath != "" {
		loadFromJSON(cfg, configPath)
	}

	// env rewrites JSON
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
	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		cfg.TrustedSubnet = envTrustedSubnet
	}

	return cfg
}

func loadFromJSON(cfg *Config, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var jc jsonConfig
	if err := json.Unmarshal(data, &jc); err != nil {
		return
	}

	if jc.ServerAddr != "" {
		cfg.ServerAddr = jc.ServerAddr
	}
	if jc.BaseURL != "" {
		cfg.BaseURL = jc.BaseURL
	}
	if jc.FileStoragePath != "" {
		cfg.FileStoragePath = jc.FileStoragePath
	}
	if jc.DSN != "" {
		cfg.DSN = jc.DSN
	}
	if jc.EnableHTTPS {
		cfg.EnableHTTPS = true
	}
	if jc.TrustedSubnet != "" {
		cfg.TrustedSubnet = jc.TrustedSubnet
	}
}
