package input

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

var cepPattern = regexp.MustCompile(`^\d{8}$`)

// Handler processes incoming CEP requests and forwards to Service B.
type Handler struct {
	serviceBURL string
	httpClient  *http.Client
	logger      *log.Logger
}

// NewHandler creates a new input service handler.
func NewHandler(serviceBURL string, httpClient *http.Client, logger *log.Logger) *Handler {
	return &Handler{
		serviceBURL: serviceBURL,
		httpClient:  httpClient,
		logger:      logger,
	}
}

type inputRequest struct {
	CEP string `json:"cep"`
}

type errorResponse struct {
	Message string `json:"message"`
}

// HandleCEP processes POST requests with CEP input.
func (h *Handler) HandleCEP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Message: "method not allowed"})
		return
	}

	var req inputRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, errorResponse{Message: "invalid request body"})
		return
	}

	// Validate CEP format (8 digits, string)
	if !cepPattern.MatchString(req.CEP) {
		h.writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Message: "invalid zipcode"})
		return
	}

	// Forward to Service B
	response, err := h.forwardToServiceB(r.Context(), req.CEP)
	if err != nil {
		h.logger.Printf("error forwarding to service B: %v", err)
		h.writeJSON(w, http.StatusInternalServerError, errorResponse{Message: "internal server error"})
		return
	}

	// Return Service B response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)
	if _, err := io.Copy(w, response.Body); err != nil {
		h.logger.Printf("error copying response: %v", err)
	}
	response.Body.Close()
}

func (h *Handler) forwardToServiceB(ctx context.Context, cep string) (*http.Response, error) {
	url := fmt.Sprintf("%s/weather/%s", h.serviceBURL, cep)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	return h.httpClient.Do(req)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Printf("error encoding response: %v", err)
	}
}
