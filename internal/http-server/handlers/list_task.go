package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HellUpa/taskmanager/internal/app"
	middlewares "github.com/HellUpa/taskmanager/internal/http-server/middleware"
	"github.com/google/uuid"
)

// listTasksHandler handles GET requests to list all tasks.
func ListTasksHandler(tm *app.TaskManagerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middlewares.UserIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tm.Log.Debug("List task handler for", "userID", userID)

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
