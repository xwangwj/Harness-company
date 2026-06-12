package verification

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateReport(ctx context.Context, input CreateReportInput) (*VerificationReport, error) {
	suggestions, _ := json.Marshal(input.Suggestions)
	report := &VerificationReport{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO verification_reports (workflow_id, task_id, result_score, path_score, environment_score, conclusion, suggestions)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, workflow_id, task_id, result_score, path_score, environment_score, overall_score, conclusion, suggestions, created_at`,
		input.WorkflowID, input.TaskID, input.ResultScore, input.PathScore, input.EnvironmentScore, input.Conclusion, suggestions,
	).Scan(&report.ID, &report.WorkflowID, &report.TaskID, &report.ResultScore, &report.PathScore, &report.EnvironmentScore, &report.OverallScore, &report.Conclusion, &suggestions, &report.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create report: %w", err)
	}
	json.Unmarshal(suggestions, &report.Suggestions)
	return report, nil
}

func (r *Repository) GetReport(ctx context.Context, id uuid.UUID) (*VerificationReport, error) {
	report := &VerificationReport{}
	var suggestions []byte
	err := r.db.QueryRow(ctx,
		`SELECT id, workflow_id, task_id, result_score, path_score, environment_score, overall_score, conclusion, suggestions, created_at
		 FROM verification_reports WHERE id = $1`, id,
	).Scan(&report.ID, &report.WorkflowID, &report.TaskID, &report.ResultScore, &report.PathScore, &report.EnvironmentScore, &report.OverallScore, &report.Conclusion, &suggestions, &report.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get report: %w", err)
	}
	json.Unmarshal(suggestions, &report.Suggestions)
	return report, nil
}

func (r *Repository) ListReports(ctx context.Context, workflowID *uuid.UUID, limit int) ([]VerificationReport, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var query string
	var args []any
	if workflowID != nil {
		query = `SELECT id, workflow_id, task_id, result_score, path_score, environment_score, overall_score, conclusion, suggestions, created_at
				 FROM verification_reports WHERE workflow_id = $1 ORDER BY created_at DESC LIMIT $2`
		args = []any{*workflowID, limit}
	} else {
		query = `SELECT id, workflow_id, task_id, result_score, path_score, environment_score, overall_score, conclusion, suggestions, created_at
				 FROM verification_reports ORDER BY created_at DESC LIMIT $1`
		args = []any{limit}
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()

	var reports []VerificationReport
	for rows.Next() {
		var report VerificationReport
		var suggestions []byte
		if err := rows.Scan(&report.ID, &report.WorkflowID, &report.TaskID, &report.ResultScore, &report.PathScore, &report.EnvironmentScore, &report.OverallScore, &report.Conclusion, &suggestions, &report.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		json.Unmarshal(suggestions, &report.Suggestions)
		reports = append(reports, report)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list reports iteration: %w", err)
	}
	return reports, nil
}

func (r *Repository) AssignReview(ctx context.Context, input AssignReviewInput) (*ReviewAssignment, error) {
	review := &ReviewAssignment{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO review_assignments (report_id, level, reviewer_id, reviewer_type, status)
		 VALUES ($1, $2, $3, $4, 'pending')
		 RETURNING id, report_id, level, reviewer_id, reviewer_type, status, result, completed_at, created_at`,
		input.ReportID, input.Level, input.ReviewerID, input.ReviewerType,
	).Scan(&review.ID, &review.ReportID, &review.Level, &review.ReviewerID, &review.ReviewerType, &review.Status, &review.Result, &review.CompletedAt, &review.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("assign review: %w", err)
	}
	return review, nil
}

func (r *Repository) CompleteReview(ctx context.Context, reviewID uuid.UUID, result map[string]any) error {
	resultJSON, _ := json.Marshal(result)
	_, err := r.db.Exec(ctx,
		`UPDATE review_assignments SET status = 'completed', result = $1, completed_at = NOW() WHERE id = $2`,
		resultJSON, reviewID,
	)
	if err != nil {
		return fmt.Errorf("complete review: %w", err)
	}
	return nil
}

func (r *Repository) GetReviewsByReport(ctx context.Context, reportID uuid.UUID) ([]ReviewAssignment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, report_id, level, reviewer_id, reviewer_type, status, result, completed_at, created_at
		 FROM review_assignments WHERE report_id = $1 ORDER BY created_at`, reportID)
	if err != nil {
		return nil, fmt.Errorf("get reviews by report: %w", err)
	}
	defer rows.Close()

	var reviews []ReviewAssignment
	for rows.Next() {
		var review ReviewAssignment
		if err := rows.Scan(&review.ID, &review.ReportID, &review.Level, &review.ReviewerID, &review.ReviewerType, &review.Status, &review.Result, &review.CompletedAt, &review.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get reviews iteration: %w", err)
	}
	return reviews, nil
}

func (r *Repository) UpdateOverallScore(ctx context.Context, reportID uuid.UUID, score float64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE verification_reports SET overall_score = $1 WHERE id = $2`,
		score, reportID,
	)
	if err != nil {
		return fmt.Errorf("update overall score: %w", err)
	}
	return nil
}
