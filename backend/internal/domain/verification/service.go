package verification

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation error")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateReport(ctx context.Context, input CreateReportInput) (*VerificationReport, error) {
	report, err := s.repo.CreateReport(ctx, input)
	if err != nil {
		return nil, err
	}
	score := calculateOverallScore(input.ResultScore, input.PathScore, input.EnvironmentScore)
	if score != nil {
		if err := s.repo.UpdateOverallScore(ctx, report.ID, *score); err != nil {
			return nil, fmt.Errorf("update overall score: %w", err)
		}
		report.OverallScore = score
	}
	return report, nil
}

func (s *Service) GetReport(ctx context.Context, id uuid.UUID) (*VerificationReport, error) {
	report, err := s.repo.GetReport(ctx, id)
	if err != nil {
		return nil, err
	}
	reviews, err := s.repo.GetReviewsByReport(ctx, id)
	if err != nil {
		return nil, err
	}
	report.Reviews = reviews
	return report, nil
}

func (s *Service) ListReports(ctx context.Context, workflowID *uuid.UUID, limit int) ([]VerificationReport, error) {
	return s.repo.ListReports(ctx, workflowID, limit)
}

func (s *Service) AssignReview(ctx context.Context, input AssignReviewInput) (*ReviewAssignment, error) {
	switch input.Level {
	case "L1", "L2", "L3":
	default:
		return nil, fmt.Errorf("%w: invalid level %q, must be L1, L2, or L3", ErrValidation, input.Level)
	}
	switch input.ReviewerType {
	case "machine", "ai", "expert":
	default:
		return nil, fmt.Errorf("%w: invalid reviewer_type %q, must be machine, ai, or expert", ErrValidation, input.ReviewerType)
	}
	return s.repo.AssignReview(ctx, input)
}

func (s *Service) CompleteReview(ctx context.Context, reviewID uuid.UUID, result map[string]any) error {
	return s.repo.CompleteReview(ctx, reviewID, result)
}

func calculateOverallScore(result, path, env *float64) *float64 {
	if result == nil && path == nil && env == nil {
		return nil
	}
	r := 0.0
	if result != nil {
		r = *result
	}
	p := 0.0
	if path != nil {
		p = *path
	}
	e := 0.0
	if env != nil {
		e = *env
	}
	score := r*0.4 + p*0.35 + e*0.25
	return &score
}
