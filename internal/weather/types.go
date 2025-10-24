package weather

import "errors"

// ErrInvalidCEP indicates the provided CEP is malformed.
var ErrInvalidCEP = errors.New("invalid zipcode")

// ErrNotFound indicates that the CEP could not be resolved.
var ErrNotFound = errors.New("can not find zipcode")

// Location represents the city and state resolved from a CEP.
type Location struct {
	City  string
	State string
}

// Temperatures holds the temperatures in three units of measurement.
type Temperatures struct {
	Celsius    float64 `json:"temp_C"`
	Fahrenheit float64 `json:"temp_F"`
	Kelvin     float64 `json:"temp_K"`
}
