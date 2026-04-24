package posts

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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

func (h *Handler) CreatePost(c *gin.Context) {
	viewerID := c.GetString(users.ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	post, err := h.svc.CreatePost(c.Request.Context(), CreatePostInput{
		AuthorID:        viewerID,
		ContentMarkdown: req.ContentMarkdown,
		EmbeddedImages:  req.EmbeddedImages,
	})
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusCreated, CreatePostResponse{Post: post})
}

func (h *Handler) ListPublicTimeline(c *gin.Context) {
	limit := 20
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var before *time.Time
	if raw := c.Query("before"); raw != "" {
		if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
			before = &parsed
		}
	}

	posts, err := h.svc.ListPublicTimeline(c.Request.Context(), limit, before)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusOK, PublicTimelineResponse{Posts: posts})
}

func (h *Handler) GetPost(c *gin.Context) {
	postID := c.Param("postId")
	post, err := h.svc.GetPost(c.Request.Context(), postID)
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrPostNotFound {
			status = http.StatusNotFound
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusOK, PostResponse{Post: post})
}

func (h *Handler) DeletePost(c *gin.Context) {
	viewerID := c.GetString(users.ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}

	postID := c.Param("postId")
	if err := h.svc.DeletePost(c.Request.Context(), postID, viewerID); err != nil {
		status := http.StatusBadRequest
		if err == ErrPostNotFound {
			status = http.StatusNotFound
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
