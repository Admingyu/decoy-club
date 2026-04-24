package http_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	apphttp "decoy-club/backend/internal/http"
)

func TestRouterRegistersHealthRoute(t *testing.T) {
	router := apphttp.NewRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRouterRegistersAuthRoutes(t *testing.T) {
	router := apphttp.NewRouter(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Fatal("expected auth register route to be registered")
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{}`))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Fatal("expected auth login route to be registered")
	}
}
