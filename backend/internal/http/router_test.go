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

func TestRouterRegistersCommentRoutes(t *testing.T) {
	router := apphttp.NewRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/507f1f77bcf86cd799439011/comments", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code == http.StatusNotFound {
		t.Fatal("expected list comments route to be registered")
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/comments/507f1f77bcf86cd799439011/replies", bytes.NewBufferString(`{}`))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code == http.StatusNotFound {
		t.Fatal("expected reply route to be registered")
	}
}

func TestRouterRegistersNotificationRoutes(t *testing.T) {
	router := apphttp.NewRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/unread-count", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code == http.StatusNotFound {
		t.Fatal("expected unread-count route to be registered")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code == http.StatusNotFound {
		t.Fatal("expected notifications list route to be registered")
	}
}

func TestRouterRegistersUploadRoute(t *testing.T) {
	router := apphttp.NewRouter(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/uploads/images", bytes.NewBufferString(""))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code == http.StatusNotFound {
		t.Fatal("expected upload route to be registered")
	}
}
