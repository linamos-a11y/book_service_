package service

import (
	"context"

	"github.com/fiverr/book_service_/internal/repository/redis"
)

// ActivityService holds the business rules for the per-user activity log
// and has no knowledge of HTTP.
type ActivityService interface {
	RecordAction(ctx context.Context, username, method, path string) error
	GetActions(ctx context.Context, username string) ([]string, error)
}

type activityService struct {
	repo redis.ActivityRepository
}

// NewActivityService builds an ActivityService backed by the given
// ActivityRepository.
func NewActivityService(repo redis.ActivityRepository) ActivityService {
	return &activityService{repo: repo}
}

func (s *activityService) RecordAction(ctx context.Context, username, method, path string) error {
	return s.repo.RecordAction(ctx, username, method+" "+path)
}

func (s *activityService) GetActions(ctx context.Context, username string) ([]string, error) {
	actions, err := s.repo.GetActions(ctx, username)
	if err != nil {
		return nil, err
	}
	if actions == nil {
		actions = []string{}
	}
	return actions, nil
}
