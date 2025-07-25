package api

import (
	"context"
	"fmt"
	"net/http"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/time/api/handlers"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/time/scheduler"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
	gootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// Server represents the API server
type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	logger     logging.Logger
}

// Config holds the server configuration
type Config struct {
	Port string
}

// Dependencies holds the server dependencies
type Dependencies struct {
	Logger    logging.Logger
	Scheduler *scheduler.TimeBasedScheduler
}

// NewServer creates a new API server
func NewServer(cfg Config, deps Dependencies) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Initialize OpenTelemetry tracer
	_, err := InitTracer()
	if err != nil {
		deps.Logger.Error("Failed to initialize OpenTelemetry tracer", "error", err)
	}

	// Create server instance
	srv := &Server{
		router: router,
		logger: deps.Logger,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%s", cfg.Port),
			Handler: router,
		},
	}

	// Setup middleware
	srv.setupMiddleware(deps)

	// Setup routes
	srv.setupRoutes(deps)

	return srv
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.Info("Starting API server", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping API server")
	return s.httpServer.Shutdown(ctx)
}

// setupMiddleware sets up the middleware for the server
func (s *Server) setupMiddleware(deps Dependencies) {
	s.router.Use(gin.Recovery())
	s.router.Use(TraceMiddleware())
	s.router.Use(MetricsMiddleware())
	s.router.Use(LoggerMiddleware(deps.Logger))
	s.router.Use(ErrorMiddleware(deps.Logger))
}

// setupRoutes sets up the routes for the server
func (s *Server) setupRoutes(deps Dependencies) {
	// Create handlers
	statusHandler := handlers.NewStatusHandler(deps.Logger)
	metricsHandler := handlers.NewMetricsHandler(deps.Logger)
	schedulerHandler := handlers.NewSchedulerHandler(deps.Logger, deps.Scheduler)

	// Health and monitoring endpoints
	s.router.GET("/status", statusHandler.Status)
	s.router.GET("/metrics", metricsHandler.Metrics)

	// Scheduler management endpoints
	api := s.router.Group("/api/v1")
	{
		api.GET("/scheduler/stats", schedulerHandler.GetStats)
	}
}

// InitTracer sets up OpenTelemetry tracing with OTLP exporter for Tempo
// Set TEMPO_OTLP_ENDPOINT env var to override the default (localhost:4318)
func InitTracer() (func(context.Context) error, error) {
	endpoint := os.Getenv("TEMPO_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4318" // default to local Tempo
	}
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("triggerx-scheduler-time"),
		)),
	)
	gootel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}
