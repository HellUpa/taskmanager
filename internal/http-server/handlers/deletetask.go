package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/HellUpa/taskmanager/internal/app"
	"github.com/go-chi/chi/v5"
)

// deleteTaskHandler handles DELETE requests to delete a task.
func DeleteTaskHandler(tm *app.TaskManagerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}

		if err := tm.DeleteTask(r.Context(), int32(id)); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to delete task: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
