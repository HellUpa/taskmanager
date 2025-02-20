package app

import (
	"context"
	"time"

	"github.com/HellUpa/gRPC-CRUD/internal/db"
	"github.com/HellUpa/gRPC-CRUD/internal/models"
	pb "github.com/HellUpa/gRPC-CRUD/pb/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TaskManagerService implements the gRPC TaskManager service.
type TaskManagerService struct {
	pb.UnimplementedTaskManagerServer // Embed for forward compatibility
	db                                *db.PostgresDB
}

// NewTaskManagerService creates a new TaskManagerService.
func NewTaskManagerService(db *db.PostgresDB) *TaskManagerService {
	return &TaskManagerService{db: db}
}

// CreateTask handles the CreateTask RPC.
func (s *TaskManagerService) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.Task, error) {
	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil && req.DueDate != "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid due_date format")
	}

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     dueDate,
	}

	id, err := s.db.CreateTask(ctx, task)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create task: %v", err)
	}
	task.ID = int(id) // Convert int32 to int
	created_at := time.Now()
	updated_at := time.Now()
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

// GetTask handles the GetTask RPC.
func (s *TaskManagerService) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	task, err := s.db.GetTask(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get task: %v", err)
	}
	if task == nil {
		return nil, status.Errorf(codes.NotFound, "task not found")
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

// UpdateTask handles the UpdateTask RPC.
func (s *TaskManagerService) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.Task, error) {
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

	if err := s.db.UpdateTask(ctx, task); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update task: %v", err)
	}
	updated_at := time.Now()
	return &pb.Task{
		Id:          req.Id,
		Title:       task.Title,
		Description: task.Description,
		DueDate:     task.DueDate.Format(time.RFC3339),
		Completed:   task.Completed,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339), // Assuming you have created_at in model
		UpdatedAt:   updated_at.Format(time.RFC3339),
	}, nil
}

// DeleteTask handles the DeleteTask RPC.
func (s *TaskManagerService) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	if err := s.db.DeleteTask(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete task: %v", err)
	}

	return &pb.DeleteTaskResponse{Success: true}, nil
}

// ListTasks handles the ListTasks RPC.
func (s *TaskManagerService) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	tasks, err := s.db.ListTasks(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tasks: %v", err)
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
