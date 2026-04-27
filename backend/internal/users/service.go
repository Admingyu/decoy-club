package users

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

var ErrCannotFollowSelf = errors.New("cannot follow self")
var ErrStatusTooLong = errors.New("status text is too long")

const (
	MaxStatusTextLength   = 80
	MaxStatusPresetLength = 32
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return s.repo.FindByUsername(ctx, username)
}

func (s *Service) GetProfile(ctx context.Context, username, viewerID string) (*Profile, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return s.profileFromUser(ctx, user, viewerID)
}

func (s *Service) GetProfileForUser(ctx context.Context, user *User, viewerID string) (*Profile, error) {
	if user == nil {
		return nil, ErrUserNotFound
	}

	return s.profileFromUser(ctx, user, viewerID)
}

func (s *Service) profileFromUser(ctx context.Context, user *User, viewerID string) (*Profile, error) {
	profile := &Profile{
		Username:          user.Username,
		AvatarURL:         user.AvatarURL,
		Bio:               user.Bio,
		StatusText:        user.StatusText,
		StatusPreset:      user.StatusPreset,
		PostCount:         user.PostCount,
		ReplyCount:        user.ReplyCount,
		FollowersCount:    user.FollowersCount,
		FollowingCount:    user.FollowingCount,
		ReceivedLikeCount: user.ReceivedLikeCount,
		GivenLikeCount:    user.GivenLikeCount,
	}

	if viewerID != "" && viewerID != user.ID.Hex() {
		following, err := s.repo.IsFollowing(ctx, viewerID, user.ID.Hex())
		if err != nil {
			return nil, err
		}
		profile.Following = following
	}

	return profile, nil
}

func (s *Service) Follow(ctx context.Context, followerID, followeeID string) error {
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}

	return s.repo.RunFollowTransaction(ctx, followerID, followeeID)
}

func (s *Service) Unfollow(ctx context.Context, followerID, followeeID string) error {
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}

	return s.repo.RunUnfollowTransaction(ctx, followerID, followeeID)
}

func (s *Service) UpdateStatus(ctx context.Context, userID, statusText, statusPreset string) (*Profile, error) {
	statusText = strings.TrimSpace(statusText)
	statusPreset = strings.TrimSpace(statusPreset)
	if utf8.RuneCountInString(statusText) > MaxStatusTextLength || utf8.RuneCountInString(statusPreset) > MaxStatusPresetLength {
		return nil, ErrStatusTooLong
	}

	user, err := s.repo.UpdateStatus(ctx, userID, statusText, statusPreset)
	if err != nil {
		return nil, err
	}

	return s.profileFromUser(ctx, user, userID)
}
