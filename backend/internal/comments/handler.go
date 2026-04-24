package comments

import (
	"context"
	"encoding/json"
	"net/http"

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

func (h *Handler) ListPostComments(c *gin.Context) {
	comments, err := h.svc.ListCommentsByPostID(c.Request.Context(), c.Param("postId"))
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusOK, CommentListResponse{Comments: h.buildCommentTree(c.Request.Context(), comments)})
}

func (h *Handler) CreatePostComment(c *gin.Context) {
	viewerID := c.GetString(users.ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	comment, err := h.svc.CreateComment(c.Request.Context(), CreateCommentInput{
		PostID:          c.Param("postId"),
		AuthorID:        viewerID,
		ContentMarkdown: req.ContentMarkdown,
	})
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusCreated, CommentResponse{Comment: h.newCommentView(c.Request.Context(), comment)})
}

func (h *Handler) ReplyToComment(c *gin.Context) {
	viewerID := c.GetString(users.ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	comment, err := h.svc.ReplyToComment(c.Request.Context(), c.Param("commentId"), viewerID, req.ContentMarkdown)
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrCommentNotFound {
			status = http.StatusNotFound
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusCreated, CommentResponse{Comment: h.newCommentView(c.Request.Context(), comment)})
}

func (h *Handler) DeleteComment(c *gin.Context) {
	viewerID := c.GetString(users.ContextKeyViewerID)
	if viewerID == "" {
		response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing viewer identity"})
		return
	}

	err := h.svc.DeleteComment(c.Request.Context(), c.Param("commentId"), viewerID)
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrCommentNotFound {
			status = http.StatusNotFound
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) newCommentView(ctx context.Context, comment *Comment) *CommentView {
	if comment == nil {
		return nil
	}
	username := ""
	if h.usersRepo != nil {
		if user, err := h.usersRepo.FindByID(ctx, comment.AuthorID.Hex()); err == nil && user != nil {
			username = user.Username
		}
	}
	return NewCommentViewWithAuthor(comment, username)
}

func (h *Handler) buildCommentTree(ctx context.Context, comments []*Comment) []*CommentView {
	if len(comments) == 0 {
		return []*CommentView{}
	}
	authorUsernames := make(map[string]string)
	if h.usersRepo != nil {
		for _, comment := range comments {
			authorID := comment.AuthorID.Hex()
			if _, ok := authorUsernames[authorID]; ok {
				continue
			}
			if user, err := h.usersRepo.FindByID(ctx, authorID); err == nil && user != nil {
				authorUsernames[authorID] = user.Username
			}
		}
	}
	return BuildCommentTreeWithAuthors(comments, authorUsernames)
}
