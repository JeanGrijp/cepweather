package viacep

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/JeanGrijp/cepweather/internal/weather"
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.lookupURL(cep), http.NoBody)
	if err != nil {
		return weather.Location{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return weather.Location{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return weather.Location{}, fmt.Errorf("viacep: unexpected status %d", resp.StatusCode)
	}

	var payload struct {
		Localidade string `json:"localidade"`
		UF         string `json:"uf"`
		Erro       bool   `json:"erro"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return weather.Location{}, err
	}

	if payload.Erro || payload.Localidade == "" {
		return weather.Location{}, weather.ErrNotFound
	}

	return weather.Location{
		City:  payload.Localidade,
		State: payload.UF,
	}, nil
}

func (c *Client) lookupURL(cep string) string {
	return fmt.Sprintf("%s/%s/json/", strings.TrimSuffix(c.baseURL, "/"), cep)
}
