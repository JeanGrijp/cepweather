package weather

import (
	"context"
	"errors"
	"testing"
)

type stubLocationProvider struct {
	location Location
	err      error
}

func (s stubLocationProvider) Lookup(ctx context.Context, cep string) (Location, error) {
	return s.location, s.err
}

type stubTemperatureProvider struct {
	temp float64
	err  error
}

func (s stubTemperatureProvider) CurrentTemperatureC(ctx context.Context, location Location) (float64, error) {
	return s.temp, s.err
}

func TestServiceGetByCEPSuccess(t *testing.T) {
	service := NewService(
		stubLocationProvider{location: Location{City: "São Paulo", State: "SP"}},
		stubTemperatureProvider{temp: 25.2},
	)

	temps, err := service.GetByCEP(context.Background(), "12345678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFloat(t, temps.Celsius, 25.2)
	assertFloat(t, temps.Fahrenheit, 77.4)
	assertFloat(t, temps.Kelvin, 298.2)
}

func TestServiceGetByCEPInvalidFormat(t *testing.T) {
	service := NewService(
		stubLocationProvider{},
		stubTemperatureProvider{},
	)

	_, err := service.GetByCEP(context.Background(), "abcd5678")
	if !errors.Is(err, ErrInvalidCEP) {
		t.Fatalf("expected ErrInvalidCEP, got %v", err)
	}
}

func TestServiceGetByCEPNotFound(t *testing.T) {
	service := NewService(
		stubLocationProvider{err: ErrNotFound},
		stubTemperatureProvider{},
	)

	_, err := service.GetByCEP(context.Background(), "12345678")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServicePropagatesLocationError(t *testing.T) {
	wantErr := errors.New("network down")
	service := NewService(
		stubLocationProvider{err: wantErr},
		stubTemperatureProvider{},
	)

	_, err := service.GetByCEP(context.Background(), "12345678")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestServicePropagatesTemperatureError(t *testing.T) {
	wantErr := errors.New("weather timeout")
	service := NewService(
		stubLocationProvider{location: Location{City: "São Paulo", State: "SP"}},
		stubTemperatureProvider{err: wantErr},
	)

	_, err := service.GetByCEP(context.Background(), "12345678")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	if got != want {
		t.Fatalf("expected %.1f got %.1f", want, got)
	}
}
