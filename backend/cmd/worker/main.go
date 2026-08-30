package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/config"
	"github.com/spider/spider/pkg/runtime"
	"github.com/spider/spider/pkg/workerctl"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	settings, err := config.Load()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	opts := workerctl.RunOptions{
		Settings: settings,
		Models:   promptShieldModels(settings),
		Metadata: map[string]interface{}{"role": "prompt-shield-serving"},
	}
	if err := workerctl.Run(ctx, opts); err != nil {
		slog.Error("worker", "error", err)
		os.Exit(1)
	}
}

func promptShieldModels(settings *config.Settings) []apis.LoadedModel {
	active := settings.PromptShieldModel
	models := make([]apis.LoadedModel, 0, len(config.PromptShieldCatalog))
	for _, m := range config.PromptShieldCatalog {
		status := runtime.ModelStatusReady
		if m.ID != active {
			status = "AVAILABLE"
		}
		models = append(models, apis.LoadedModel{Name: m.ID, Status: status})
	}
	return models
}
