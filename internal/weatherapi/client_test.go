package weatherapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/JeanGrijp/cepweather/internal/weather"
)

type fakeRoundTripper func(*http.Request) (*http.Response, error)

func (f fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestCurrentTemperatureCSuccess(t *testing.T) {
	var receivedQuery url.Values

	rt := fakeRoundTripper(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/current.json" {
			t.Fatalf("unexpected path: %s", req.URL.Path)
		}
		receivedQuery = req.URL.Query()
		body := `{"current":{"temp_c":26.4}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient(&http.Client{Transport: rt}, "https://weather.test", "apikey")

	temp, err := client.CurrentTemperatureC(context.Background(), weather.Location{City: "São Paulo", State: "SP"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if temp != 26.4 {
		t.Fatalf("expected 26.4, got %.1f", temp)
	}

	if receivedQuery.Get("key") != "apikey" {
		t.Fatalf("expected api key to be set")
	}
	if q := receivedQuery.Get("q"); q != "São Paulo, SP" {
		t.Fatalf("unexpected location query: %s", q)
	}
}

func TestCurrentTemperatureCNotFound(t *testing.T) {
	rt := fakeRoundTripper(func(req *http.Request) (*http.Response, error) {
		payload, _ := json.Marshal(map[string]any{
			"error": map[string]any{
				"message": "No matching location found.",
			},
		})
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader(string(payload))),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient(&http.Client{Transport: rt}, "https://weather.test", "apikey")

	_, err := client.CurrentTemperatureC(context.Background(), weather.Location{City: "São Paulo"})
	if !errors.Is(err, weather.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCurrentTemperatureCUnexpectedError(t *testing.T) {
	rt := fakeRoundTripper(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("boom")),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient(&http.Client{Transport: rt}, "https://weather.test", "apikey")

	_, err := client.CurrentTemperatureC(context.Background(), weather.Location{City: "São Paulo"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
