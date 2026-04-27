package notifications

import (
	"encoding/json"
	"net/http"
	"strconv"

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

func (h *Handler) GetUnreadCount(c *gin.Context) {
	viewerID := c.GetString(users.ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}

	counts, total, err := h.svc.GetUnreadCount(c.Request.Context(), viewerID)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	byType := make(map[string]int64, len(counts))
	for k, v := range counts {
		byType[string(k)] = v
	}
	response.JSON(c, http.StatusOK, UnreadCountResponse{Total: total, ByType: byType})
}

func (h *Handler) ListNotifications(c *gin.Context) {
	viewerID := c.GetString(users.ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}
	page := 1
	if raw := c.Query("page"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			page = parsed
		}
	}
	pageSize := 20
	if raw := c.Query("page_size"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	unreadOnly := c.Query("status") == "unread"
	types := notificationTypesForFilter(c.Query("type"))

	notifications, err := h.svc.ListNotifications(c.Request.Context(), viewerID, unreadOnly, types, page, pageSize)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response.JSON(c, http.StatusOK, NotificationListResponse{Notifications: newNotificationListItems(notifications, false)})
}

func notificationTypesForFilter(filter string) []NotificationType {
	switch filter {
	case "likes":
		return []NotificationType{TypePostLiked}
	case "comments":
		return []NotificationType{TypePostCommented, TypeCommentReplied}
	case "mentions":
		return []NotificationType{TypeUserMentioned}
	default:
		return nil
	}
}

func (h *Handler) GetNotification(c *gin.Context) {
	viewerID := c.GetString(users.ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}

	notification, err := h.svc.GetNotification(c.Request.Context(), viewerID, c.Param("notificationId"))
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrNotificationNotFound {
			status = http.StatusNotFound
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}
	response.JSON(c, http.StatusOK, NotificationDetailResponse{Notification: newNotificationListItem(notification, true)})
}

func (h *Handler) MarkRead(c *gin.Context) {
	viewerID := c.GetString(users.ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}

	var req MarkReadRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), viewerID, req.NotificationIDs); err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
