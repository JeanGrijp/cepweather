package viacep

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JeanGrijp/cepweather/internal/weather"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestLookupSuccess(t *testing.T) {
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://example.com/12345678/json/" {
			t.Fatalf("unexpected URL: %s", req.URL.String())
		}
		body := `{"localidade":"São Paulo","uf":"SP"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient(&http.Client{Transport: rt}, "https://example.com")

	location, err := client.Lookup(context.Background(), "12345678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if location.City != "São Paulo" || location.State != "SP" {
		t.Fatalf("unexpected location: %+v", location)
	}
}

func TestLookupNotFound(t *testing.T) {
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"erro": true}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient(&http.Client{Transport: rt}, "https://example.com")

	_, err := client.Lookup(context.Background(), "12345678")
	if !errors.Is(err, weather.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestLookupUnexpectedStatus(t *testing.T) {
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("boom")),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient(&http.Client{Transport: rt}, "https://example.com")

	_, err := client.Lookup(context.Background(), "12345678")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
