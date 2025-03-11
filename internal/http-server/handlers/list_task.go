package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HellUpa/taskmanager/internal/app"
	"github.com/HellUpa/taskmanager/internal/telemetry"
	"github.com/google/uuid"
)

// listTasksHandler handles GET requests to list all tasks.
func ListTasksHandler(tm *app.TaskManagerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(telemetry.UserIDKey).(uuid.UUID) // Get user ID from context
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized) // No userID = unauthorized
			return
		}

		tasks, err := tm.ListTasks(r.Context(), userID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list tasks: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tasks)
	}
}
