package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/HellUpa/taskmanager/internal/app"

	kratos "github.com/ory/kratos-client-go"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	//ScopesKey contextKey = "scopes"
)

// AuthMiddleware creates a middleware that authenticates requests using Kratos sessions.
func AuthMiddleware(kratosClient *kratos.APIClient, tm *app.TaskManagerService, ui_ip string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the session cookie. The cookie name is set by Kratos.
			cookie, err := r.Cookie("ory_kratos_session")
			if err != nil {
				tm.Log.Info("Unauthorized: no session cookie, redirect to login")
				http.Redirect(w, r, fmt.Sprintf("http://%v:4433/self-service/login/browser", ui_ip), http.StatusSeeOther)
				return
			}

			// Verify the session with Kratos.
			session, resp, err := kratosClient.FrontendAPI.ToSession(r.Context()).Cookie(cookie.String()).Execute()
			if err != nil {
				tm.Log.Warn("Kratos session verification failed", "error", err)
				if resp != nil {
					tm.Log.Warn("Kratos session response", "http_status", resp.StatusCode)
				}
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			defer resp.Body.Close()

			if !session.GetActive() {
				tm.Log.Warn("Kratos session is inactive")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract the Kratos ID.
			kratosID := session.Identity.Id
			// Get User by Kratos ID.
			user, err := tm.GetUserByKratosID(r.Context(), kratosID)
			if err != nil {
				tm.Log.Error("Failed to get user from db by Kratos ID", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if user == nil {
				tm.Log.Error("Failed to find user by Kratos ID, after success auth", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			tm.Log.Debug("User found by Kratos ID, write in ctx", "user_id", user.ID)
			// Store the user ID in the context.
			ctx := context.WithValue(r.Context(), UserIDKey, user.ID)

			// Call the next handler in the chain, with the updated context.
			next.ServeHTTP(w, r.WithContext(ctx))
			return

		})
	}
}
