package identity

import (
	"time"

	"github.com/google/uuid"
)

type RoleType string

const (
	RolePlanner  RoleType = "planner"
	RoleExecutor RoleType = "executor"
	RoleReviewer RoleType = "reviewer"
)

type PermissionLevel string

const (
	PermissionL1 PermissionLevel = "L1"
	PermissionL2 PermissionLevel = "L2"
	PermissionL3 PermissionLevel = "L3"
	PermissionL4 PermissionLevel = "L4"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	Roles        []Role    `json:"roles,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AIAgent struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	ModelType       string          `json:"model_type"`
	APIKeyHash      string          `json:"-"`
	Capabilities    []string        `json:"capabilities"`
	PermissionLevel PermissionLevel `json:"permission_level"`
	Metadata        map[string]any  `json:"metadata,omitempty"`
	IsActive        bool            `json:"is_active"`
	Roles           []Role          `json:"roles,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type Role struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	RoleType    RoleType        `json:"role_type"`
	Description string          `json:"description,omitempty"`
	Permissions []string        `json:"permissions"`
}

type CreateUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAgentInput struct {
	Name            string          `json:"name"`
	ModelType       string          `json:"model_type"`
	Capabilities    []string        `json:"capabilities"`
	PermissionLevel PermissionLevel `json:"permission_level"`
	Metadata        map[string]any  `json:"metadata,omitempty"`
}
