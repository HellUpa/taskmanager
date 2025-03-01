package telemetry

import (
	"fmt"
	"net/http"
)

// HealthCheckHandler проверяет состояние сервиса.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
