package identity

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &User{}
	err = r.db.QueryRow(ctx,
		`INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3)
		 RETURNING id, name, email, password_hash, avatar_url, created_at, updated_at`,
		input.Name, input.Email, string(hash),
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, password_hash, avatar_url, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, password_hash, avatar_url, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *Repository) CreateAgent(ctx context.Context, input CreateAgentInput) (*AIAgent, string, error) {
	apiKey := uuid.New().String()
	keyHash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash api key: %w", err)
	}

	capJSON, err := json.Marshal(input.Capabilities)
	if err != nil {
		return nil, "", fmt.Errorf("marshal capabilities: %w", err)
	}
	metaJSON, err := json.Marshal(input.Metadata)
	if err != nil {
		return nil, "", fmt.Errorf("marshal metadata: %w", err)
	}

	agent := &AIAgent{}
	err = r.db.QueryRow(ctx,
		`INSERT INTO ai_agents (name, model_type, api_key_hash, capabilities, permission_level, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, name, model_type, api_key_hash, capabilities, permission_level, metadata, is_active, created_at, updated_at`,
		input.Name, input.ModelType, string(keyHash), capJSON, input.PermissionLevel, metaJSON,
	).Scan(&agent.ID, &agent.Name, &agent.ModelType, &agent.APIKeyHash, &capJSON, &agent.PermissionLevel, &metaJSON, &agent.IsActive, &agent.CreatedAt, &agent.UpdatedAt)
	if err != nil {
		return nil, "", fmt.Errorf("create agent: %w", err)
	}
	if err := json.Unmarshal(capJSON, &agent.Capabilities); err != nil {
		return nil, "", fmt.Errorf("unmarshal capabilities: %w", err)
	}
	if err := json.Unmarshal(metaJSON, &agent.Metadata); err != nil {
		return nil, "", fmt.Errorf("unmarshal metadata: %w", err)
	}
	return agent, apiKey, nil
}

func (r *Repository) GetAgentByID(ctx context.Context, id uuid.UUID) (*AIAgent, error) {
	agent := &AIAgent{}
	var capJSON, metaJSON []byte
	err := r.db.QueryRow(ctx,
		`SELECT id, name, model_type, api_key_hash, capabilities, permission_level, metadata, is_active, created_at, updated_at
		 FROM ai_agents WHERE id = $1`, id,
	).Scan(&agent.ID, &agent.Name, &agent.ModelType, &agent.APIKeyHash, &capJSON, &agent.PermissionLevel, &metaJSON, &agent.IsActive, &agent.CreatedAt, &agent.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get agent by id: %w", err)
	}
	if err := json.Unmarshal(capJSON, &agent.Capabilities); err != nil {
		return nil, fmt.Errorf("unmarshal capabilities: %w", err)
	}
	if err := json.Unmarshal(metaJSON, &agent.Metadata); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}
	return agent, nil
}

func (r *Repository) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, role_type, description, permissions FROM roles ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var permJSON []byte
		if err := rows.Scan(&role.ID, &role.Name, &role.RoleType, &role.Description, &permJSON); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		if err := json.Unmarshal(permJSON, &role.Permissions); err != nil {
			return nil, fmt.Errorf("unmarshal role permissions: %w", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list roles iteration: %w", err)
	}
	return roles, nil
}
