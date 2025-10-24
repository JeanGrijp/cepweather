package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JeanGrijp/cepweather/internal/api"
	"github.com/JeanGrijp/cepweather/internal/viacep"
	"github.com/JeanGrijp/cepweather/internal/weather"
	"github.com/JeanGrijp/cepweather/internal/weatherapi"
)

const (
	defaultAddr          = ":8080"
	defaultViaCEPBaseURL = "https://viacep.com.br/ws"
	defaultWeatherAPIURL = "https://api.weatherapi.com/v1"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC)

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	viaCEPBaseURL := getenv("VIACEP_BASE_URL", defaultViaCEPBaseURL)
	weatherAPIBaseURL := getenv("WEATHER_API_BASE_URL", defaultWeatherAPIURL)
	weatherAPIKey := os.Getenv("WEATHER_API_KEY")
	if weatherAPIKey == "" {
		logger.Fatal("WEATHER_API_KEY environment variable is required")
	}

	locationClient := viacep.NewClient(httpClient, viaCEPBaseURL)
	weatherClient := weatherapi.NewClient(httpClient, weatherAPIBaseURL, weatherAPIKey)

	service := weather.NewService(locationClient, weatherClient)

	port := getenv("PORT", defaultAddr)
	// Cloud Run passa PORT sem ":", então adicionamos se necessário
	if port != "" && port[0] != ':' {
		port = ":" + port
	}

	server := &http.Server{
		Addr:    port,
		Handler: api.NewRouter(service, logger),
	}

	go func() {
		logger.Printf("starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server error: %v", err)
		}
	}()

	shutdown(server, logger)
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func shutdown(server *http.Server, logger *log.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Printf("graceful shutdown failed: %v", err)
	} else {
		logger.Println("server stopped gracefully")
	}
}
