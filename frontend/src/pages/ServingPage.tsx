import { useQuery } from "@tanstack/react-query";
import { Card } from "@/components/ui/card";
import { api } from "@/services/api";

export function ServingPage() {
  const nodes = useQuery({ queryKey: ["serving", "nodes"], queryFn: api.servingNodes, staleTime: 15_000 });
  const models = useQuery({ queryKey: ["serving", "models"], queryFn: api.servingModels, staleTime: 15_000 });

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-medium">Serving</h1>
      <p className="text-sm text-zinc-400">
        Cluster serving nodes exist so SPIDER can evaluate defenses under realistic LLM-serving load.
      </p>
      <Card>
        <h2 className="mb-3 text-sm uppercase text-zinc-500">Nodes</h2>
        <ul className="space-y-2 font-mono text-sm">
          {(nodes.data ?? []).map((node) => (
            <li key={String(node.worker_id)}>
              {String(node.worker_id)} · {String(node.status)}
            </li>
          ))}
          {(nodes.data ?? []).length === 0 ? <li className="text-zinc-500">No serving nodes registered.</li> : null}
        </ul>
      </Card>
      <Card>
        <h2 className="mb-3 text-sm uppercase text-zinc-500">Loaded models</h2>
        <ul className="space-y-2 font-mono text-sm">
          {(models.data ?? []).map((model, index) => (
            <li key={`${model.worker_id}-${index}`}>
              {String(model.name)} @ {String(model.worker_id)} · {String(model.status)}
            </li>
          ))}
        </ul>
      </Card>
    </div>
  );
}
