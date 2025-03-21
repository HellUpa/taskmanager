package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/HellUpa/taskmanager/internal/db"
	logu "github.com/HellUpa/taskmanager/internal/logger/logger-utils"
	"github.com/HellUpa/taskmanager/internal/models"
	"github.com/google/uuid"
)

type TaskManagerService struct {
	db  *db.PostgresDB
	Log *slog.Logger
}

func NewTaskManagerService(log *slog.Logger, db *db.PostgresDB) *TaskManagerService {
	log.Debug("Initializing TaskManagerService")
	return &TaskManagerService{
		db:  db,
		Log: log,
	}
}

// CreateTask creates a new task.
func (s *TaskManagerService) CreateTask(ctx context.Context, task *models.Task, userID uuid.UUID) (int32, error) {
	s.Log.Debug("Starting CreateTask", slog.String("userID", userID.String()))
	task.UserID = userID

	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.Log.Error("Rollback failed", logu.Err(rollbackErr))
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
	s.Log.Debug("Task created successfully", slog.Int("taskID", int(id)))
	return id, nil
}

// GetTask retrieves a task by its ID.
func (s *TaskManagerService) GetTask(ctx context.Context, id int32, userID uuid.UUID) (*models.Task, error) {
	s.Log.Debug("Starting GetTask", slog.Int("taskID", int(id)), slog.String("userID", userID.String()))
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.Log.Error("Rollback failed", logu.Err(rollbackErr))
			}
		}
	}()

	task, err := s.db.GetTaskTx(ctx, tx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	s.Log.Debug("Task retrieved successfully", slog.Any("task", task))
	return task, nil
}

// UpdateTask updates a task.
func (s *TaskManagerService) UpdateTask(ctx context.Context, task *models.Task) error {
	s.Log.Debug("Starting UpdateTask", slog.Int("taskID", int(task.ID)))
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.Log.Error("Rollback failed", logu.Err(rollbackErr))
			}
		}
	}()

	if err := s.db.UpdateTaskTx(ctx, tx, task); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("task with id %d not found: %w", task.ID, err)
		}
		return fmt.Errorf("failed to update task: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	s.Log.Debug("Task updated successfully", slog.Int("taskID", int(task.ID)))
	return nil
}

// DeleteTask deletes a task by its ID.
func (s *TaskManagerService) DeleteTask(ctx context.Context, id int32, userID uuid.UUID) error {
	s.Log.Debug("Starting DeleteTask", slog.Int("taskID", int(id)), slog.String("userID", userID.String()))
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.Log.Error("Rollback failed", logu.Err(rollbackErr))
			}
		}
	}()

	if err := s.db.DeleteTaskTx(ctx, tx, id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("task with id %d not found: %w", id, err)
		}
		return fmt.Errorf("failed to delete task: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	s.Log.Debug("Task deleted successfully", slog.Int("taskID", int(id)))
	return nil
}

func (s *TaskManagerService) ListTasks(ctx context.Context, userID uuid.UUID) ([]*models.Task, error) {
	s.Log.Debug("Starting ListTasks", slog.String("userID", userID.String()))
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.Log.Error("Rollback failed", logu.Err(rollbackErr))
			}
		}
	}()

	tasks, err := s.db.ListTasksTx(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	s.Log.Debug("Tasks listed successfully")
	return tasks, nil
}

// CreateUser creates a new task.
func (s *TaskManagerService) CreateUser(ctx context.Context, user *models.User) error {
	s.Log.Debug("Starting CreateUser", slog.Any("user", user))
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.Log.Error("Rollback failed", logu.Err(rollbackErr))
			}
		}
	}()

	if err := s.db.CreateUserTx(ctx, tx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	s.Log.Debug("User created successfully", slog.String("userID", user.ID.String()))
	return nil
}

// GetUserByKratosID retrieves a user by their Kratos ID.
func (s *TaskManagerService) GetUserByKratosID(ctx context.Context, kratosID string) (*models.User, error) {
	s.Log.Debug("Starting GetUserByKratosID", slog.String("kratosID", kratosID))
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.Log.Error("Rollback failed", logu.Err(rollbackErr))
			}
		}
	}()

	user, err := s.db.GetUserByKratosIDTx(ctx, tx, kratosID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by Kratos ID: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	s.Log.Debug("User retrieved successfully", slog.Any("user", user))
	return user, nil
}

// GetUserByID retrieves a user by their  ID.
func (s *TaskManagerService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	s.Log.Debug("Starting GetUserByID", slog.String("userID", id.String()))
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.Log.Error("Rollback failed", logu.Err(rollbackErr))
			}
		}
	}()

	user, err := s.db.GetUserByIDTx(ctx, tx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by  ID: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	s.Log.Debug("User retrieved successfully", slog.Any("user", user))
	return user, nil
}
