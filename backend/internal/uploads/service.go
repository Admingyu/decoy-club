package uploads

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrInvalidOwnerID = errors.New("invalid owner id")
var ErrInvalidUploadType = errors.New("only image uploads are supported")
var ErrEmptyUpload = errors.New("file is empty")

type Service struct {
	repo          Repository
	publicBaseURL string
	uploadDir     string
}

func NewService(repo Repository, publicBaseURL, uploadDir string) *Service {
	return &Service{
		repo:          repo,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
		uploadDir:     uploadDir,
	}
}

type SaveImageInput struct {
	OwnerUserID string
	FileName    string
	MimeType    string
	Size        int64
	Reader      io.Reader
}

func (s *Service) SaveImage(ctx context.Context, input SaveImageInput) (*UploadedFile, error) {
	ownerID, err := bson.ObjectIDFromHex(input.OwnerUserID)
	if err != nil {
		return nil, ErrInvalidOwnerID
	}
	if input.Reader == nil || input.Size == 0 {
		return nil, ErrEmptyUpload
	}
	if !strings.HasPrefix(strings.ToLower(input.MimeType), "image/") {
		return nil, ErrInvalidUploadType
	}

	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return nil, err
	}

	extension := filepath.Ext(input.FileName)
	if extension == "" {
		extension = mimeExtension(input.MimeType)
	}
	fileID := bson.NewObjectID().Hex()
	relativePath := filepath.ToSlash(filepath.Join("images", time.Now().UTC().Format("20060102"), fileID+extension))
	absolutePath := filepath.Join(s.uploadDir, relativePath)
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return nil, err
	}

	dst, err := os.Create(absolutePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	written, err := io.Copy(dst, input.Reader)
	if err != nil {
		return nil, err
	}
	if written == 0 {
		return nil, ErrEmptyUpload
	}

	uploadedFile := &UploadedFile{
		OwnerUserID: ownerID,
		FileName:    filepath.Base(input.FileName),
		MimeType:    input.MimeType,
		Size:        written,
		StoragePath: absolutePath,
		PublicURL:   s.publicBaseURL + "/uploads/" + relativePath,
		CreatedAt:   time.Now().UTC(),
	}

	return s.repo.CreateUploadedFile(ctx, uploadedFile)
}

func mimeExtension(mimeType string) string {
	switch strings.ToLower(mimeType) {
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}
