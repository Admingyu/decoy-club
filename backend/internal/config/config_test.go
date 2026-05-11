package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsConfigYMLInsteadOfEnvironment(t *testing.T) {
	dir := t.TempDir()
	writeConfigFile(t, filepath.Join(dir, "config.yml"))
	chdir(t, dir)

	t.Setenv("PORT", ":1111")
	t.Setenv("MONGO_URI", "mongodb://env-host:27017")
	t.Setenv("DATABASE_NAME", "env_database")
	t.Setenv("JWT_SECRET", "env-secret")
	t.Setenv("PUBLIC_BASE_URL", "http://env.example")
	t.Setenv("UPLOAD_DIR", "./env-uploads")
	t.Setenv("FRONTEND_ORIGIN", "http://env-frontend.example")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	assertConfig(t, cfg)
}

func TestLoadFindsBackendConfigYMLFromRepositoryRoot(t *testing.T) {
	dir := t.TempDir()
	backendDir := filepath.Join(dir, "backend")
	if err := os.Mkdir(backendDir, 0o755); err != nil {
		t.Fatalf("create backend dir: %v", err)
	}
	writeConfigFile(t, filepath.Join(backendDir, "config.yml"))
	chdir(t, dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	assertConfig(t, cfg)
}

func TestLoadDefaultsPublicBaseURLToFrontendProxyOrigin(t *testing.T) {
	dir := t.TempDir()
	content := []byte(`port: ":9090"
mongo_uri: mongodb://config-host:27017
database_name: config_database
jwt_secret: config-secret
upload_dir: ./config-uploads
frontend_origin: http://localhost:3000
`)
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), content, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	chdir(t, dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.PublicBaseURL != "http://localhost:3000" {
		t.Fatalf("public base url mismatch: got %q, want %q", cfg.PublicBaseURL, "http://localhost:3000")
	}
}

func writeConfigFile(t *testing.T, path string) {
	t.Helper()

	content := []byte(`port: ":9090"
mongo_uri: mongodb://config-host:27017
database_name: config_database
jwt_secret: config-secret
public_base_url: http://config.example
upload_dir: ./config-uploads
frontend_origin: http://config-frontend.example
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
}

func assertConfig(t *testing.T, cfg Config) {
	t.Helper()

	expected := Config{
		Port:           ":9090",
		MongoURI:       "mongodb://config-host:27017",
		DatabaseName:   "config_database",
		JWTSecret:      "config-secret",
		PublicBaseURL:  "http://config.example",
		UploadDir:      "./config-uploads",
		FrontendOrigin: "http://config-frontend.example",
	}
	if cfg != expected {
		t.Fatalf("config mismatch\n got: %#v\nwant: %#v", cfg, expected)
	}
}
