import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { api } from "@/services/api";
import { Boxes, Check, ExternalLink, HardDrive } from "lucide-react";

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
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Prompt-Shield Models</h1>
          <p className="text-sm text-slate-500">
            Fine-tuned Flan-T5 sequence classifiers from the Prompt-Shield collection.
          </p>
        </div>

        <a
          href="https://huggingface.co/collections/robbypambudi/prompt-shield"
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center gap-1.5 rounded-lg border border-line bg-white px-3 py-1.5 text-xs font-medium text-slate-700 shadow-xs hover:bg-slate-50 hover:text-slate-900"
        >
          <ExternalLink className="h-3.5 w-3.5" />
          Hugging Face Collection
        </a>
      </div>

      {/* Model Catalog Grid */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        {(data ?? []).map((model) => {
          const id = String(model.id);
          const active = Boolean(model.active);
          const status = String(model.status ?? "AVAILABLE");

          return (
            <Card
              key={id}
              className={`flex flex-col justify-between p-5 space-y-4 transition ${
                active ? "ring-2 ring-accent border-accent bg-blue-50/20" : ""
              }`}
            >
              <div className="space-y-2">
                <div className="flex items-start justify-between gap-2">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <Boxes className="h-4 w-4 text-accent" />
                      <h3 className="font-semibold text-slate-900 text-sm">{String(model.name)}</h3>
                    </div>
                    <div className="font-mono text-xs text-slate-500">{id}</div>
                  </div>

                  <div className="flex items-center gap-1.5">
                    {active && <Badge value="ACTIVE" showDot pulse />}
                    <Badge value={status} />
                  </div>
                </div>

                <p className="text-xs text-slate-600">
                  {String(model.description || "Sequence classification model for prompt injection detection.")}
                </p>
              </div>

              <div className="flex items-center justify-between border-t border-line-subtle pt-3">
                <div className="flex items-center gap-2 font-mono text-xs text-slate-500">
                  <HardDrive className="h-3.5 w-3.5 text-slate-400" />
                  <span>{String(model.params)} params</span>
                </div>

                <Button
                  size="sm"
                  variant={active ? "outline" : "primary"}
                  disabled={active || activate.isPending}
                  onClick={() => activate.mutate(id)}
                  icon={active ? <Check className="h-3.5 w-3.5 text-emerald-600" /> : undefined}
                >
                  {active ? "Active Model" : activate.isPending ? "Activating..." : "Activate Model"}
                </Button>
              </div>
            </Card>
          );
        })}

        {!isLoading && (data ?? []).length === 0 && (
          <div className="col-span-full py-12 text-center text-xs text-slate-400">
            No models found in catalog. Ensure the Prompt-Shield sidecar is running.
          </div>
        )}
      </div>
    </div>
  );
}
