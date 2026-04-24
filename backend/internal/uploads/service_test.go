package uploads

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestSaveImageStoresFileAndReturnsPublicURL(t *testing.T) {
	repo := NewMemoryRepository()
	dir := t.TempDir()
	svc := NewService(repo, "http://localhost:8080", dir)

	file, err := svc.SaveImage(context.Background(), SaveImageInput{
		OwnerUserID: "507f1f77bcf86cd799439011",
		FileName:    "avatar.png",
		MimeType:    "image/png",
		Size:        int64(len("png-data")),
		Reader:      strings.NewReader("png-data"),
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if file.PublicURL == "" {
		t.Fatal("expected public url")
	}
	if _, err := os.Stat(file.StoragePath); err != nil {
		t.Fatalf("expected file on disk, got %v", err)
	}
}

func TestSaveImageRejectsNonImages(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo, "http://localhost:8080", t.TempDir())

	_, err := svc.SaveImage(context.Background(), SaveImageInput{
		OwnerUserID: "507f1f77bcf86cd799439011",
		FileName:    "notes.txt",
		MimeType:    "text/plain",
		Size:        int64(len("hello")),
		Reader:      strings.NewReader("hello"),
	})
	if err != ErrInvalidUploadType {
		t.Fatalf("expected ErrInvalidUploadType, got %v", err)
	}
}
