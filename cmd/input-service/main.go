package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JeanGrijp/cepweather/internal/input"
	"github.com/JeanGrijp/cepweather/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	defaultAddr        = ":8081"
	defaultServiceBURL = "http://localhost:8080"
	defaultZipkinURL   = "http://zipkin:9411/api/v2/spans"
)

func main() {
	logger := log.New(os.Stdout, "[SERVICE-A] ", log.LstdFlags|log.LUTC)

	// Initialize OpenTelemetry
	zipkinURL := getenv("ZIPKIN_URL", defaultZipkinURL)
	shutdown, err := telemetry.InitTracer("service-a", zipkinURL)
	if err != nil {
		logger.Printf("failed to initialize tracer: %v", err)
	} else {
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				logger.Printf("failed to shutdown tracer: %v", err)
			}
		}()
		logger.Println("OpenTelemetry initialized with Zipkin exporter")
	}

	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	serviceBURL := getenv("SERVICE_B_URL", defaultServiceBURL)
	handler := input.NewHandler(serviceBURL, httpClient, logger)

	mux := http.NewServeMux()
	mux.Handle("/", otelhttp.NewHandler(http.HandlerFunc(handler.HandleCEP), "handle-cep"))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			logger.Printf("failed to write healthz response: %v", err)
		}
	})

	port := getenv("PORT", defaultAddr)
	if port != "" && port[0] != ':' {
		port = ":" + port
	}

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	go func() {
		logger.Printf("starting input service on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server error: %v", err)
		}
	}()

	shutdownServer(server, logger)
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func shutdownServer(server *http.Server, logger *log.Logger) {
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
