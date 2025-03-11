package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/HellUpa/taskmanager/internal/config"
	"github.com/HellUpa/taskmanager/internal/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
)

type PostgresDB struct {
	DB  *sql.DB
	log *slog.Logger
}

func NewPostgresDB(log *slog.Logger, cfg config.DatabaseConfig) (*PostgresDB, error) {
	// Connection string.  Consider using flags.
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Ping the database to check the connection.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Info("Connected to Database")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+cfg.MigrationsPath, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}
	m.Up()
	log.Info("Database migration completed")
	return &PostgresDB{DB: db, log: log}, nil
}

// CreateUserTx creates a new user within a transaction.
func (pdb *PostgresDB) CreateUserTx(ctx context.Context, tx *sql.Tx, user *models.User) error {
	_, err := tx.ExecContext(ctx,
		"INSERT INTO users (id, kratos_id) VALUES ($1, $2)",
		user.ID, user.KratosID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUserByKratosIDTx retrieves a user by their Kratos ID within a transaction.
func (pdb *PostgresDB) GetUserByKratosIDTx(ctx context.Context, tx *sql.Tx, kratosID string) (*models.User, error) {
	user := &models.User{}
	err := tx.QueryRowContext(ctx,
		"SELECT id, kratos_id FROM users WHERE kratos_id = $1", kratosID).
		Scan(&user.ID, &user.KratosID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by Kratos ID: %w", err)
	}
	return user, nil
}

// GetUserByIDTx retrieves a user by their ID within a transaction.
func (pdb *PostgresDB) GetUserByIDTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := tx.QueryRowContext(ctx,
		"SELECT id, kratos_id FROM users WHERE id = $1", id).
		Scan(&user.ID, &user.KratosID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

// CreateTaskTx creates a new task within a transaction.
func (pdb *PostgresDB) CreateTaskTx(ctx context.Context, tx *sql.Tx, task *models.Task) (int32, error) {
	var id int32
	err := tx.QueryRowContext(ctx,
		"INSERT INTO tasks (title, description, due_date, user_id) VALUES ($1, $2, $3, $4) RETURNING id",
		task.Title, task.Description, task.DueDate, task.UserID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %w", err)
	}
	return id, nil
}

// GetTaskTx retrieves a task by its ID within a transaction.
func (pdb *PostgresDB) GetTaskTx(ctx context.Context, tx *sql.Tx, id int32, userID uuid.UUID) (*models.Task, error) {
	task := &models.Task{}
	err := tx.QueryRowContext(ctx,
		"SELECT id, user_id, title, description, due_date, completed, created_at, updated_at FROM tasks WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.DueDate, &task.Completed, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Task not found
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// UpdateTaskTx updates an existing task within a transaction, and checks user ownership.
func (pdb *PostgresDB) UpdateTaskTx(ctx context.Context, tx *sql.Tx, task *models.Task) error {
	result, err := tx.ExecContext(ctx,
		"UPDATE tasks SET title = $1, description = $2, due_date = $3, completed = $4, updated_at = NOW() WHERE id = $5 AND user_id = $6",
		task.Title, task.Description, task.DueDate, task.Completed, task.ID, task.UserID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteTaskTx deletes a task by its ID within a transaction, and checks user ownership.
func (pdb *PostgresDB) DeleteTaskTx(ctx context.Context, tx *sql.Tx, id int32, userID uuid.UUID) error {
	result, err := tx.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ListTasksTx retrieves all tasks for a specific user within a transaction.
func (pdb *PostgresDB) ListTasksTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID) ([]*models.Task, error) {
	rows, err := tx.QueryContext(ctx, "SELECT id, user_id, title, description, due_date, completed, created_at, updated_at FROM tasks WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.DueDate, &task.Completed, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return tasks, nil
}

// Close closes the database connection.
func (pdb *PostgresDB) Close() error {
	return pdb.DB.Close()
}
