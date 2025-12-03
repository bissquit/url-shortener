package config

import "flag"

type Config struct {
	ServerAddr string
	BaseURL    string
}

func GetDefaultConfig() *Config {
	return &Config{
		ServerAddr: ":8080",
		BaseURL:    "http://localhost:8080",
	}
}

func GetConfig() *Config {
	cfg := GetDefaultConfig()

	flag.StringVar(&cfg.ServerAddr, "a", cfg.ServerAddr,
		"server address in host:port format (default :8080)")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL,
		"base URL (default http://localhost:8080)")
	flag.Parse()

	return cfg
}
