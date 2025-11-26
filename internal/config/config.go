package config

type Config struct {
	BaseURL string
}

func NewConfig() *Config {
	return &Config{
		BaseURL: "http://localhost:8080",
	}
}
