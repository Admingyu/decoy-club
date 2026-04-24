package posts

import "time"

type CreatePostRequest struct {
	ContentMarkdown string   `json:"content_markdown"`
	EmbeddedImages  []string `json:"embedded_images"`
}

type PostView struct {
	ID              string     `json:"id"`
	AuthorID        string     `json:"author_id"`
	AuthorUsername  string     `json:"author_username,omitempty"`
	ContentMarkdown string     `json:"content_markdown"`
	ContentHTML     string     `json:"content_html"`
	EmbeddedImages  []string   `json:"embedded_images"`
	LikeCount       int64      `json:"like_count"`
	LikedByViewer   bool       `json:"liked_by_viewer"`
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
	return NewPostViewWithAuthorAndLike(post, "", false)
}

func NewPostViewWithAuthor(post *Post, authorUsername string) *PostView {
	return NewPostViewWithAuthorAndLike(post, authorUsername, false)
}

func NewPostViewWithAuthorAndLike(post *Post, authorUsername string, likedByViewer bool) *PostView {
	if post == nil {
		return nil
	}

	embeddedImages := append([]string{}, post.EmbeddedImages...)

	return &PostView{
		ID:              post.ID.Hex(),
		AuthorID:        post.AuthorID.Hex(),
		AuthorUsername:  authorUsername,
		ContentMarkdown: post.ContentMarkdown,
		ContentHTML:     post.ContentHTML,
		EmbeddedImages:  embeddedImages,
		LikeCount:       post.LikeCount,
		LikedByViewer:   likedByViewer,
		CommentCount:    post.CommentCount,
		CreatedAt:       post.CreatedAt,
		UpdatedAt:       post.UpdatedAt,
		DeletedAt:       post.DeletedAt,
	}
}

func NewPostViews(posts []*Post) []*PostView {
	return NewPostViewsWithAuthorsAndLikes(posts, nil, nil)
}

func NewPostViewsWithAuthors(posts []*Post, authorUsernames map[string]string) []*PostView {
	return NewPostViewsWithAuthorsAndLikes(posts, authorUsernames, nil)
}

func NewPostViewsWithAuthorsAndLikes(posts []*Post, authorUsernames map[string]string, likedPostIDs map[string]bool) []*PostView {
	if len(posts) == 0 {
		return []*PostView{}
	}

	views := make([]*PostView, 0, len(posts))
	for _, post := range posts {
		username := ""
		if authorUsernames != nil {
			username = authorUsernames[post.AuthorID.Hex()]
		}
		liked := false
		if likedPostIDs != nil {
			liked = likedPostIDs[post.ID.Hex()]
		}
		views = append(views, NewPostViewWithAuthorAndLike(post, username, liked))
	}

	return views
}
