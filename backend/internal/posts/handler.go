package posts

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"decoy-club/backend/internal/common/response"
	"decoy-club/backend/internal/users"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc       *Service
	usersRepo users.Repository
}

func NewHandler(svc *Service, usersRepo users.Repository) *Handler {
	return &Handler{svc: svc, usersRepo: usersRepo}
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
		status := http.StatusBadRequest
		if err == ErrInvalidAuthorID {
			status = http.StatusUnauthorized
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusCreated, CreatePostResponse{Post: h.newPostView(c.Request.Context(), post, viewerID)})
}

func (h *Handler) ListPublicTimeline(c *gin.Context) {
	limit := 20
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var before *time.Time
	beforeID := c.Query("before_id")
	if raw := c.Query("before"); raw != "" {
		if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
			before = &parsed
		}
	}

	posts, err := h.svc.ListPublicTimeline(c.Request.Context(), limit, before, beforeID)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusOK, PublicTimelineResponse{Posts: h.newPostViews(c.Request.Context(), posts, c.GetString(users.ContextKeyViewerID))})
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

	response.JSON(c, http.StatusOK, PostResponse{Post: h.newPostView(c.Request.Context(), post, c.GetString(users.ContextKeyViewerID))})
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

func (h *Handler) LikePost(c *gin.Context) {
	h.likeAction(c, h.svc.LikePost)
}

func (h *Handler) UnlikePost(c *gin.Context) {
	h.likeAction(c, h.svc.UnlikePost)
}

func (h *Handler) likeAction(c *gin.Context, action func(ctx context.Context, postID, userID string) error) {
	viewerID := c.GetString(users.ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}

	if err := action(c.Request.Context(), c.Param("postId"), viewerID); err != nil {
		status := http.StatusBadRequest
		if err == ErrPostNotFound {
			status = http.StatusNotFound
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListFollowingTimeline(c *gin.Context) {
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

	size := 20
	if raw := c.Query("size"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			size = parsed
		}
	}

	posts, err := h.svc.ListFollowingTimeline(c.Request.Context(), viewerID, page, size)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusOK, FollowingTimelineResponse{Posts: h.newPostViews(c.Request.Context(), posts, viewerID)})
}

func (h *Handler) ListPostsByUsername(c *gin.Context) {
	username := c.Param("username")
	user, err := h.usersRepo.FindByUsername(c.Request.Context(), username)
	if err != nil {
		status := http.StatusBadRequest
		if err == users.ErrUserNotFound {
			status = http.StatusNotFound
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	posts, err := h.svc.ListPostsByAuthorID(c.Request.Context(), user.ID.Hex(), 20)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authorUsernames := map[string]string{user.ID.Hex(): user.Username}
	likedPostIDs := h.likedPostIDs(c.Request.Context(), c.GetString(users.ContextKeyViewerID), posts)
	response.JSON(c, http.StatusOK, PublicTimelineResponse{Posts: NewPostViewsWithAuthorsAndLikes(posts, authorUsernames, likedPostIDs)})
}

func (h *Handler) newPostView(ctx context.Context, post *Post, viewerID string) *PostView {
	if post == nil {
		return nil
	}
	username := ""
	liked := false
	if h.usersRepo != nil {
		if user, err := h.usersRepo.FindByID(ctx, post.AuthorID.Hex()); err == nil && user != nil {
			username = user.Username
		}
	}
	if viewerID != "" {
		if likedResult, err := h.svc.repo.HasLike(ctx, post.ID.Hex(), viewerID); err == nil {
			liked = likedResult
		}
	}
	return NewPostViewWithAuthorAndLike(post, username, liked)
}

func (h *Handler) newPostViews(ctx context.Context, posts []*Post, viewerID string) []*PostView {
	if len(posts) == 0 {
		return []*PostView{}
	}
	authorUsernames := make(map[string]string)
	if h.usersRepo != nil {
		for _, post := range posts {
			authorID := post.AuthorID.Hex()
			if _, ok := authorUsernames[authorID]; ok {
				continue
			}
			if user, err := h.usersRepo.FindByID(ctx, authorID); err == nil && user != nil {
				authorUsernames[authorID] = user.Username
			}
		}
	}
	return NewPostViewsWithAuthorsAndLikes(posts, authorUsernames, h.likedPostIDs(ctx, viewerID, posts))
}

func (h *Handler) likedPostIDs(ctx context.Context, viewerID string, posts []*Post) map[string]bool {
	if viewerID == "" || len(posts) == 0 {
		return nil
	}
	postIDs := make([]string, 0, len(posts))
	for _, post := range posts {
		postIDs = append(postIDs, post.ID.Hex())
	}
	likedPostIDs, err := h.svc.repo.ListLikedPostIDs(ctx, viewerID, postIDs)
	if err != nil {
		return nil
	}
	return likedPostIDs
}
