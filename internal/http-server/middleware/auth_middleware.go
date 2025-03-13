package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/HellUpa/taskmanager/internal/app"

	hydra "github.com/ory/hydra-client-go/v2"
	kratos "github.com/ory/kratos-client-go"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	ScopesKey contextKey = "scopes"
)

// AuthMiddleware creates a middleware that authenticates requests using Kratos sessions.
func AuthMiddleware(kratosClient *kratos.APIClient, hydraClient *hydra.APIClient, tm *app.TaskManagerService, ui_ip string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
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
			}

			// Extract the token (if it's a Bearer token)
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}
			token := tokenParts[1]

			// Introspect the token with Hydra.
			introspectResult, resp, err := hydraClient.OAuth2API.IntrospectOAuth2Token(r.Context()).Token(token).Execute()

			if err != nil {
				tm.Log.Error("Hydra token introspection failed", "error", err)
				if resp != nil {
					tm.Log.Error("Kratos session response", "http_status", resp.StatusCode)
				}
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			defer resp.Body.Close()

			// Check if the token is active.
			if !introspectResult.GetActive() {
				tm.Log.Warn("Token is not active", "token", token)
				http.Error(w, "Unauthorized: Token is not active", http.StatusUnauthorized)
				return
			}

			// Extract user information (Kratos ID, scopes).
			kratosID, ok := introspectResult.GetExt()["sub"].(string) // "sub" claim usually contains the user ID
			if !ok || kratosID == "" {
				tm.Log.Error("Kratos ID missing in token introspection response")
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			// Get user id
			user, err := tm.GetUserByKratosID(r.Context(), kratosID)
			if err != nil {
				tm.Log.Error("Failed to get user from db by Kratos ID", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if user == nil {
				tm.Log.Error("Failed to find user by Kratos ID, after succes auth", "kratos_id", kratosID)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Convert the scopes to a slice of strings.
			scopes := introspectResult.GetScope()
			if scopes == "" {
				tm.Log.Warn("Scopes missing in token introspection response")
			}
			scopeList := strings.Split(scopes, " ") // Split string by space

			// 6. Store the user ID and scopes in the context.
			ctx := context.WithValue(r.Context(), UserIDKey, user.ID)
			ctx = context.WithValue(ctx, ScopesKey, scopeList)

			// Call the next handler in the chain, with the updated context.
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
