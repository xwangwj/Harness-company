package layer

import (
	"time"

	"github.com/google/uuid"
)

type LayerType string

const (
	LayerStrategic   LayerType = "strategic"
	LayerTactical    LayerType = "tactical"
	LayerOperational LayerType = "operational"
)

type LayerConfig struct {
	ID        uuid.UUID      `json:"id"`
	MVRUID    uuid.UUID      `json:"mvru_id"`
	Layer     LayerType      `json:"layer"`
	Config    map[string]any `json:"config"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type LayerRoutingRule struct {
	ID          uuid.UUID      `json:"id"`
	SourceLayer LayerType      `json:"source_layer"`
	TargetLayer LayerType      `json:"target_layer"`
	Condition   map[string]any `json:"condition"`
	Priority    int            `json:"priority"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
}
