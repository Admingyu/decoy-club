package activities

import (
	"net/http"
	"strconv"

	"decoy-club/backend/internal/comments"
	"decoy-club/backend/internal/common/response"
	"decoy-club/backend/internal/posts"
	"decoy-club/backend/internal/users"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc          *Service
	usersRepo    users.Repository
	postsRepo    posts.Repository
	commentsRepo comments.Repository
}

func NewHandler(svc *Service, usersRepo users.Repository, postsRepo posts.Repository, commentsRepo comments.Repository) *Handler {
	return &Handler{svc: svc, usersRepo: usersRepo, postsRepo: postsRepo, commentsRepo: commentsRepo}
}

func (h *Handler) GetCounts(c *gin.Context) {
	user, err := h.userFromParam(c)
	if err != nil {
		h.writeUserError(c, err)
		return
	}

	counts, err := h.svc.Counts(c.Request.Context(), user.ID.Hex())
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusOK, CountsResponse{Counts: counts})
}

func (h *Handler) List(c *gin.Context) {
	viewerID := c.GetString(users.ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}

	user, err := h.userFromParam(c)
	if err != nil {
		h.writeUserError(c, err)
		return
	}
	if user.ID.Hex() != viewerID {
		response.JSON(c, http.StatusForbidden, gin.H{"error": "activity records are private"})
		return
	}

	activityType, err := ParseActivityType(c.Query("type"))
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := 20
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	activities, err := h.svc.List(c.Request.Context(), viewerID, activityType, limit)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusOK, ActivityListResponse{Activities: h.newActivityViews(c, activities, viewerID)})
}

func (h *Handler) userFromParam(c *gin.Context) (*users.User, error) {
	return h.usersRepo.FindByUsername(c.Request.Context(), c.Param("username"))
}

func (h *Handler) writeUserError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	if err == users.ErrUserNotFound {
		status = http.StatusNotFound
	}
	response.JSON(c, status, gin.H{"error": err.Error()})
}

func (h *Handler) newActivityViews(c *gin.Context, activities []*Activity, viewerID string) []*ActivityView {
	if len(activities) == 0 {
		return []*ActivityView{}
	}

	views := make([]*ActivityView, 0, len(activities))
	for _, activity := range activities {
		view := &ActivityView{
			ID:        activity.ID.Hex(),
			Type:      activity.Type,
			Count:     activity.Count,
			CreatedAt: activity.CreatedAt,
			UpdatedAt: activity.UpdatedAt,
		}
		if h.postsRepo != nil {
			if post, err := h.postsRepo.FindPostByID(c.Request.Context(), activity.PostID.Hex()); err == nil && post != nil {
				authorUsername := ""
				if h.usersRepo != nil {
					if author, err := h.usersRepo.FindByID(c.Request.Context(), post.AuthorID.Hex()); err == nil && author != nil {
						authorUsername = author.Username
					}
				}
				liked := false
				if viewerID != "" {
					if likedResult, err := h.postsRepo.HasLike(c.Request.Context(), post.ID.Hex(), viewerID); err == nil {
						liked = likedResult
					}
				}
				view.Post = posts.NewPostViewWithAuthorAndLike(post, authorUsername, liked)
			}
		}
		if activity.CommentID != nil && h.commentsRepo != nil {
			if comment, err := h.commentsRepo.FindCommentByID(c.Request.Context(), activity.CommentID.Hex()); err == nil && comment != nil {
				authorUsername := ""
				if h.usersRepo != nil {
					if author, err := h.usersRepo.FindByID(c.Request.Context(), comment.AuthorID.Hex()); err == nil && author != nil {
						authorUsername = author.Username
					}
				}
				liked := false
				if viewerID != "" {
					if likedResult, err := h.commentsRepo.HasLike(c.Request.Context(), comment.ID.Hex(), viewerID); err == nil {
						liked = likedResult
					}
				}
				view.Comment = comments.NewCommentViewWithAuthorAndLike(comment, authorUsername, liked)
			}
		}
		views = append(views, view)
	}

	return views
}
