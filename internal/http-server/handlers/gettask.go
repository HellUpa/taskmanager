package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/HellUpa/taskmanager/internal/app"
	"github.com/go-chi/chi/v5"
)

// getTaskHandler handles GET requests to retrieve a task by ID.
func GetTaskHandler(tm *app.TaskManagerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}

		task, err := tm.GetTask(r.Context(), int32(id))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get task: %v", err), http.StatusInternalServerError)
			return
		}

		if task == nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(task)
	}
}
