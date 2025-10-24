package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JeanGrijp/cepweather/internal/weather"
)

type stubService struct {
	temps   weather.Temperatures
	err     error
	lastCEP string
}

func (s *stubService) GetByCEP(ctx context.Context, cep string) (weather.Temperatures, error) {
	s.lastCEP = cep
	if s.err != nil {
		return weather.Temperatures{}, s.err
	}
	return s.temps, nil
}

func TestWeatherHandlerSuccess(t *testing.T) {
	stub := &stubService{
		temps: weather.Temperatures{
			Celsius:    28.5,
			Fahrenheit: 83.3,
			Kelvin:     301.5,
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/weather/12345678", nil)

	NewRouter(stub, log.New(io.Discard, "", 0)).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json content-type, got %s", ct)
	}

	var body weather.Temperatures
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if body != stub.temps {
		t.Fatalf("expected body %+v, got %+v", stub.temps, body)
	}

	if stub.lastCEP != "12345678" {
		t.Fatalf("expected CEP to be forwarded, got %s", stub.lastCEP)
	}
}

func TestWeatherHandlerInvalidCEP(t *testing.T) {
	stub := &stubService{err: weather.ErrInvalidCEP}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/weather/INVALID", nil)

	NewRouter(stub, log.New(io.Discard, "", 0)).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", recorder.Code)
	}

	assertMessage(t, recorder.Body.Bytes(), "invalid zipcode")
}

func TestWeatherHandlerNotFound(t *testing.T) {
	stub := &stubService{err: weather.ErrNotFound}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/weather/12345678", nil)

	NewRouter(stub, log.New(io.Discard, "", 0)).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}

	assertMessage(t, recorder.Body.Bytes(), "can not find zipcode")
}

func TestWeatherHandlerMethodNotAllowed(t *testing.T) {
	stub := &stubService{}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/weather/12345678", nil)

	NewRouter(stub, log.New(io.Discard, "", 0)).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", recorder.Code)
	}
}

func assertMessage(t *testing.T, body []byte, expected string) {
	t.Helper()
	var payload map[string]string
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if payload["message"] != expected {
		t.Fatalf("expected message %q, got %q", expected, payload["message"])
	}
}
