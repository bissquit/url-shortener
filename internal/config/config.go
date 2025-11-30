package config

import "flag"

type Config struct {
	ServerAddr string
	BaseURL    string
}

func New() *Config {
	return &Config{
		ServerAddr: ":8080",
		BaseURL:    "http://localhost:8080",
	}
}

func (c *Config) ParseFlags() {
	flag.StringVar(&c.ServerAddr, "a", c.ServerAddr, "server address in host:port format (default :8080)")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "base URL (default http://localhost:8080)")
	flag.Parse()
}
