package weatherapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/JeanGrijp/cepweather/internal/weather"
)

// Client implements weather.TemperatureProvider using WeatherAPI.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewClient creates a WeatherAPI client.
func NewClient(httpClient *http.Client, baseURL, apiKey string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		apiKey:     apiKey,
	}
}

// CurrentTemperatureC fetches the current Celsius temperature for the given location.
func (c *Client) CurrentTemperatureC(ctx context.Context, location weather.Location) (float64, error) {
	endpoint := fmt.Sprintf("%s/current.json", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return 0, err
	}

	q := req.URL.Query()
	q.Set("key", c.apiKey)
	q.Set("q", buildQuery(location))
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, c.handleErrorResponse(resp)
	}

	var payload struct {
		Current struct {
			TempC float64 `json:"temp_c"`
		} `json:"current"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}

	return payload.Current.TempC, nil
}

func buildQuery(location weather.Location) string {
	values := []string{location.City}
	if location.State != "" {
		values = append(values, location.State)
	}
	return strings.Join(values, ", ")
}

func (c *Client) handleErrorResponse(resp *http.Response) error {
	var payload struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("weatherapi: unexpected status %d", resp.StatusCode)
	}

	message := strings.ToLower(payload.Error.Message)
	if resp.StatusCode == http.StatusBadRequest && strings.Contains(message, "no matching location found") {
		return weather.ErrNotFound
	}

	return fmt.Errorf("weatherapi: %s", payload.Error.Message)
}
