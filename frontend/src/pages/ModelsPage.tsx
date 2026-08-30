import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { api } from "@/services/api";

export function ModelsPage() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ["serving", "models"],
    queryFn: api.servingModels,
    staleTime: 10_000,
  });

  const activate = useMutation({
    mutationFn: (model: string) => api.activateModel(model),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["serving"] });
      queryClient.invalidateQueries({ queryKey: ["workers"] });
    },
  });

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-semibold text-slate-900">Prompt-Shield models</h1>
        <p className="text-sm text-slate-500">
          Fine-tuned Flan-T5 classifiers from the{" "}
          <a
            className="text-accent hover:underline"
            href="https://huggingface.co/collections/robbypambudi/prompt-shield"
            rel="noreferrer"
            target="_blank"
          >
            Prompt-Shield collection
          </a>
          . Activate one model for runtime detection and inference.
        </p>
      </div>
      <Card className="divide-y divide-line p-0">
        {(data ?? []).map((model) => {
          const id = String(model.id);
          const active = Boolean(model.active);
          const status = String(model.status ?? "AVAILABLE");
          return (
            <div key={id} className="flex flex-wrap items-center justify-between gap-3 px-4 py-4">
              <div className="space-y-1">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-medium text-slate-900">{String(model.name)}</span>
                  {active ? <Badge value="ACTIVE" /> : null}
                  <Badge value={status} />
                </div>
                <p className="font-mono text-xs text-slate-600">{id}</p>
                <p className="text-sm text-slate-500">
                  {String(model.params)} params · {String(model.description ?? "")}
                </p>
              </div>
              <Button disabled={active || activate.isPending} onClick={() => activate.mutate(id)}>
                {active ? "Active" : "Activate"}
              </Button>
            </div>
          );
        })}
        {!isLoading && (data ?? []).length === 0 ? (
          <p className="px-4 py-6 text-sm text-slate-500">
            No models in catalog. Ensure prompt-shield sidecar and spider-worker are running.
          </p>
        ) : null}
      </Card>
    </div>
  );
}
