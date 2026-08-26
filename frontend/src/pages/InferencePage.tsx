import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { api } from "@/services/api";
import { formatMs } from "@/lib/utils";

export function InferencePage() {
  const [model, setModel] = useState("meta-llama/Llama-3.1-8B-Instruct");
  const [prompt, setPrompt] = useState("Explain distributed systems");
  const history = useQuery({
    queryKey: ["inference", "history"],
    queryFn: api.inferenceHistory,
    staleTime: 10_000,
  });
  const submit = useMutation({ mutationFn: () => api.inference(model, prompt) });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-medium">Protected inference</h1>
        <p className="text-sm text-zinc-400">
          Every request is inspected by SPIDER before the Mock LLM (or a real provider) is called.
        </p>
      </div>
      <Card className="space-y-3">
        <Input value={model} onChange={(e) => setModel(e.target.value)} />
        <textarea
          className="min-h-32 w-full rounded-md border border-line bg-zinc-950 p-3 font-mono text-sm"
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
        />
        <Button onClick={() => submit.mutate()} disabled={submit.isPending}>
          Submit
        </Button>
        {submit.data ? (
          <div className="space-y-2 font-mono text-sm">
            <div className="flex gap-2">
              <Badge value={submit.data.decision} />
              <span>{submit.data.status}</span>
            </div>
            <div>overhead={formatMs(submit.data.security_overhead_ms)}</div>
            <div>e2e={formatMs(submit.data.end_to_end_latency_ms)}</div>
            <pre className="whitespace-pre-wrap text-zinc-300">
              {submit.data.output ?? "(blocked — LLM not called)"}
            </pre>
          </div>
        ) : null}
      </Card>
      <Card>
        <h2 className="mb-3 text-sm uppercase text-zinc-500">Recent requests</h2>
        <ul className="space-y-2 font-mono text-sm">
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
