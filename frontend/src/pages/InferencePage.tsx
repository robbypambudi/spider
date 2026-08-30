import { useEffect, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { api } from "@/services/api";
import { formatMs } from "@/lib/utils";
import {
  Bot,
  Clock,
  Send,
} from "lucide-react";

const DEFAULT_MODEL = "robbypambudi/prompt-shield-flan-t5-small";

export function InferencePage() {
  const catalog = useQuery({
    queryKey: ["serving", "catalog"],
    queryFn: api.servingCatalog,
    staleTime: 30_000,
  });

  const activeModel = catalog.data?.find((m) => m.active)?.id ?? DEFAULT_MODEL;
  const [model, setModel] = useState(DEFAULT_MODEL);
  const [prompt, setPrompt] = useState("Explain how distributed consensus works in distributed databases.");

  useEffect(() => {
    if (catalog.data?.length) {
      setModel(activeModel);
    }
  }, [activeModel, catalog.data]);

  const history = useQuery({
    queryKey: ["inference", "history"],
    queryFn: api.inferenceHistory,
    staleTime: 10_000,
  });

  const submit = useMutation({
    mutationFn: () => api.inference(model, prompt),
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Protected LLM Inference</h1>
        <p className="text-sm text-slate-500">
          Requests are guarded by the SPIDER defense pipeline. Injections are blocked before reaching compute workers.
        </p>
      </div>

      {/* Main Request Form */}
      <Card className="space-y-4 p-6">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line-subtle pb-3">
          <div className="flex items-center gap-2">
            <Bot className="h-5 w-5 text-accent" />
            <h2 className="text-sm font-bold text-slate-900">Inference Request Sandbox</h2>
          </div>

          {/* Model selection */}
          <div className="flex items-center gap-2 text-xs">
            <span className="text-slate-500 font-medium">Model:</span>
            <select
              className="rounded-lg border border-line bg-white px-3 py-1.5 font-mono text-xs text-slate-900 shadow-xs outline-none focus:border-accent"
              value={model}
              onChange={(e) => setModel(e.target.value)}
            >
              {(catalog.data ?? [{ id: DEFAULT_MODEL, name: "Prompt-Shield Flan-T5 Small", params: "60.8M" }]).map(
                (entry) => (
                  <option key={entry.id} value={entry.id}>
                    {entry.name} ({entry.params}){entry.active ? " · active" : ""}
                  </option>
                ),
              )}
            </select>
          </div>
        </div>

        {/* Prompt Input */}
        <div className="space-y-2">
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-500">
            Prompt Content
          </label>
          <textarea
            rows={4}
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="Enter prompt for protected inference..."
            className="w-full rounded-xl border border-line bg-slate-50/50 p-4 font-mono text-sm text-slate-900 placeholder:text-slate-400 outline-none transition focus:border-accent focus:bg-white focus:ring-2 focus:ring-accent/20"
          />
          <div className="flex items-center justify-between text-xs text-slate-400">
            <span>{prompt.length} characters</span>
            <Button
              onClick={() => submit.mutate()}
              disabled={submit.isPending || !prompt.trim()}
              size="md"
              icon={<Send className="h-3.5 w-3.5" />}
            >
              {submit.isPending ? "Evaluating & Serving..." : "Send Protected Request"}
            </Button>
          </div>
        </div>

        {/* Execution Response */}
        {submit.data && (
          <div className="mt-6 space-y-4 pt-4 border-t border-line-subtle">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <span className="text-xs font-semibold text-slate-500">Defense Verdict:</span>
                <Badge value={submit.data.decision} showDot pulse />
                <span className="font-mono text-xs text-slate-500">Status: {submit.data.status}</span>
              </div>

              {/* Latency Breakdown Badges */}
              <div className="flex items-center gap-2 text-xs font-mono">
                <span className="rounded bg-teal-50 px-2 py-0.5 text-teal-700 ring-1 ring-teal-200">
                  Sec Overhead: {formatMs(submit.data.security_overhead_ms)}
                </span>
                <span className="rounded bg-blue-50 px-2 py-0.5 text-blue-700 ring-1 ring-blue-200">
                  E2E Latency: {formatMs(submit.data.end_to_end_latency_ms)}
                </span>
              </div>
            </div>

            {/* Model Response Box */}
            <div
              className={`rounded-xl border p-4 font-mono text-sm ${
                submit.data.decision === "BLOCK"
                  ? "border-rose-200 bg-rose-50/50 text-rose-900"
                  : "border-slate-200 bg-slate-50 text-slate-800"
              }`}
            >
              <div className="mb-2 text-[11px] font-semibold uppercase tracking-wider text-slate-400">
                {submit.data.decision === "BLOCK" ? "Enforcement Intercept" : "Model Output"}
              </div>
              <pre className="whitespace-pre-wrap font-mono text-xs">
                {submit.data.output ?? "Blocked by SPIDER runtime defense. Model inference was NOT executed."}
              </pre>
            </div>
          </div>
        )}
      </Card>

      {/* Recent Inference History */}
      <Card className="space-y-4 p-5">
        <div className="flex items-center justify-between border-b border-line-subtle pb-3">
          <div className="flex items-center gap-2">
            <Clock className="h-4 w-4 text-slate-500" />
            <h3 className="text-sm font-bold text-slate-900">Recent Protected Inference Runs</h3>
          </div>
          <span className="font-mono text-xs text-slate-500">{(history.data ?? []).length} recent</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="border-b border-line bg-slate-50 text-[11px] uppercase tracking-wider text-slate-500 font-semibold">
              <tr>
                <th className="px-3 py-2.5">Request ID</th>
                <th className="px-3 py-2.5">Decision</th>
                <th className="px-3 py-2.5">Status</th>
                <th className="px-3 py-2.5">Model</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line-subtle font-mono">
              {(history.data ?? []).map((row) => (
                <tr key={String(row.request_id)} className="hover:bg-slate-50/60 transition">
                  <td className="px-3 py-2.5 text-slate-700 font-bold">
                    {String(row.request_id).slice(0, 8)}...
                  </td>
                  <td className="px-3 py-2.5">
                    <Badge value={String(row.decision)} showDot />
                  </td>
                  <td className="px-3 py-2.5 text-slate-600">{String(row.status)}</td>
                  <td className="px-3 py-2.5 text-slate-500">{String(row.model)}</td>
                </tr>
              ))}
              {(history.data ?? []).length === 0 && (
                <tr>
                  <td colSpan={4} className="py-6 text-center text-slate-400 font-sans">
                    No recent inference requests recorded.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}
