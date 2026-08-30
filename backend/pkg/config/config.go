package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Settings struct {
	Env                     string
	DatabaseURL             string
	RedisURL                string
	APIHost                 string
	APIPort                 int
	JWTSecret               string
	JWTAlgorithm            string
	JWTExpireMinutes        int
	WorkerToken             string
	CORSOrigins             string
	DefaultDetector         string
	DefaultSecurityPolicy   string
	DefaultThreshold        float64
	FailMode                string
	LogPromptContent        bool
	PersistPromptContent    bool
	WorkerHeartbeatInterval int
	WorkerOfflineTimeout    int
	Chunker                 string
	ChunkSize               int
	ChunkOverlap            int
	DefaultModel            string
	ServingProvider         string
	BootstrapAdminEmail     string
	BootstrapAdminPassword  string
	APIBaseURL              string
	PromptShieldEndpoint    string
	PromptShieldModel       string
	HFToken                 string
}

func Load() (*Settings, error) {
	s := &Settings{
		Env:                     getEnv("SPIDER_ENV", "development"),
		DatabaseURL:             normalizeDatabaseURL(getEnv("DATABASE_URL", "postgres://spider:spider@localhost:5432/spider?sslmode=disable")),
		RedisURL:                getEnv("REDIS_URL", "redis://localhost:6379/0"),
		APIHost:                 getEnv("SPIDER_API_HOST", "0.0.0.0"),
		APIPort:                 getEnvInt("SPIDER_API_PORT", 8000),
		JWTSecret:               getEnv("SPIDER_JWT_SECRET", "change-me"),
		JWTAlgorithm:            getEnv("SPIDER_JWT_ALGORITHM", "HS256"),
		JWTExpireMinutes:        getEnvInt("SPIDER_JWT_EXPIRE_MINUTES", 1440),
		WorkerToken:             getEnv("SPIDER_WORKER_TOKEN", "development-worker-token"),
		CORSOrigins:             getEnv("SPIDER_CORS_ORIGINS", "http://localhost:5173"),
		DefaultDetector:         getEnv("SPIDER_DEFAULT_DETECTOR", "prompt-shield+rules"),
		DefaultSecurityPolicy:   getEnv("SPIDER_DEFAULT_SECURITY_POLICY", "threshold"),
		DefaultThreshold:        getEnvFloat("SPIDER_DEFAULT_THRESHOLD", 0.5),
		FailMode:                getEnv("SPIDER_FAIL_MODE", "closed"),
		LogPromptContent:        getEnvBool("SPIDER_LOG_PROMPT_CONTENT", false),
		PersistPromptContent:    getEnvBool("SPIDER_PERSIST_PROMPT_CONTENT", false),
		WorkerHeartbeatInterval: getEnvInt("SPIDER_WORKER_HEARTBEAT_INTERVAL", 10),
		WorkerOfflineTimeout:    getEnvInt("SPIDER_WORKER_OFFLINE_TIMEOUT", 30),
		Chunker:                 getEnv("SPIDER_CHUNKER", "token"),
		ChunkSize:               getEnvInt("SPIDER_CHUNK_SIZE", 256),
		ChunkOverlap:            getEnvInt("SPIDER_CHUNK_OVERLAP", 0),
		DefaultModel:            getEnv("SPIDER_DEFAULT_MODEL", PromptShieldSmall),
		ServingProvider:         getEnv("SPIDER_SERVING_PROVIDER", "prompt-shield"),
		BootstrapAdminEmail:     getEnv("SPIDER_BOOTSTRAP_ADMIN_EMAIL", "admin@spider.local"),
		BootstrapAdminPassword:  getEnv("SPIDER_BOOTSTRAP_ADMIN_PASSWORD", "spider-admin"),
		APIBaseURL:              getEnv("SPIDER_API_BASE_URL", "http://localhost:8000"),
		PromptShieldEndpoint:    getEnv("SPIDER_PROMPT_SHIELD_ENDPOINT", "http://localhost:8081"),
		PromptShieldModel:       getEnv("SPIDER_PROMPT_SHIELD_MODEL", PromptShieldSmall),
		HFToken:                 getEnv("HF_TOKEN", ""),
	}
	if s.DefaultThreshold < 0 || s.DefaultThreshold > 1 {
		return nil, fmt.Errorf("SPIDER_DEFAULT_THRESHOLD must be between 0.0 and 1.0")
	}
	return s, nil
}

func (s *Settings) CORSOriginList() []string {
	parts := strings.Split(s.CORSOrigins, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func (s *Settings) IsDevelopment() bool {
	switch strings.ToLower(s.Env) {
	case "development", "dev", "test":
		return true
	default:
		return false
	}
}

func normalizeDatabaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Replace(raw, "postgresql+asyncpg://", "postgres://", 1)
	if strings.HasPrefix(raw, "postgresql://") {
		raw = "postgres://" + strings.TrimPrefix(raw, "postgresql://")
	}
	if !strings.Contains(raw, "sslmode=") {
		if strings.Contains(raw, "?") {
			raw += "&sslmode=disable"
		} else {
			raw += "?sslmode=disable"
		}
	}
	return raw
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
