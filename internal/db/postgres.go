package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/HellUpa/taskmanager/internal/config"
	"github.com/HellUpa/taskmanager/internal/models"
)

type PostgresDB struct {
	DB *sql.DB
}

func NewPostgresDB(cfg config.DatabaseConfig) (*PostgresDB, error) {
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
	log.Println("Connected to Database")
	return &PostgresDB{DB: db}, nil
}

// CreateTask creates a new task in the database.
func (pdb *PostgresDB) CreateTask(ctx context.Context, task *models.Task) (int32, error) {
	var id int32
	err := pdb.DB.QueryRowContext(ctx,
		"INSERT INTO tasks (title, description, due_date) VALUES ($1, $2, $3) RETURNING id",
		task.Title, task.Description, task.DueDate).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %w", err)
	}
	return id, nil
}

// GetTask retrieves a task by its ID.
func (pdb *PostgresDB) GetTask(ctx context.Context, id int32) (*models.Task, error) {
	task := &models.Task{}
	err := pdb.DB.QueryRowContext(ctx,
		"SELECT id, title, description, due_date, completed, created_at, updated_at FROM tasks WHERE id = $1", id).
		Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Completed, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Task not found
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// UpdateTask updates an existing task.
func (pdb *PostgresDB) UpdateTask(ctx context.Context, task *models.Task) error {
	_, err := pdb.DB.ExecContext(ctx,
		"UPDATE tasks SET title = $1, description = $2, due_date = $3, completed = $4, updated_at = NOW() WHERE id = $5",
		task.Title, task.Description, task.DueDate, task.Completed, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

// DeleteTask deletes a task by its ID.
func (pdb *PostgresDB) DeleteTask(ctx context.Context, id int32) error {
	_, err := pdb.DB.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// ListTasks retrieves all tasks.
func (pdb *PostgresDB) ListTasks(ctx context.Context) ([]*models.Task, error) {
	rows, err := pdb.DB.QueryContext(ctx, "SELECT id, title, description, due_date, completed, created_at, updated_at FROM tasks")
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Completed, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return tasks, nil
}

// CreateTaskTx creates a new task within a transaction.
func (pdb *PostgresDB) CreateTaskTx(ctx context.Context, tx *sql.Tx, task *models.Task) (int32, error) {
	var id int32
	err := tx.QueryRowContext(ctx,
		"INSERT INTO tasks (title, description, due_date) VALUES ($1, $2, $3) RETURNING id",
		task.Title, task.Description, task.DueDate).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %w", err)
	}
	return id, nil
}

// GetTaskTx retrieves a task by its ID within a transaction.
func (pdb *PostgresDB) GetTaskTx(ctx context.Context, tx *sql.Tx, id int32) (*models.Task, error) {
	task := &models.Task{}
	err := tx.QueryRowContext(ctx,
		"SELECT id, title, description, due_date, completed, created_at, updated_at FROM tasks WHERE id = $1", id).
		Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Completed, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Task not found
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// UpdateTaskTx updates an existing task within a transaction.
func (pdb *PostgresDB) UpdateTaskTx(ctx context.Context, tx *sql.Tx, task *models.Task) error {
	_, err := tx.ExecContext(ctx,
		"UPDATE tasks SET title = $1, description = $2, due_date = $3, completed = $4, updated_at = NOW() WHERE id = $5",
		task.Title, task.Description, task.DueDate, task.Completed, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

// DeleteTaskTx deletes a task by its ID within a transaction.
func (pdb *PostgresDB) DeleteTaskTx(ctx context.Context, tx *sql.Tx, id int32) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// ListTasksTx retrieves all tasks within a transaction.
func (pdb *PostgresDB) ListTasksTx(ctx context.Context, tx *sql.Tx) ([]*models.Task, error) {
	rows, err := tx.QueryContext(ctx, "SELECT id, title, description, due_date, completed, created_at, updated_at FROM tasks")
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Completed, &task.CreatedAt, &task.UpdatedAt); err != nil {
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
