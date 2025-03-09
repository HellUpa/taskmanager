package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/HellUpa/taskmanager/internal/db"
	logu "github.com/HellUpa/taskmanager/internal/logger/logger-utils"
	"github.com/HellUpa/taskmanager/internal/models"
)

type TaskManagerService struct {
	db  *db.PostgresDB
	log *slog.Logger
}

func NewTaskManagerService(log *slog.Logger, db *db.PostgresDB) *TaskManagerService {
	return &TaskManagerService{
		db:  db,
		log: log,
	}
}

// CreateTask creates a new task.
func (s *TaskManagerService) CreateTask(ctx context.Context, task *models.Task) (int32, error) {
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.log.Error("Rollback failed", logu.Err(rollbackErr))
			}
		}
	}()

	id, err := s.db.CreateTaskTx(ctx, tx, task)
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return id, nil
}

// GetTask retrieves a task by its ID.
func (s *TaskManagerService) GetTask(ctx context.Context, id int32) (*models.Task, error) {
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.log.Error("Rollback failed", logu.Err(rollbackErr))
			}
		}
	}()

	task, err := s.db.GetTaskTx(ctx, tx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return task, nil
}

// UpdateTask updates a task.
func (s *TaskManagerService) UpdateTask(ctx context.Context, task *models.Task) error {
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.log.Error("Rollback failed", logu.Err(rollbackErr))
			}
		}
	}()

	if err := s.db.UpdateTaskTx(ctx, tx, task); err != nil {
		return fmt.Errorf("failed to update task: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteTask deletes a task by its ID.
func (s *TaskManagerService) DeleteTask(ctx context.Context, id int32) error {
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.log.Error("Rollback failed", logu.Err(rollbackErr))
			}
		}
	}()

	if err := s.db.DeleteTaskTx(ctx, tx, id); err != nil {
		return fmt.Errorf("failed to delete task: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *TaskManagerService) ListTasks(ctx context.Context) ([]*models.Task, error) {
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.log.Error("Rollback failed", logu.Err(rollbackErr))
			}
		}
	}()

	tasks, err := s.db.ListTasksTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return tasks, nil
}
