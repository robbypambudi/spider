import type {
  InferenceResponse,
  MetricsSummary,
  ScanDetail,
  ScanListItem,
  ScanResponse,
  WorkerView,
} from "@/types/api";

const TOKEN_KEY = "spider.token";

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

const baseUrl = import.meta.env.VITE_API_URL ?? "";

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Content-Type", "application/json");
  const token = getToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);
  const response = await fetch(`${baseUrl}${path}`, { ...init, headers });
  if (response.status === 401) {
    clearToken();
    if (!path.includes("/auth/login")) {
      window.location.href = "/login";
    }
  }
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Request failed (${response.status})`);
  }
  return (await response.json()) as T;
}

export const api = {
  login: (email: string, password: string) =>
    request<{ access_token: string; role: string; email: string }>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  me: () => request<{ id: string; email: string; role: string }>("/api/v1/auth/me"),
  summary: () => request<MetricsSummary>("/api/v1/metrics/summary"),
  scans: () => request<ScanListItem[]>("/api/v1/security/scans"),
  scan: (id: string) => request<ScanDetail>(`/api/v1/security/scans/${id}`),
  inspect: (text: string) =>
    request<ScanResponse>("/api/v1/security/scan", {
      method: "POST",
      body: JSON.stringify({ text }),
    }),
  detectors: () =>
    request<Array<{ name: string; status: string; warning: string }>>("/api/v1/security/detectors"),
  policies: () => request<Array<Record<string, unknown>>>("/api/v1/security/policies"),
  inference: (model: string, prompt: string) =>
    request<InferenceResponse>("/api/v1/inference", {
      method: "POST",
      body: JSON.stringify({ model, prompt, security: { enabled: true } }),
    }),
  inferenceHistory: () => request<Array<Record<string, unknown>>>("/api/v1/inference"),
  workers: () => request<WorkerView[]>("/api/v1/workers"),
  worker: (id: string) => request<WorkerView>(`/api/v1/workers/${id}`),
  servingNodes: () => request<Array<Record<string, unknown>>>("/api/v1/serving/nodes"),
  servingModels: () => request<Array<Record<string, unknown>>>("/api/v1/serving/models"),
};
