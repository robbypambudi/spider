package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/spider/spider/dao/store"
	"github.com/spider/spider/internal/app"
	"github.com/spider/spider/internal/middleware"
	"github.com/spider/spider/internal/service"
	"github.com/spider/spider/internal/spidererrors"
	"github.com/spider/spider/pkg/apis"
)

func MountAPI(r chi.Router, c *app.Container) {
	authMW := middleware.Auth(c.Auth)
	workerMW := middleware.WorkerAuth(c.Settings.WorkerToken)

	r.Post("/auth/login", loginHandler(c))
	r.With(authMW).Get("/auth/me", meHandler())

	r.With(authMW).Post("/security/scan", scanHandler(c))
	r.With(authMW).Post("/security/scan/pdf", scanPDFHandler(c))
	r.With(authMW).Get("/security/scans", listScansHandler(c))
	r.With(authMW).Get("/security/scans/{scan_id}", getScanHandler(c))
	r.With(authMW).Get("/security/detectors", listDetectorsHandler(c))
	r.With(authMW).Get("/security/policies", listPoliciesHandler(c))
	r.With(authMW).Get("/security/policies/{policy_id}", getPolicyHandler(c))
	r.With(authMW).Post("/security/policies", createPolicyHandler(c))
	r.With(authMW).Put("/security/policies/{policy_id}", updatePolicyHandler(c))
	r.With(authMW).Delete("/security/policies/{policy_id}", deletePolicyHandler(c))
	r.With(authMW).Post("/security/policies/{policy_id}/activate", activatePolicyHandler(c))
	r.With(authMW).Get("/settings/runtime", runtimeSettingsHandler(c))
	r.With(authMW).Post("/security/evaluate", evaluateHandler(c))

	r.With(authMW).Post("/inference", inferHandler(c))
	r.With(authMW).Get("/inference", listInferenceHandler(c))

	r.With(workerMW).Post("/workers/register", registerWorkerHandler(c))
	r.With(workerMW).Post("/workers/{worker_id}/heartbeat", heartbeatHandler(c))
	r.With(authMW).Get("/workers", listWorkersHandler(c))
	r.With(authMW).Get("/workers/{worker_id}", getWorkerHandler(c))
	r.With(authMW).Put("/workers/{worker_id}", updateWorkerHandler(c))
	r.With(authMW).Delete("/workers/{worker_id}", deleteWorkerHandler(c))
	r.With(authMW).Post("/workers/prune-offline", pruneOfflineWorkersHandler(c))

	r.With(authMW).Get("/serving/nodes", servingNodesHandler(c))
	r.With(authMW).Get("/serving/models", servingModelsHandler(c))
	r.With(authMW).Get("/serving/catalog", servingCatalogHandler(c))
	r.With(authMW).Post("/serving/models/activate", activateModelHandler(c))
	r.With(authMW).Get("/metrics/summary", metricsSummaryHandler(c))
	r.With(authMW).Get("/jobs", jobsHandler())
}

func decodeJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return spidererrors.Validation("Invalid request body")
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return spidererrors.Validation("Invalid JSON")
	}
	return nil
}

func loginHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body apis.LoginRequest
		if err := decodeJSON(r, &body); err != nil {
			WriteError(w, err)
			return
		}
		user, err := c.Auth.Authenticate(r.Context(), body.Email, body.Password)
		if err != nil {
			WriteError(w, err)
			return
		}
		token, err := c.Auth.IssueToken(user)
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, apis.TokenResponse{
			AccessToken: token, TokenType: "bearer", Role: user.Role, Email: user.Email,
		})
	}
}

func meHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())
		WriteJSON(w, http.StatusOK, map[string]string{
			"id": user.ID.String(), "email": user.Email, "role": user.Role,
		})
	}
}

func toScanResponse(result apis.SecurityResult, scanID *string) apis.ScanResponse {
	var threshold *float64
	if v, ok := result.Metadata["threshold"].(float64); ok {
		threshold = &v
	}
	var model *string
	if v, ok := result.Metadata["model"].(string); ok {
		model = &v
	}
	views := make([]apis.DetectorView, 0, len(result.DetectorResults))
	for _, item := range result.DetectorResults {
		lat := item.LatencyMs
		views = append(views, apis.DetectorView{
			Detector: item.Detector, Score: item.Score, IsInjection: item.IsInjection,
			LatencyMs: &lat, Metadata: item.Metadata,
		})
	}
	return apis.ScanResponse{
		ScanID: scanID, RequestID: result.RequestID, Decision: string(result.Decision),
		Score: result.Score, ChunksScanned: result.ChunksScanned, Policy: result.Policy,
		Threshold: threshold, LatencyMs: result.TotalLatencyMs, Detectors: views, Model: model,
	}
}

func inspectOpts(user *store.User, source string, model *string, meta map[string]interface{}, persist bool) service.InspectOptions {
	opts := service.InspectOptions{Source: &source, Model: model, Metadata: meta, Persist: persist}
	if user != nil {
		opts.UserID = &user.ID
	}
	return opts
}

func scanHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body apis.ScanRequest
		if err := decodeJSON(r, &body); err != nil {
			WriteError(w, err)
			return
		}
		user, _ := middleware.UserFromContext(r.Context())
		source := "api"
		if body.Source != nil {
			source = *body.Source
		}
		result, stored, err := c.Security.Inspect(r.Context(), body.Text, inspectOpts(user, source, body.Model, body.Metadata, true))
		if err != nil {
			WriteError(w, err)
			return
		}
		var scanID *string
		if stored != nil {
			s := stored.ID.String()
			scanID = &s
		}
		WriteJSON(w, http.StatusOK, toScanResponse(result, scanID))
	}
}

func listScansHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := queryInt(r, "limit", 50)
		offset := queryInt(r, "offset", 0)
		rows, err := c.Scans.ListScans(r.Context(), limit, offset)
		if err != nil {
			WriteError(w, err)
			return
		}
		out := make([]map[string]interface{}, 0, len(rows))
		for _, row := range rows {
			out = append(out, map[string]interface{}{
				"id": row.ID.String(), "request_id": row.RequestID.String(),
				"decision": row.Decision, "score": row.Score, "threshold": row.Threshold,
				"detector": row.Detector, "policy": row.Policy, "chunks_scanned": row.ChunksScanned,
				"chunking_strategy": row.ChunkingStrategy, "latency_ms": row.LatencyMs,
				"prompt_hash": row.PromptHash, "prompt_length": row.PromptLength,
				"model_target": row.ModelTarget, "created_at": row.CreatedAt.Format(time.RFC3339),
			})
		}
		WriteJSON(w, http.StatusOK, out)
	}
}

func getScanHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "scan_id"))
		if err != nil {
			WriteError(w, spidererrors.NotFound("Invalid scan id"))
			return
		}
		row, err := c.Scans.Get(r.Context(), id)
		if err != nil {
			WriteError(w, spidererrors.NotFound("Scan not found"))
			return
		}
		chunks, _ := c.Scans.ChunkResults(r.Context(), id)
		execs, _ := c.Scans.DetectorExecutions(r.Context(), id)
		chunkList := make([]map[string]interface{}, 0, len(chunks))
		for _, ch := range chunks {
			chunkList = append(chunkList, map[string]interface{}{
				"index": ch.ChunkIndex, "detector": ch.Detector, "score": ch.Score,
				"is_injection": ch.IsInjection, "latency_ms": ch.LatencyMs,
			})
		}
		detList := make([]map[string]interface{}, 0, len(execs))
		for _, d := range execs {
			var meta map[string]interface{}
			_ = json.Unmarshal([]byte(d.MetadataJSON), &meta)
			entry := map[string]interface{}{
				"detector": d.Detector, "version": d.DetectorVersion, "score": d.Score,
				"is_injection": d.IsInjection, "latency_ms": d.LatencyMs, "threshold": d.Threshold,
				"metadata": meta,
			}
			if meta != nil {
				if snip, ok := meta["snippet"]; ok {
					entry["snippet"] = snip
				}
				if ct, ok := meta["chunk_text"]; ok {
					entry["chunk_text"] = ct
				}
				if ci, ok := meta["chunk_index"]; ok {
					entry["chunk_index"] = ci
				}
			}
			detList = append(detList, entry)
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"id": row.ID.String(), "request_id": row.RequestID.String(),
			"decision": row.Decision, "score": row.Score, "threshold": row.Threshold,
			"detector": row.Detector, "detector_version": row.DetectorVersion,
			"policy": row.Policy, "chunks_scanned": row.ChunksScanned,
			"chunking_strategy": row.ChunkingStrategy, "latency_ms": row.LatencyMs,
			"prompt_hash": row.PromptHash, "prompt_length": row.PromptLength,
			"model_target": row.ModelTarget, "worker_id": row.WorkerID,
			"created_at": row.CreatedAt.Format(time.RFC3339),
			"chunks": chunkList, "detectors": detList,
		})
	}
}

func listDetectorsHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, c.Security.ListDetectors())
	}
}

func listPoliciesHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := c.Policy.ListPolicies(r.Context())
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, items)
	}
}

func requireAdmin(c *app.Container, r *http.Request) (*store.User, error) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		return nil, spidererrors.Authentication("")
	}
	if err := c.Auth.RequireRoles(user, "ADMIN"); err != nil {
		return nil, err
	}
	return user, nil
}

func getPolicyHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "policy_id"))
		if err != nil {
			WriteError(w, spidererrors.Validation("invalid policy_id"))
			return
		}
		item, err := c.Policy.GetPolicy(r.Context(), id)
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, item)
	}
}

func createPolicyHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := requireAdmin(c, r); err != nil {
			WriteError(w, err)
			return
		}
		var body apis.CreatePolicyRequest
		if err := decodeJSON(r, &body); err != nil {
			WriteError(w, err)
			return
		}
		item, err := c.Policy.CreatePolicy(r.Context(), body)
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusCreated, item)
	}
}

func updatePolicyHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := requireAdmin(c, r); err != nil {
			WriteError(w, err)
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "policy_id"))
		if err != nil {
			WriteError(w, spidererrors.Validation("invalid policy_id"))
			return
		}
		var body apis.UpdatePolicyRequest
		if err := decodeJSON(r, &body); err != nil {
			WriteError(w, err)
			return
		}
		item, err := c.Policy.UpdatePolicy(r.Context(), id, body)
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, item)
	}
}

func deletePolicyHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := requireAdmin(c, r); err != nil {
			WriteError(w, err)
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "policy_id"))
		if err != nil {
			WriteError(w, spidererrors.Validation("invalid policy_id"))
			return
		}
		if err := c.Policy.DeletePolicy(r.Context(), id); err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

func activatePolicyHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := requireAdmin(c, r); err != nil {
			WriteError(w, err)
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "policy_id"))
		if err != nil {
			WriteError(w, spidererrors.Validation("invalid policy_id"))
			return
		}
		item, err := c.Policy.SetDefaultPolicy(r.Context(), id)
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, item)
	}
}

func runtimeSettingsHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, apis.RuntimeSettingsView{
			DefaultDetector: c.Settings.DefaultDetector,
			FailMode:        c.Settings.FailMode,
			Chunker:         c.Settings.Chunker,
			ChunkSize:       c.Settings.ChunkSize,
			ChunkOverlap:    c.Settings.ChunkOverlap,
		})
	}
}

func evaluateHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body apis.EvaluateRequest
		if err := decodeJSON(r, &body); err != nil {
			WriteError(w, err)
			return
		}
		report := c.Security.Evaluate(r.Context(), body.Samples, body.Threshold, body.TargetFPR)
		WriteJSON(w, http.StatusOK, report)
	}
}

func inferHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body apis.InferenceHTTPRequest
		if err := decodeJSON(r, &body); err != nil {
			WriteError(w, err)
			return
		}
		user, _ := middleware.UserFromContext(r.Context())
		req := apis.InferenceRequest{Model: body.Model, Prompt: body.Prompt, SecurityEnabled: true}
		if body.MaxTokens != nil {
			req.MaxTokens = *body.MaxTokens
		} else {
			req.MaxTokens = 256
		}
		if body.Temperature != nil {
			req.Temperature = *body.Temperature
		}
		if body.Security != nil {
			req.SecurityEnabled = body.Security.Enabled
		}
		var userID *uuid.UUID
		if user != nil {
			userID = &user.ID
		}
		resp, err := c.Inference.Infer(r.Context(), req, userID)
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, resp)
	}
}

func listInferenceHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := queryInt(r, "limit", 50)
		rows, err := c.Inferences.ListRecent(r.Context(), limit)
		if err != nil {
			WriteError(w, err)
			return
		}
		out := make([]map[string]interface{}, 0, len(rows))
		for _, row := range rows {
			out = append(out, map[string]interface{}{
				"id": row.ID.String(), "request_id": row.RequestID.String(),
				"model": row.Model, "status": row.Status, "decision": row.Decision,
				"worker_id": row.WorkerID, "end_to_end_latency_ms": row.EndToEndLatencyMs,
				"security_overhead_ms": row.SecurityOverheadMs,
				"created_at": row.CreatedAt.Format(time.RFC3339),
			})
		}
		WriteJSON(w, http.StatusOK, out)
	}
}

func registerWorkerHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body apis.WorkerResource
		if err := decodeJSON(r, &body); err != nil {
			WriteError(w, err)
			return
		}
		res, err := c.Worker.Register(r.Context(), body)
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, res)
	}
}

func heartbeatHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body apis.WorkerHeartbeat
		if err := decodeJSON(r, &body); err != nil {
			WriteError(w, err)
			return
		}
		body.WorkerID = chi.URLParam(r, "worker_id")
		res, err := c.Worker.Heartbeat(r.Context(), body)
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, res)
	}
}

func listWorkersHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := c.Worker.ListWorkers(r.Context())
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, items)
	}
}

func getWorkerHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := c.Worker.Inspect(r.Context(), chi.URLParam(r, "worker_id"))
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, res)
	}
}

func updateWorkerHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workerID := chi.URLParam(r, "worker_id")
		var body apis.UpdateWorkerRequest
		if err := decodeJSON(r, &body); err != nil {
			WriteError(w, err)
			return
		}
		res, err := c.Worker.Update(r.Context(), workerID, body)
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, res)
	}
}

func deleteWorkerHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workerID := chi.URLParam(r, "worker_id")
		if err := c.Worker.Delete(r.Context(), workerID); err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted", "worker_id": workerID})
	}
}

func pruneOfflineWorkersHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		count, err := c.Worker.PruneOffline(r.Context())
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{"status": "pruned", "count": count})
	}
}

func servingNodesHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := c.Serving.ListNodes(r.Context())
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, items)
	}
}

func servingModelsHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := c.Serving.ListModels(r.Context())
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, items)
	}
}

func servingCatalogHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, c.Serving.Catalog())
	}
}

func activateModelHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model string `json:"model"`
		}
		if err := decodeJSON(r, &body); err != nil || body.Model == "" {
			WriteError(w, spidererrors.Validation("model is required"))
			return
		}
		result, err := c.Serving.ActivateModel(r.Context(), body.Model)
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, result)
	}
}

func metricsSummaryHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summary, err := c.MetricsSvc.DashboardSummary(r.Context())
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, summary)
	}
}

func jobsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, []interface{}{})
	}
}

func queryInt(r *http.Request, key string, fallback int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
