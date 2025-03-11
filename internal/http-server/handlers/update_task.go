package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/HellUpa/taskmanager/internal/app"
	"github.com/HellUpa/taskmanager/internal/models"
	"github.com/HellUpa/taskmanager/internal/telemetry"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// updateTaskHandler handles PUT requests to update an existing task.
func UpdateTaskHandler(tm *app.TaskManagerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(telemetry.UserIDKey).(uuid.UUID) // Get user ID from context
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}

		var task models.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		task.ID = int32(id)
		task.UserID = userID

		if err := tm.UpdateTask(r.Context(), &task); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to update task: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(task)
	}
}
