package viacep

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/JeanGrijp/cepweather/internal/weather"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Client implements weather.LocationProvider using the ViaCEP API.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a ViaCEP client with the provided HTTP client and base URL.
func NewClient(httpClient *http.Client, baseURL string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
	}
}

// Lookup resolves a CEP to a Location using ViaCEP.
func (c *Client) Lookup(ctx context.Context, cep string) (weather.Location, error) {
	tracer := otel.Tracer("viacep-client")
	ctx, span := tracer.Start(ctx, "viacep.Lookup",
		trace.WithAttributes(attribute.String("cep", cep)))
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.lookupURL(cep), http.NoBody)
	if err != nil {
		span.RecordError(err)
		return weather.Location{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return weather.Location{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("viacep: unexpected status %d", resp.StatusCode)
		span.RecordError(err)
		return weather.Location{}, err
	}

	var payload struct {
		Localidade string `json:"localidade"`
		UF         string `json:"uf"`
		Erro       any    `json:"erro"` // ViaCEP pode retornar bool ou string "true"
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		span.RecordError(err)
		return weather.Location{}, err
	}

	// ViaCEP retorna "erro": "true" (string) ou "erro": true (bool) quando o CEP não existe
	hasError := false
	if payload.Erro != nil {
		switch v := payload.Erro.(type) {
		case bool:
			hasError = v
		case string:
			hasError = v == "true"
		}
	}

	if hasError || payload.Localidade == "" {
		span.RecordError(weather.ErrNotFound)
		return weather.Location{}, weather.ErrNotFound
	}

	location := weather.Location{
		City:  payload.Localidade,
		State: payload.UF,
	}

	span.SetAttributes(
		attribute.String("city", location.City),
		attribute.String("state", location.State),
	)

	return location, nil
}

func (c *Client) lookupURL(cep string) string {
	return fmt.Sprintf("%s/%s/json/", strings.TrimSuffix(c.baseURL, "/"), cep)
}
