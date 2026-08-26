import { useQuery } from "@tanstack/react-query";
import { Card } from "@/components/ui/card";
import { api } from "@/services/api";
import { formatMs, formatPct } from "@/lib/utils";

export function DashboardPage() {
  const { data } = useQuery({
    queryKey: ["metrics", "summary"],
    queryFn: api.summary,
    staleTime: 15_000,
  });

  const cards = [
    ["Total scans", data?.total_scans ?? 0],
    ["Allowed", data?.allowed ?? 0],
    ["Blocked", data?.blocked ?? 0],
    ["Review", data?.review ?? 0],
    ["Detection rate", formatPct(data?.detection_rate)],
    ["Avg detection latency", formatMs(data?.avg_detection_latency_ms)],
    ["P95 detection latency", formatMs(data?.p95_detection_latency_ms)],
    ["Avg security overhead", formatMs(data?.avg_security_overhead_ms)],
    ["Active serving nodes", data?.active_serving_nodes ?? 0],
    ["Total GPUs", data?.total_gpus ?? 0],
  ] as const;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-medium">Security overview</h1>
        <p className="text-sm text-zinc-400">
          Prompt injection defense telemetry for the SPIDER control plane.
        </p>
      </div>
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
        {cards.map(([label, value]) => (
          <Card key={label}>
            <div className="text-xs uppercase tracking-wide text-zinc-500">{label}</div>
            <div className="mt-2 font-mono text-2xl">{value}</div>
          </Card>
        ))}
      </div>
    </div>
  );
}
