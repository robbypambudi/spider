import { useQuery } from "@tanstack/react-query";
import { Card } from "@/components/ui/card";
import { api } from "@/services/api";

export function ModelsPage() {
  const { data } = useQuery({
    queryKey: ["serving", "models"],
    queryFn: api.servingModels,
    staleTime: 15_000,
  });

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-medium">Models</h1>
      <Card>
        <ul className="space-y-2 font-mono text-sm">
          {(data ?? []).map((model, index) => (
            <li key={`${model.name}-${index}`}>
              {String(model.name)} · {String(model.status)} · worker={String(model.worker_id)}
            </li>
          ))}
          {(data ?? []).length === 0 ? (
            <li className="text-zinc-500">Start spider-worker to report loaded models (default: mock-llm).</li>
          ) : null}
        </ul>
      </Card>
    </div>
  );
}
