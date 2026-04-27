package users

const ContextKeyViewerID = "users.viewer_id"

type Profile struct {
	Username          string `json:"username"`
	AvatarURL         string `json:"avatar_url"`
	Bio               string `json:"bio"`
	StatusText        string `json:"status_text"`
	StatusPreset      string `json:"status_preset"`
	PostCount         int64  `json:"post_count"`
	ReplyCount        int64  `json:"reply_count"`
	FollowersCount    int64  `json:"followers_count"`
	FollowingCount    int64  `json:"following_count"`
	ReceivedLikeCount int64  `json:"received_like_count"`
	GivenLikeCount    int64  `json:"given_like_count"`
	Following         bool   `json:"following"`
}

type ProfileResponse struct {
	Profile Profile `json:"profile"`
}

type UpdateStatusRequest struct {
	StatusText   string `json:"status_text"`
	StatusPreset string `json:"status_preset"`
}
