package config

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Port           string `yaml:"port"`
	MongoURI       string `yaml:"mongo_uri"`
	DatabaseName   string `yaml:"database_name"`
	JWTSecret      string `yaml:"jwt_secret"`
	PublicBaseURL  string `yaml:"public_base_url"`
	UploadDir      string `yaml:"upload_dir"`
	FrontendOrigin string `yaml:"frontend_origin"`
}

func Load() (Config, error) {
	return LoadFile(defaultConfigPath())
}

func LoadFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg := defaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func defaultConfigPath() string {
	candidates := []string{
		"config.yml",
		filepath.Join("backend", "config.yml"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "config.yml"
}

func defaultConfig() Config {
	return Config{
		Port:           ":8080",
		MongoURI:       "mongodb://localhost:27017",
		DatabaseName:   "decoy_club",
		JWTSecret:      "dev-jwt-secret",
		PublicBaseURL:  "http://localhost:3000",
		UploadDir:      "./uploads",
		FrontendOrigin: "http://localhost:3000",
	}
}
