package posts

type CreatePostRequest struct {
	ContentMarkdown string   `json:"content_markdown"`
	EmbeddedImages  []string `json:"embedded_images"`
}

type CreatePostResponse struct {
	Post *Post `json:"post"`
}

type PostResponse struct {
	Post *Post `json:"post"`
}

type PublicTimelineResponse struct {
	Posts []*Post `json:"posts"`
}
