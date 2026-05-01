package posts

import (
	"time"

	"decoy-club/backend/internal/users"
)

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
	Topics          []string   `json:"topics"`
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

type TopicView struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	PostCount  int64     `json:"post_count"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastPostAt time.Time `json:"last_post_at"`
}

type TrendingTopicsResponse struct {
	Topics []*TopicView `json:"topics"`
}

type PostLikerView struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type PostLikeDetailResponse struct {
	Likers []*PostLikerView `json:"likers"`
	Total  int64            `json:"total"`
}

func NewPostLikerView(user *users.User) *PostLikerView {
	if user == nil {
		return nil
	}
	return &PostLikerView{
		ID:        user.ID.Hex(),
		Username:  user.Username,
		AvatarURL: user.AvatarURL,
	}
}

func NewPostLikerViews(likers []*users.User) []*PostLikerView {
	if len(likers) == 0 {
		return []*PostLikerView{}
	}
	views := make([]*PostLikerView, 0, len(likers))
	for _, liker := range likers {
		views = append(views, NewPostLikerView(liker))
	}
	return views
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
	topics := append([]string{}, post.Topics...)

	return &PostView{
		ID:              post.ID.Hex(),
		AuthorID:        post.AuthorID.Hex(),
		AuthorUsername:  authorUsername,
		ContentMarkdown: post.ContentMarkdown,
		ContentHTML:     post.ContentHTML,
		EmbeddedImages:  embeddedImages,
		Topics:          topics,
		LikeCount:       post.LikeCount,
		LikedByViewer:   likedByViewer,
		CommentCount:    post.CommentCount,
		CreatedAt:       post.CreatedAt,
		UpdatedAt:       post.UpdatedAt,
		DeletedAt:       post.DeletedAt,
	}
}

func NewTopicView(topic *Topic) *TopicView {
	if topic == nil {
		return nil
	}
	return &TopicView{
		ID:         topic.ID.Hex(),
		Name:       topic.Name,
		PostCount:  topic.PostCount,
		UpdatedAt:  topic.UpdatedAt,
		LastPostAt: topic.LastPostAt,
	}
}

func NewTopicViews(topics []*Topic) []*TopicView {
	if len(topics) == 0 {
		return []*TopicView{}
	}
	views := make([]*TopicView, 0, len(topics))
	for _, topic := range topics {
		views = append(views, NewTopicView(topic))
	}
	return views
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
