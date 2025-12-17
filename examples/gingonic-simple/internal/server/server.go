package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twistingmercury/otelx"
	otelxgin "github.com/twistingmercury/otelx/middleware/gin"
	"github.com/twistingmercury/otelx_examples/gingonic/simple/internal/handlers"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Initialize configures the gin engine with middleware and routes.
func Initialize(engine *gin.Engine, tel *otelx.Telemetry) error {
	// Middleware
	engine.Use(gin.Recovery())
	engine.Use(otelgin.Middleware("uuid-api"))
	engine.Use(otelxgin.LoggingMiddleware(tel,
		otelxgin.WithSkipPaths("/health"),
	))

	// Routes
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	handler, err := handlers.NewHandler(tel)
	if err != nil {
		return err
	}
	engine.GET("/api/uuid", handler.GetUUID)

	return nil
}

// Run starts the HTTP server and blocks until context is cancelled.
func Run(ctx context.Context, engine *gin.Engine, tel *otelx.Telemetry) error {
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		tel.Logger.Info().Str("addr", srv.Addr).Msg("starting server")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		tel.Logger.Info().Msg("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
