package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr      string
	BaseURL         string
	FileStoragePath string
}

func GetDefaultConfig() *Config {
	return &Config{
		ServerAddr:      ":8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "./storage.json",
	}
}

func GetConfig() *Config {
	cfg := GetDefaultConfig()

	flag.StringVar(&cfg.ServerAddr, "a", cfg.ServerAddr,
		"server address in host:port format (default :8080)")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL,
		"base URL (default http://localhost:8080)")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath,
		"file storage path (default './storage.json')")
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

	return cfg
}
