package activities

import (
	"time"

	"decoy-club/backend/internal/comments"
	"decoy-club/backend/internal/posts"
)

type CountsResponse struct {
	Counts Counts `json:"counts"`
}

type ActivityView struct {
	ID        string                `json:"id"`
	Type      ActivityType          `json:"type"`
	Count     int64                 `json:"count"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	Post      *posts.PostView       `json:"post,omitempty"`
	Comment   *comments.CommentView `json:"comment,omitempty"`
}

type ActivityListResponse struct {
	Activities []*ActivityView `json:"activities"`
}
