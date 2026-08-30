package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spider/spider/dao"
	"github.com/spider/spider/dao/store"
	"github.com/spider/spider/internal/service"
	"github.com/spider/spider/internal/telemetry"
	"github.com/spider/spider/pkg/config"
	"github.com/spider/spider/pkg/security/detectors"
	"github.com/spider/spider/pkg/security/enforcement"
	"github.com/spider/spider/pkg/security/pipeline"
	"github.com/spider/spider/pkg/serving/providers"
	"github.com/spider/spider/pkg/serving/router"
)

type Container struct {
	Settings *config.Settings
	DB       *dao.DB
	Pool     *pgxpool.Pool

	Users      *store.UserRepo
	Policies   *store.PolicyRepo
	Workers    *store.WorkerRepo
	Scans      *store.SecurityRepo
	Inferences *store.InferenceRepo

	Pipeline *pipeline.Pipeline
	Enforcer *enforcement.Enforcer
	Router   *router.ServingRouter
	Provider providers.LLMProvider
	Metrics  *telemetry.Metrics

	Auth       *service.AuthService
	Security   *service.SecurityService
	Inference  *service.InferenceService
	Worker     *service.WorkerService
	Policy     *service.PolicyService
	Serving    *service.ServingService
	MetricsSvc *service.MetricsService
}

func buildProvider(settings *config.Settings) providers.LLMProvider {
	switch settings.ServingProvider {
	case "mock":
		return providers.NewMockLLMProvider()
	default:
		return providers.NewPromptShieldProvider(settings.PromptShieldEndpoint, settings.PromptShieldModel)
	}
}

func Build(settings *config.Settings, db *dao.DB) (*Container, error) {
	detectors.RegisterPromptShield(settings.PromptShieldEndpoint, settings.PromptShieldModel)
	detectors.RegisterPromptShieldEnsemble(settings.PromptShieldEndpoint, settings.PromptShieldModel)

	pipe, err := pipeline.BuildDefault(pipeline.BuildOptions{
		DetectorName:         settings.DefaultDetector,
		Threshold:            settings.DefaultThreshold,
		Chunker:              settings.Chunker,
		ChunkSize:            settings.ChunkSize,
		ChunkOverlap:         settings.ChunkOverlap,
		FailMode:             settings.FailMode,
		PromptShieldEndpoint: settings.PromptShieldEndpoint,
	})
	if err != nil {
		return nil, fmt.Errorf("build pipeline: %w", err)
	}

	provider := buildProvider(settings)
	var shieldProvider *providers.PromptShieldProvider
	if ps, ok := provider.(*providers.PromptShieldProvider); ok {
		shieldProvider = ps
	}
	servingRouter := router.New(provider, nil)
	metrics := telemetry.NewMetrics()

	pool := db.Pool
	users := store.NewUserRepo(pool)
	policies := store.NewPolicyRepo(pool)
	workers := store.NewWorkerRepo(pool)
	scans := store.NewSecurityRepo(pool)
	inferences := store.NewInferenceRepo(pool)

	authSvc := &service.AuthService{Users: users, Settings: settings}
	policySvc := &service.PolicyService{Policies: policies, Settings: settings, Pipeline: pipe}
	secSvc := &service.SecurityService{
		Pipeline: pipe, Policy: policySvc, Scans: scans, Settings: settings, Metrics: metrics,
	}
	infSvc := &service.InferenceService{
		Security: secSvc, Enforcer: pipe.Enforcer, Router: servingRouter,
		Inferences: inferences, Workers: workers, Metrics: metrics,
	}

	return &Container{
		Settings: settings, DB: db, Pool: pool,
		Users: users, Policies: policies, Workers: workers,
		Scans: scans, Inferences: inferences,
		Pipeline: pipe, Enforcer: pipe.Enforcer, Router: servingRouter,
		Provider: provider, Metrics: metrics,
		Auth: authSvc, Security: secSvc, Inference: infSvc,
		Worker: &service.WorkerService{Workers: workers, Token: settings.WorkerToken},
		Policy: policySvc,
		Serving: &service.ServingService{
			Workers: workers, Settings: settings, Pipeline: pipe, Provider: shieldProvider,
		},
		MetricsSvc: &service.MetricsService{Scans: scans, Workers: workers, Infer: inferences},
	}, nil
}

func Bootstrap(c *Container) error {
	if err := c.DB.Migrate(c.Settings); err != nil {
		return err
	}
	if err := store.Bootstrap(c.DB.Pool, c.Settings); err != nil {
		return err
	}
	return c.Policy.ApplyActivePolicy(context.Background())
}
