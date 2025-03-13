package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HellUpa/taskmanager/internal/app"
	"github.com/HellUpa/taskmanager/internal/models"

	"github.com/google/uuid"
)

// KratosWebhookPayload represents the payload sent by Kratos in the webhook.
// This structure needs to match the structure of the webhook payload sent by Kratos.
// Refer to the Kratos documentation for the exact structure.
type KratosWebhookPayload struct {
	ID string `json:"userId"`
}

// KratosRegistrationWebhookHandler handles webhooks from Kratos after user registration.
func KratosRegistrationWebhookHandler(tm *app.TaskManagerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Verify the request (check Authorization header, if you set one up)
		//    (Omitted for brevity in this example, but VERY important in production)

		var payload KratosWebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		kratosID := payload.ID
		if kratosID == "" {
			http.Error(w, "Kratos ID missing in webhook payload", http.StatusBadRequest)
			return
		}

		newUser := &models.User{
			ID:       uuid.New(),
			KratosID: kratosID,
		}

		if err := tm.CreateUser(r.Context(), newUser); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
