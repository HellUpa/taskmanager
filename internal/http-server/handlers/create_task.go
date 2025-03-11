package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HellUpa/taskmanager/internal/app"
	"github.com/HellUpa/taskmanager/internal/models"
	"github.com/HellUpa/taskmanager/internal/telemetry"
	"github.com/google/uuid"
)

// createTaskHandler handles POST requests to create a new task.
func CreateTaskHandler(tm *app.TaskManagerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(telemetry.UserIDKey).(uuid.UUID) // Get user ID from context
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var task models.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		id, err := tm.CreateTask(r.Context(), &task, userID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create task: %v", err), http.StatusInternalServerError)
			return
		}

		task.ID = id
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	}
}
