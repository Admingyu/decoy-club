package users

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"decoy-club/backend/internal/common/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetProfile(c *gin.Context) {
	username := c.Param("username")
	viewerID := c.GetString(ContextKeyViewerID)

	profile, err := h.svc.GetProfile(c.Request.Context(), username, viewerID)
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrUserNotFound {
			status = http.StatusNotFound
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusOK, ProfileResponse{Profile: *profile})
}

func (h *Handler) SearchUsers(c *gin.Context) {
	viewerID := c.GetString(ContextKeyViewerID)
	limit := DefaultSearchLimit
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			response.JSON(c, http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
		limit = parsedLimit
	}

	users, err := h.svc.SearchUsers(c.Request.Context(), c.Query("q"), viewerID, limit)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusOK, SearchUsersResponse{Users: users})
}

func (h *Handler) Follow(c *gin.Context) {
	h.followAction(c, h.svc.Follow)
}

func (h *Handler) Unfollow(c *gin.Context) {
	h.followAction(c, h.svc.Unfollow)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	viewerID := c.GetString(ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	profile, err := h.svc.UpdateStatus(c.Request.Context(), viewerID, req.StatusText, req.StatusPreset)
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrUserNotFound {
			status = http.StatusNotFound
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusOK, ProfileResponse{Profile: *profile})
}

func (h *Handler) followAction(c *gin.Context, action func(ctx context.Context, followerID, followeeID string) error) {
	viewerID := c.GetString(ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}

	username := c.Param("username")
	followee, err := h.svc.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrUserNotFound {
			status = http.StatusNotFound
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	if err := action(c.Request.Context(), viewerID, followee.ID.Hex()); err != nil {
		status := http.StatusBadRequest
		if err == ErrCannotFollowSelf {
			status = http.StatusBadRequest
		} else if err == ErrUserNotFound {
			status = http.StatusNotFound
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.svc.GetProfileForUser(c.Request.Context(), followee, viewerID)
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrUserNotFound {
			status = http.StatusNotFound
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusOK, ProfileResponse{Profile: *profile})
}
