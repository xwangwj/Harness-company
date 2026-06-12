package workflow

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

func (r *Repository) CreateTemplate(ctx context.Context, input CreateWorkflowInput) (*WorkflowTemplate, error) {
	stagesJSON, _ := json.Marshal(input.Stages)
	rulesJSON, _ := json.Marshal(input.RoutingRules)

	t := &WorkflowTemplate{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO workflow_templates (name, description, stages, assignee_type, required_weight, routing_rules)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, name, description, stages, assignee_type, required_weight, routing_rules, is_active, created_at, updated_at`,
		input.Name, input.Description, stagesJSON, input.AssigneeType, input.RequiredWeight, rulesJSON,
	).Scan(&t.ID, &t.Name, &t.Description, &stagesJSON, &t.AssigneeType, &t.RequiredWeight, &rulesJSON, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	json.Unmarshal(stagesJSON, &t.Stages)
	json.Unmarshal(rulesJSON, &t.RoutingRules)
	return t, nil
}

func (r *Repository) GetTemplate(ctx context.Context, id uuid.UUID) (*WorkflowTemplate, error) {
	t := &WorkflowTemplate{}
	var stagesJSON, rulesJSON []byte
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, stages, assignee_type, required_weight, routing_rules, is_active, created_at, updated_at
		 FROM workflow_templates WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.Description, &stagesJSON, &t.AssigneeType, &t.RequiredWeight, &rulesJSON, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}
	json.Unmarshal(stagesJSON, &t.Stages)
	json.Unmarshal(rulesJSON, &t.RoutingRules)
	return t, nil
}

func (r *Repository) ListTemplates(ctx context.Context) ([]WorkflowTemplate, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, stages, assignee_type, required_weight, routing_rules, is_active, created_at, updated_at
		 FROM workflow_templates ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	var templates []WorkflowTemplate
	for rows.Next() {
		var t WorkflowTemplate
		var stagesJSON, rulesJSON []byte
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &stagesJSON, &t.AssigneeType, &t.RequiredWeight, &rulesJSON, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		json.Unmarshal(stagesJSON, &t.Stages)
		json.Unmarshal(rulesJSON, &t.RoutingRules)
		templates = append(templates, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list templates iteration: %w", err)
	}
	return templates, nil
}

func (r *Repository) CreateInstance(ctx context.Context, input StartWorkflowInput) (*WorkflowInstance, error) {
	contextJSON, _ := json.Marshal(input.Context)
	inst := &WorkflowInstance{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO workflow_instances (template_id, context) VALUES ($1, $2)
		 RETURNING id, template_id, status, current_stage, context, trace_id, created_at, updated_at`,
		input.TemplateID, contextJSON,
	).Scan(&inst.ID, &inst.TemplateID, &inst.Status, &inst.CurrentStage, &contextJSON, &inst.TraceID, &inst.CreatedAt, &inst.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create instance: %w", err)
	}
	json.Unmarshal(contextJSON, &inst.Context)
	return inst, nil
}

func (r *Repository) GetInstance(ctx context.Context, id uuid.UUID) (*WorkflowInstance, error) {
	inst := &WorkflowInstance{}
	var contextJSON []byte
	err := r.db.QueryRow(ctx,
		`SELECT id, template_id, status, current_stage, context, trace_id, created_at, updated_at
		 FROM workflow_instances WHERE id = $1`, id,
	).Scan(&inst.ID, &inst.TemplateID, &inst.Status, &inst.CurrentStage, &contextJSON, &inst.TraceID, &inst.CreatedAt, &inst.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get instance: %w", err)
	}
	json.Unmarshal(contextJSON, &inst.Context)
	return inst, nil
}

func (r *Repository) UpdateInstanceStatus(ctx context.Context, id uuid.UUID, status WorkflowStatus) error {
	_, err := r.db.Exec(ctx, `UPDATE workflow_instances SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	if err != nil {
		return fmt.Errorf("update instance status: %w", err)
	}
	return nil
}

func (r *Repository) UpdateInstanceStage(ctx context.Context, id uuid.UUID, stage int) error {
	_, err := r.db.Exec(ctx, `UPDATE workflow_instances SET current_stage = $1, updated_at = NOW() WHERE id = $2`, stage, id)
	if err != nil {
		return fmt.Errorf("update instance stage: %w", err)
	}
	return nil
}

func (r *Repository) CreateTask(ctx context.Context, task *Task) (*Task, error) {
	inputJSON, _ := json.Marshal(task.Input)
	err := r.db.QueryRow(ctx,
		`INSERT INTO tasks (workflow_id, stage, stage_type, assignee_id, assignee_type, input, weight_snapshot, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, workflow_id, stage, stage_type, assignee_id, assignee_type, input, output, weight_snapshot, status, created_at, updated_at`,
		task.WorkflowID, task.Stage, task.StageType, task.AssigneeID, task.AssigneeType, inputJSON, task.WeightSnapshot, task.Status,
	).Scan(&task.ID, &task.WorkflowID, &task.Stage, &task.StageType, &task.AssigneeID, &task.AssigneeType, &inputJSON, &task.Output, &task.WeightSnapshot, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	json.Unmarshal(inputJSON, &task.Input)
	return task, nil
}

func (r *Repository) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status TaskStatus, output map[string]any) error {
	outputJSON, _ := json.Marshal(output)
	_, err := r.db.Exec(ctx,
		`UPDATE tasks SET status = $1, output = $2, updated_at = NOW() WHERE id = $3`, status, outputJSON, id)
	if err != nil {
		return fmt.Errorf("update task status: %w", err)
	}
	return nil
}

func (r *Repository) GetTasksByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]Task, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, workflow_id, stage, stage_type, assignee_id, assignee_type, input, output, weight_snapshot, status, created_at, updated_at
		 FROM tasks WHERE workflow_id = $1 ORDER BY stage, created_at`, workflowID)
	if err != nil {
		return nil, fmt.Errorf("get tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var inputJSON, outputJSON []byte
		if err := rows.Scan(&task.ID, &task.WorkflowID, &task.Stage, &task.StageType, &task.AssigneeID, &task.AssigneeType, &inputJSON, &outputJSON, &task.WeightSnapshot, &task.Status, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		json.Unmarshal(inputJSON, &task.Input)
		if outputJSON != nil {
			json.Unmarshal(outputJSON, &task.Output)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get tasks iteration: %w", err)
	}
	return tasks, nil
}

func (r *Repository) RecordDecision(ctx context.Context, d *Decision) (*Decision, error) {
	inputJSON, _ := json.Marshal(d.Input)
	outputJSON, _ := json.Marshal(d.Output)
	err := r.db.QueryRow(ctx,
		`INSERT INTO decisions (task_id, decision_maker_id, maker_type, weight, input, output, reasoning, outcome)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, task_id, decision_maker_id, maker_type, weight, input, output, reasoning, outcome, created_at`,
		d.TaskID, d.DecisionMakerID, d.MakerType, d.Weight, inputJSON, outputJSON, d.Reasoning, d.Outcome,
	).Scan(&d.ID, &d.TaskID, &d.DecisionMakerID, &d.MakerType, &d.Weight, &inputJSON, &outputJSON, &d.Reasoning, &d.Outcome, &d.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("record decision: %w", err)
	}
	json.Unmarshal(inputJSON, &d.Input)
	json.Unmarshal(outputJSON, &d.Output)
	return d, nil
}

func (r *Repository) GetWorkflowContext(ctx context.Context, workflowID uuid.UUID) (*WorkflowContext, error) {
	wc := &WorkflowContext{}
	var memJSON, expJSON []byte
	err := r.db.QueryRow(ctx,
		`SELECT id, workflow_id, working_memory, injected_experience, principle_notes, created_at, updated_at
		 FROM workflow_contexts WHERE workflow_id = $1`, workflowID,
	).Scan(&wc.ID, &wc.WorkflowID, &memJSON, &expJSON, &wc.PrincipleNotes, &wc.CreatedAt, &wc.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get workflow context: %w", err)
	}
	json.Unmarshal(memJSON, &wc.WorkingMemory)
	json.Unmarshal(expJSON, &wc.InjectedExperience)
	return wc, nil
}

func (r *Repository) UpsertWorkflowContext(ctx context.Context, wc *WorkflowContext) error {
	memJSON, _ := json.Marshal(wc.WorkingMemory)
	expJSON, _ := json.Marshal(wc.InjectedExperience)
	_, err := r.db.Exec(ctx,
		`INSERT INTO workflow_contexts (workflow_id, working_memory, injected_experience, principle_notes)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (workflow_id) DO UPDATE SET working_memory = $2, injected_experience = $3, principle_notes = $4, updated_at = NOW()`,
		wc.WorkflowID, memJSON, expJSON, wc.PrincipleNotes)
	if err != nil {
		return fmt.Errorf("upsert workflow context: %w", err)
	}
	return nil
}
