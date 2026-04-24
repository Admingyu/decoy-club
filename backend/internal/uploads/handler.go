package uploads

import (
	"net/http"

	"decoy-club/backend/internal/common/response"
	"decoy-club/backend/internal/users"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) UploadImage(c *gin.Context) {
	viewerID := c.GetString(users.ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": "missing file"})
		return
	}
	defer file.Close()

	uploaded, err := h.svc.SaveImage(c.Request.Context(), SaveImageInput{
		OwnerUserID: viewerID,
		FileName:    header.Filename,
		MimeType:    header.Header.Get("Content-Type"),
		Size:        header.Size,
		Reader:      file,
	})
	if err != nil {
		status := http.StatusBadRequest
		if err != ErrInvalidOwnerID && err != ErrInvalidUploadType && err != ErrEmptyUpload {
			status = http.StatusInternalServerError
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusCreated, UploadImageResponse{File: NewUploadedFileView(uploaded)})
}
