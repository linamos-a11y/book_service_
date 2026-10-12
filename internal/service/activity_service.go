package service

import "context"

// ActivityService holds the business rules for the per-user activity log
// and has no knowledge of HTTP.
type ActivityService struct {
	repo ActivityRepository
}

// NewActivityService builds an ActivityService backed by the given
// ActivityRepository.
func NewActivityService(repo ActivityRepository) *ActivityService {
	return &ActivityService{repo: repo}
}

func (s *ActivityService) RecordAction(ctx context.Context, username, method, path string) error {
	return s.repo.RecordAction(ctx, username, method+" "+path)
}

func (s *ActivityService) GetActions(ctx context.Context, username string) ([]string, error) {
	actions, err := s.repo.GetActions(ctx, username)
	if err != nil {
		return nil, err
	}
	if actions == nil {
		actions = []string{}
	}
	return actions, nil
}
