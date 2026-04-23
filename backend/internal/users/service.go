package users

import (
	"context"
	"errors"
)

var ErrCannotFollowSelf = errors.New("cannot follow self")

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

	profile := &Profile{
		Username:          user.Username,
		AvatarURL:         user.AvatarURL,
		Bio:               user.Bio,
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
