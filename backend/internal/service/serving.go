package service

import (
	"context"
	"fmt"

	"github.com/spider/spider/dao/store"
	"github.com/spider/spider/internal/spidererrors"
	"github.com/spider/spider/pkg/config"
	"github.com/spider/spider/pkg/security/detectors"
	"github.com/spider/spider/pkg/security/pipeline"
	"github.com/spider/spider/pkg/serving/providers"
)

type ServingService struct {
	Workers  *store.WorkerRepo
	Settings *config.Settings
	Pipeline *pipeline.Pipeline
	Provider *providers.PromptShieldProvider
}

func (s *ServingService) ListNodes(ctx context.Context) ([]map[string]interface{}, error) {
	return s.Workers.ListServingNodes(ctx)
}

func (s *ServingService) ListModels(ctx context.Context) ([]map[string]interface{}, error) {
	registered, err := s.Workers.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	active := s.Settings.PromptShieldModel
	byName := map[string]map[string]interface{}{}
	for _, row := range registered {
		name, _ := row["name"].(string)
		byName[name] = row
	}

	out := make([]map[string]interface{}, 0, len(config.PromptShieldCatalog))
	for _, catalog := range config.PromptShieldCatalog {
		entry := map[string]interface{}{
			"id":          catalog.ID,
			"name":        catalog.Name,
			"params":      catalog.Params,
			"description": catalog.Description,
			"active":      catalog.ID == active,
			"status":      "AVAILABLE",
			"worker_id":   nil,
		}
		if reg, ok := byName[catalog.ID]; ok {
			entry["status"] = reg["status"]
			entry["worker_id"] = reg["worker_id"]
		}
		out = append(out, entry)
	}
	return out, nil
}

func (s *ServingService) Catalog() []map[string]interface{} {
	active := s.Settings.PromptShieldModel
	out := make([]map[string]interface{}, 0, len(config.PromptShieldCatalog))
	for _, m := range config.PromptShieldCatalog {
		out = append(out, map[string]interface{}{
			"id":          m.ID,
			"name":        m.Name,
			"params":      m.Params,
			"description": m.Description,
			"active":      m.ID == active,
			"collection":  "https://huggingface.co/collections/robbypambudi/prompt-shield",
		})
	}
	return out
}

func (s *ServingService) ActivateModel(ctx context.Context, modelID string) (map[string]interface{}, error) {
	if !config.IsPromptShieldModel(modelID) {
		return nil, spidererrors.Validation(fmt.Sprintf("unknown model %q", modelID))
	}
	s.Settings.PromptShieldModel = modelID
	if s.Provider != nil {
		s.Provider.SetModel(modelID)
	}
	if det, ok := s.Pipeline.Detector.(*detectors.PromptShieldDetector); ok {
		det.Model = modelID
	}
	_ = ctx
	return map[string]interface{}{
		"active_model": modelID,
		"message":      "Prompt-Shield model activated for detection and inference",
	}, nil
}
