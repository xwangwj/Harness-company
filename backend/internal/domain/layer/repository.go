package layer

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

func (r *Repository) SetLayerConfig(ctx context.Context, mvruID uuid.UUID, layer LayerType, config map[string]any) error {
	configJSON, _ := json.Marshal(config)
	_, err := r.db.Exec(ctx,
		`INSERT INTO layer_configs (mvru_id, layer, config) VALUES ($1, $2, $3)
		 ON CONFLICT (mvru_id, layer) DO UPDATE SET config = $3, updated_at = NOW()`,
		mvruID, layer, configJSON)
	if err != nil {
		return fmt.Errorf("set layer config: %w", err)
	}
	return nil
}

func (r *Repository) GetLayerConfig(ctx context.Context, mvruID uuid.UUID, layer LayerType) (*LayerConfig, error) {
	lc := &LayerConfig{}
	var configJSON []byte
	err := r.db.QueryRow(ctx,
		`SELECT id, mvru_id, layer, config, created_at, updated_at
		 FROM layer_configs WHERE mvru_id = $1 AND layer = $2`,
		mvruID, layer,
	).Scan(&lc.ID, &lc.MVRUID, &lc.Layer, &configJSON, &lc.CreatedAt, &lc.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get layer config: %w", err)
	}
	json.Unmarshal(configJSON, &lc.Config)
	return lc, nil
}

func (r *Repository) ListLayerConfigs(ctx context.Context, mvruID uuid.UUID) ([]LayerConfig, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, mvru_id, layer, config, created_at, updated_at
		 FROM layer_configs WHERE mvru_id = $1 ORDER BY layer`, mvruID)
	if err != nil {
		return nil, fmt.Errorf("list layer configs: %w", err)
	}
	defer rows.Close()

	var configs []LayerConfig
	for rows.Next() {
		var lc LayerConfig
		var configJSON []byte
		if err := rows.Scan(&lc.ID, &lc.MVRUID, &lc.Layer, &configJSON, &lc.CreatedAt, &lc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan layer config: %w", err)
		}
		json.Unmarshal(configJSON, &lc.Config)
		configs = append(configs, lc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list layer configs iteration: %w", err)
	}
	return configs, nil
}

func (r *Repository) ListRoutingRules(ctx context.Context) ([]LayerRoutingRule, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, source_layer, target_layer, condition, priority, is_active, created_at
		 FROM layer_routing_rules ORDER BY priority DESC`)
	if err != nil {
		return nil, fmt.Errorf("list routing rules: %w", err)
	}
	defer rows.Close()

	var rules []LayerRoutingRule
	for rows.Next() {
		var rule LayerRoutingRule
		var condJSON []byte
		if err := rows.Scan(&rule.ID, &rule.SourceLayer, &rule.TargetLayer, &condJSON, &rule.Priority, &rule.IsActive, &rule.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan routing rule: %w", err)
		}
		json.Unmarshal(condJSON, &rule.Condition)
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list routing rules iteration: %w", err)
	}
	return rules, nil
}
