package app

import (
	"context"
	"log"
	"time"

	"github.com/HellUpa/taskmanager/internal/db"
	"github.com/HellUpa/taskmanager/internal/models"
	pb "github.com/HellUpa/taskmanager/pb/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskManagerService struct {
	pb.UnimplementedTaskManagerServer
	db *db.PostgresDB
}

func NewTaskManagerService(db *db.PostgresDB) *TaskManagerService {
	return &TaskManagerService{db: db}
}

func (s *TaskManagerService) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.Task, error) {
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("rollback failed: %v", rollbackErr)
			}
		}
	}()
	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil && req.DueDate != "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid due_date format")
	}
	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     dueDate,
	}

	id, err := s.db.CreateTaskTx(ctx, tx, task)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create task: %v", err)
	}
	task.ID = int(id)
	created_at := time.Now()
	updated_at := time.Now()

	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	return &pb.Task{
		Id:          int32(task.ID),
		Title:       task.Title,
		Description: task.Description,
		DueDate:     task.DueDate.Format(time.RFC3339),
		Completed:   task.Completed,
		CreatedAt:   created_at.Format(time.RFC3339),
		UpdatedAt:   updated_at.Format(time.RFC3339),
	}, nil
}

func (s *TaskManagerService) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("rollback failed: %v", rollbackErr)
			}
		}
	}()

	task, err := s.db.GetTaskTx(ctx, tx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get task: %v", err)
	}
	if task == nil {
		return nil, status.Errorf(codes.NotFound, "task not found")
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	return &pb.Task{
		Id:          int32(task.ID),
		Title:       task.Title,
		Description: task.Description,
		DueDate:     task.DueDate.Format(time.RFC3339),
		Completed:   task.Completed,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *TaskManagerService) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.Task, error) {
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("rollback failed: %v", rollbackErr)
			}
		}
	}()
	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil && req.DueDate != "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid due_date format")
	}
	task := &models.Task{
		ID:          int(req.Id),
		Title:       req.Title,
		Description: req.Description,
		DueDate:     dueDate,
		Completed:   req.Completed,
	}

	if err := s.db.UpdateTaskTx(ctx, tx, task); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update task: %v", err)
	}
	updated_at := time.Now()
	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	return &pb.Task{
		Id:          req.Id,
		Title:       task.Title,
		Description: task.Description,
		DueDate:     task.DueDate.Format(time.RFC3339),
		Completed:   task.Completed,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   updated_at.Format(time.RFC3339),
	}, nil
}

func (s *TaskManagerService) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("rollback failed: %v", rollbackErr)
			}
		}
	}()

	if err := s.db.DeleteTaskTx(ctx, tx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete task: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	return &pb.DeleteTaskResponse{Success: true}, nil
}

func (s *TaskManagerService) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("rollback failed: %v", rollbackErr)
			}
		}
	}()

	tasks, err := s.db.ListTasksTx(ctx, tx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tasks: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	pbTasks := make([]*pb.Task, 0, len(tasks))
	for _, task := range tasks {
		pbTasks = append(pbTasks, &pb.Task{
			Id:          int32(task.ID),
			Title:       task.Title,
			Description: task.Description,
			DueDate:     task.DueDate.Format(time.RFC3339),
			Completed:   task.Completed,
			CreatedAt:   task.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &pb.ListTasksResponse{Tasks: pbTasks}, nil
}
