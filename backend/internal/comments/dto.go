package comments

import "time"

const DeletedCommentText = "内容已删除"

type CreateCommentRequest struct {
	ContentMarkdown string `json:"content_markdown"`
}

type CommentView struct {
	ID              string         `json:"id"`
	PostID          string         `json:"post_id"`
	AuthorID        string         `json:"author_id"`
	AuthorUsername  string         `json:"author_username,omitempty"`
	ContentMarkdown string         `json:"content_markdown"`
	ContentHTML     string         `json:"content_html"`
	ParentCommentID *string        `json:"parent_comment_id,omitempty"`
	ReplyToUserID   *string        `json:"reply_to_user_id,omitempty"`
	IsDeleted       bool           `json:"is_deleted"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       *time.Time     `json:"deleted_at,omitempty"`
	Replies         []*CommentView `json:"replies"`
}

type CommentResponse struct {
	Comment *CommentView `json:"comment"`
}

type CommentListResponse struct {
	Comments []*CommentView `json:"comments"`
}

func NewCommentView(comment *Comment) *CommentView {
	return NewCommentViewWithAuthor(comment, "")
}

func NewCommentViewWithAuthor(comment *Comment, authorUsername string) *CommentView {
	if comment == nil {
		return nil
	}

	view := &CommentView{
		ID:              comment.ID.Hex(),
		PostID:          comment.PostID.Hex(),
		AuthorID:        comment.AuthorID.Hex(),
		AuthorUsername:  authorUsername,
		ContentMarkdown: comment.ContentMarkdown,
		ContentHTML:     comment.ContentHTML,
		IsDeleted:       comment.IsDeleted,
		CreatedAt:       comment.CreatedAt,
		UpdatedAt:       comment.UpdatedAt,
		DeletedAt:       comment.DeletedAt,
		Replies:         []*CommentView{},
	}

	if comment.ParentCommentID != nil {
		parentID := comment.ParentCommentID.Hex()
		view.ParentCommentID = &parentID
	}
	if comment.ReplyToUserID != nil {
		replyToUserID := comment.ReplyToUserID.Hex()
		view.ReplyToUserID = &replyToUserID
	}
	if comment.IsDeleted {
		view.ContentMarkdown = DeletedCommentText
		view.ContentHTML = "<p>" + DeletedCommentText + "</p>"
	}

	return view
}

func BuildCommentTree(comments []*Comment) []*CommentView {
	return BuildCommentTreeWithAuthors(comments, nil)
}

func BuildCommentTreeWithAuthors(comments []*Comment, authorUsernames map[string]string) []*CommentView {
	if len(comments) == 0 {
		return []*CommentView{}
	}

	roots := make([]*CommentView, 0)
	index := make(map[string]*CommentView, len(comments))

	for _, comment := range comments {
		username := ""
		if authorUsernames != nil {
			username = authorUsernames[comment.AuthorID.Hex()]
		}
		view := NewCommentViewWithAuthor(comment, username)
		index[view.ID] = view
		if comment.ParentCommentID == nil {
			roots = append(roots, view)
			continue
		}

		parent := index[comment.ParentCommentID.Hex()]
		if parent == nil {
			roots = append(roots, view)
			continue
		}
		parent.Replies = append(parent.Replies, view)
	}

	return roots
}
