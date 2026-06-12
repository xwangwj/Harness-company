package verification

import (
	"time"

	"github.com/google/uuid"
)

type VerificationReport struct {
	ID               uuid.UUID          `json:"id"`
	WorkflowID       *uuid.UUID         `json:"workflow_id,omitempty"`
	TaskID           *uuid.UUID         `json:"task_id,omitempty"`
	ResultScore      *float64           `json:"result_score,omitempty"`
	PathScore        *float64           `json:"path_score,omitempty"`
	EnvironmentScore *float64           `json:"environment_score,omitempty"`
	OverallScore     *float64           `json:"overall_score,omitempty"`
	Conclusion       string             `json:"conclusion"`
	Suggestions      []string           `json:"suggestions"`
	CreatedAt        time.Time          `json:"created_at"`
	Reviews          []ReviewAssignment `json:"reviews,omitempty"`
}

type ReviewAssignment struct {
	ID           uuid.UUID      `json:"id"`
	ReportID     uuid.UUID      `json:"report_id"`
	Level        string         `json:"level"`
	ReviewerID   *uuid.UUID     `json:"reviewer_id,omitempty"`
	ReviewerType string         `json:"reviewer_type"`
	Status       string         `json:"status"`
	Result       map[string]any `json:"result,omitempty"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

type CreateReportInput struct {
	WorkflowID       *uuid.UUID `json:"workflow_id,omitempty"`
	TaskID           *uuid.UUID `json:"task_id,omitempty"`
	ResultScore      *float64   `json:"result_score,omitempty"`
	PathScore        *float64   `json:"path_score,omitempty"`
	EnvironmentScore *float64   `json:"environment_score,omitempty"`
	Conclusion       string     `json:"conclusion"`
	Suggestions      []string   `json:"suggestions,omitempty"`
}

type AssignReviewInput struct {
	ReportID     uuid.UUID  `json:"report_id"`
	Level        string     `json:"level"`
	ReviewerID   *uuid.UUID `json:"reviewer_id,omitempty"`
	ReviewerType string     `json:"reviewer_type"`
}

type CompleteReviewInput struct {
	Result map[string]any `json:"result"`
}
