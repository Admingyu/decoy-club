package auth

import (
	"context"
	"errors"
	"time"

	"decoy-club/backend/internal/users"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Service struct {
	repo      Repository
	jwtSecret string
}

func NewService(repo Repository, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*users.User, error) {
	hash, err := HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user := &users.User{
		Username:     input.Username,
		PasswordHash: hash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return s.repo.CreateUser(ctx, user)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (string, *users.User, error) {
	user, err := s.repo.FindByUsername(ctx, input.Username)
	if err != nil {
		return "", nil, err
	}

	if err := ComparePassword(user.PasswordHash, input.Password); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := SignJWT(s.jwtSecret, user.ID.Hex(), user.Username)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}
