package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/JeanGrijp/cepweather/internal/weather"
)

// WeatherService exposes the use-case needed by the HTTP layer.
type WeatherService interface {
	GetByCEP(ctx context.Context, cep string) (weather.Temperatures, error)
}

// NewRouter constructs the HTTP router for the service.
func NewRouter(service WeatherService, logger *log.Logger) http.Handler {
	mux := http.NewServeMux()

	handler := &weatherHandler{
		service: service,
		logger:  logger,
	}

	mux.Handle("/weather/", handler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil && logger != nil {
			logger.Printf("failed to write healthz response: %v", err)
		}
	})

	return mux
}

type weatherHandler struct {
	service WeatherService
	logger  *log.Logger
}

func (h *weatherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "method not allowed"})
		return
	}

	cep := strings.TrimPrefix(r.URL.Path, "/weather/")
	if cep == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "not found"})
		return
	}

	temperatures, err := h.service.GetByCEP(r.Context(), cep)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, temperatures)
}

func (h *weatherHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, weather.ErrInvalidCEP):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
	case errors.Is(err, weather.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"message": err.Error()})
	default:
		if h.logger != nil {
			h.logger.Printf("unexpected error: %v", err)
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// Encoding errors are unexpected once headers are sent; nothing else to do.
	}
}
