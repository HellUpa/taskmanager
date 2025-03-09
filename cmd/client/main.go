package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/HellUpa/taskmanager/pb/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to the gRPC server.
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create a client for the TaskManager service.
	client := pb.NewTaskManagerClient(conn)

	// Example usage:
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Create a task.
	createRes, err := client.CreateTask(ctx, &pb.CreateTaskRequest{
		Title:       "My First Task",
		Description: "This is a test task.",
		DueDate:     time.Now().Add(24 * time.Hour).Format(time.RFC3339), // Due tomorrow
	})
	if err != nil {
		log.Fatalf("could not create task: %v", err)
	}
	fmt.Printf("Created task: %+v\n", createRes)

	// Get the created task.
	getRes, err := client.GetTask(ctx, &pb.GetTaskRequest{Id: createRes.Id})
	if err != nil {
		log.Fatalf("could not get task: %v", err)
	}
	fmt.Printf("Got task: %+v\n", getRes)

	// Update the task.
	updateRes, err := client.UpdateTask(ctx, &pb.UpdateTaskRequest{
		Id:          createRes.Id,
		Title:       "Updated Task Title",
		Description: "Updated description.",
		DueDate:     time.Now().Add(48 * time.Hour).Format(time.RFC3339), // Due in 2 days
		Completed:   true,
	})
	if err != nil {
		log.Fatalf("could not update task: %v", err)
	}
	fmt.Printf("Updated task: %+v\n", updateRes)

	// List all tasks
	listRes, err := client.ListTasks(ctx, &pb.ListTasksRequest{})
	if err != nil {
		log.Fatalf("could not list tasks: %v", err)
	}
	fmt.Printf("List tasks: %+v\n", listRes.Tasks)

	// Delete the task.
	deleteRes, err := client.DeleteTask(ctx, &pb.DeleteTaskRequest{Id: createRes.Id})
	if err != nil {
		log.Fatalf("could not delete task: %v", err)
	}
	fmt.Printf("Delete response: %+v\n", deleteRes)

	// List all tasks after delete one task
	listRes, err = client.ListTasks(ctx, &pb.ListTasksRequest{})
	if err != nil {
		log.Fatalf("could not list tasks: %v", err)
	}
	fmt.Printf("List tasks: %+v\n", listRes.Tasks)
}
