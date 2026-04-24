package auth

import (
	"encoding/json"
	"net/http"

	"decoy-club/backend/internal/common/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(c *gin.Context) {
	var input RegisterInput
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.svc.Register(c.Request.Context(), input)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusCreated, gin.H{"user": user})
}

func (h *Handler) Login(c *gin.Context) {
	var input LoginInput
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		response.JSON(c, http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	token, user, err := h.svc.Login(c.Request.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrInvalidCredentials || err == ErrUserNotFound {
			status = http.StatusUnauthorized
		}
		response.JSON(c, status, gin.H{"error": err.Error()})
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}
