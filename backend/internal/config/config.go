package config

import "os"

type Config struct {
	Port           string
	MongoURI       string
	DatabaseName   string
	JWTSecret      string
	PublicBaseURL  string
	UploadDir      string
	FrontendOrigin string
}

func Load() Config {
	return Config{
		Port:           getEnv("PORT", ":8080"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName:   getEnv("DATABASE_NAME", "decoy_club"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-jwt-secret"),
		PublicBaseURL:  getEnv("PUBLIC_BASE_URL", "http://localhost:8080"),
		UploadDir:      getEnv("UPLOAD_DIR", "./uploads"),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
