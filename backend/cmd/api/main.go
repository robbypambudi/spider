package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/spider/spider/dao"
	"github.com/spider/spider/docs"
	"github.com/spider/spider/internal/app"
	"github.com/spider/spider/internal/handler"
	"github.com/spider/spider/internal/middleware"
	"github.com/spider/spider/internal/telemetry"
	"github.com/spider/spider/pkg/config"
	"github.com/spider/spider/pkg/monitor"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	settings, err := config.Load()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := dao.Connect(ctx, settings)
	if err != nil {
		slog.Error("database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	container, err := app.Build(settings, db)
	if err != nil {
		slog.Error("container", "error", err)
		os.Exit(1)
	}
	if err := app.Bootstrap(container); err != nil {
		slog.Error("bootstrap", "error", err)
		os.Exit(1)
	}

	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger())
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   settings.CORSOriginList(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Spider-Worker-Token"},
		AllowCredentials: true,
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		handler.WriteJSON(w, http.StatusOK, monitor.HealthStatus{
			Status: "ok", Service: telemetry.ServiceName, Version: telemetry.Version,
		})
	})
	r.Get("/ready", func(w http.ResponseWriter, _ *http.Request) {
		dbOK := db.Ready(context.Background())
		status := "ok"
		if !dbOK {
			status = "degraded"
		}
		handler.WriteJSON(w, http.StatusOK, monitor.ReadinessStatus{Status: status, Database: dbOK, Redis: nil})
	})
	r.Handle("/metrics", container.Metrics.Handler())
	docs.Mount(r)

	r.Route("/api/v1", func(api chi.Router) {
		handler.MountAPI(api, container)
	})

	addr := fmt.Sprintf("%s:%d", settings.APIHost, settings.APIPort)
	srv := &http.Server{Addr: addr, Handler: r, ReadHeaderTimeout: 10 * time.Second}

	go func() {
		slog.Info("api_started", "service", telemetry.ServiceName, "version", telemetry.Version, "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
