package weather

import (
	"context"
	"math"
	"regexp"
	"strings"
)

var cepPattern = regexp.MustCompile(`^\d{8}$`)

// LocationProvider resolves a CEP into a geographic location.
type LocationProvider interface {
	Lookup(ctx context.Context, cep string) (Location, error)
}

// TemperatureProvider retrieves the current temperature in Celsius for a location.
type TemperatureProvider interface {
	CurrentTemperatureC(ctx context.Context, location Location) (float64, error)
}

// Service orchestrates location lookup and temperature retrieval.
type Service struct {
	locationProvider    LocationProvider
	temperatureProvider TemperatureProvider
}

// NewService constructs a Service with the given dependencies.
func NewService(location LocationProvider, temperature TemperatureProvider) *Service {
	return &Service{
		locationProvider:    location,
		temperatureProvider: temperature,
	}
}

// GetByCEP resolves the location for a CEP and returns the current temperatures.
func (s *Service) GetByCEP(ctx context.Context, cep string) (Temperatures, error) {
	cleanCEP, err := normalizeCEP(cep)
	if err != nil {
		return Temperatures{}, err
	}

	location, err := s.locationProvider.Lookup(ctx, cleanCEP)
	if err != nil {
		return Temperatures{}, err
	}

	celsius, err := s.temperatureProvider.CurrentTemperatureC(ctx, location)
	if err != nil {
		return Temperatures{}, err
	}

	return newTemperatures(celsius), nil
}

func normalizeCEP(cep string) (string, error) {
	trimmed := strings.TrimSpace(cep)
	if !cepPattern.MatchString(trimmed) {
		return "", ErrInvalidCEP
	}
	return trimmed, nil
}

func newTemperatures(celsius float64) Temperatures {
	fahrenheit := celsius*1.8 + 32
	kelvin := celsius + 273

	return Temperatures{
		Celsius:    roundToSingleDecimal(celsius),
		Fahrenheit: roundToSingleDecimal(fahrenheit),
		Kelvin:     roundToSingleDecimal(kelvin),
	}
}

func roundToSingleDecimal(value float64) float64 {
	return math.Round(value*10) / 10
}
