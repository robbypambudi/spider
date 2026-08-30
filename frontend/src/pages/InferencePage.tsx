import { useEffect, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { api } from "@/services/api";
import { formatMs } from "@/lib/utils";

const textareaClass =
  "min-h-32 w-full rounded-lg border border-line bg-white p-3 font-mono text-sm text-slate-900 outline-none transition focus:border-accent focus:ring-2 focus:ring-accent/20";

const DEFAULT_MODEL = "robbypambudi/prompt-shield-flan-t5-small";

export function InferencePage() {
  const catalog = useQuery({
    queryKey: ["serving", "catalog"],
    queryFn: api.servingCatalog,
    staleTime: 30_000,
  });
  const activeModel = catalog.data?.find((m) => m.active)?.id ?? DEFAULT_MODEL;
  const [model, setModel] = useState(DEFAULT_MODEL);
  const [prompt, setPrompt] = useState("Explain distributed systems");

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
  const submit = useMutation({ mutationFn: () => api.inference(model, prompt) });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-slate-900">Protected inference</h1>
        <p className="text-sm text-slate-500">
          Every request is scanned by Prompt-Shield before classification output is returned.
        </p>
      </div>
      <Card className="space-y-3">
        <label className="block text-sm font-medium text-slate-700">Prompt-Shield model</label>
        <select
          className="w-full rounded-lg border border-line bg-white p-2 font-mono text-sm text-slate-900"
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
        <textarea className={textareaClass} value={prompt} onChange={(e) => setPrompt(e.target.value)} />
        <Button onClick={() => submit.mutate()} disabled={submit.isPending}>
          Submit
        </Button>
        {submit.data ? (
          <div className="space-y-2 rounded-lg bg-slate-50 p-3 font-mono text-sm text-slate-700">
            <div className="flex gap-2">
              <Badge value={submit.data.decision} />
              <span>{submit.data.status}</span>
            </div>
            <div>overhead={formatMs(submit.data.security_overhead_ms)}</div>
            <div>e2e={formatMs(submit.data.end_to_end_latency_ms)}</div>
            <pre className="whitespace-pre-wrap text-slate-800">
              {submit.data.output ?? "(blocked — model not called)"}
            </pre>
          </div>
        ) : null}
      </Card>
      <Card>
        <h2 className="mb-3 text-sm font-medium uppercase tracking-wide text-slate-500">Recent requests</h2>
        <ul className="space-y-2 font-mono text-sm text-slate-700">
          {(history.data ?? []).map((row) => (
            <li key={String(row.request_id)}>
              {String(row.decision)} · {String(row.status)} · {String(row.model)}
            </li>
          ))}
        </ul>
      </Card>
    </div>
  );
}
