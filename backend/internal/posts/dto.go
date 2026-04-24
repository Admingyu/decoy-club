package posts

import "time"

type CreatePostRequest struct {
	ContentMarkdown string   `json:"content_markdown"`
	EmbeddedImages  []string `json:"embedded_images"`
}

type PostView struct {
	ID              string     `json:"id"`
	AuthorID        string     `json:"author_id"`
	ContentMarkdown string     `json:"content_markdown"`
	ContentHTML     string     `json:"content_html"`
	EmbeddedImages  []string   `json:"embedded_images"`
	LikeCount       int64      `json:"like_count"`
	CommentCount    int64      `json:"comment_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

type CreatePostResponse struct {
	Post *PostView `json:"post"`
}

type PostResponse struct {
	Post *PostView `json:"post"`
}

type PublicTimelineResponse struct {
	Posts []*PostView `json:"posts"`
}

type FollowingTimelineResponse struct {
	Posts []*PostView `json:"posts"`
}

func NewPostView(post *Post) *PostView {
	if post == nil {
		return nil
	}

	return &PostView{
		ID:              post.ID.Hex(),
		AuthorID:        post.AuthorID.Hex(),
		ContentMarkdown: post.ContentMarkdown,
		ContentHTML:     post.ContentHTML,
		EmbeddedImages:  append([]string(nil), post.EmbeddedImages...),
		LikeCount:       post.LikeCount,
		CommentCount:    post.CommentCount,
		CreatedAt:       post.CreatedAt,
		UpdatedAt:       post.UpdatedAt,
		DeletedAt:       post.DeletedAt,
	}
}

func NewPostViews(posts []*Post) []*PostView {
	if len(posts) == 0 {
		return []*PostView{}
	}

	views := make([]*PostView, 0, len(posts))
	for _, post := range posts {
		views = append(views, NewPostView(post))
	}

	return views
}
