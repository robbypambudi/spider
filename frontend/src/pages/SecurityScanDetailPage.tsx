import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { api } from "@/services/api";
import { formatMs } from "@/lib/utils";

export function SecurityScanDetailPage() {
  const { scanId = "" } = useParams();
  const { data } = useQuery({
    queryKey: ["security", "scan", scanId],
    queryFn: () => api.scan(scanId),
    enabled: Boolean(scanId),
    staleTime: 30_000,
  });

  if (!data) return <div className="text-slate-500">Loading scan…</div>;

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold text-slate-900">Scan {data.id.slice(0, 8)}</h1>
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <div className="text-sm text-slate-500">Decision</div>
          <div className="mt-2">
            <Badge value={data.decision} />
          </div>
        </Card>
        <Card>
          <div className="text-sm text-slate-500">Score</div>
          <div className="mt-2 font-mono text-xl text-slate-900">{data.score.toFixed(3)}</div>
        </Card>
        <Card>
          <div className="text-sm text-slate-500">Threshold</div>
          <div className="mt-2 font-mono text-xl text-slate-900">{data.threshold ?? "—"}</div>
        </Card>
        <Card>
          <div className="text-sm text-slate-500">Latency</div>
          <div className="mt-2 font-mono text-xl text-slate-900">{formatMs(data.latency_ms)}</div>
        </Card>
      </div>
      <Card className="space-y-1 font-mono text-sm text-slate-700">
        <div>
          detector={data.detector}@{data.detector_version}
        </div>
        <div>policy={data.policy}</div>
        <div>chunking={data.chunking_strategy}</div>
        <div>chunks={data.chunks_scanned}</div>
        <div>prompt_hash={data.prompt_hash}</div>
        <div>prompt_length={data.prompt_length}</div>
        <div>model={data.model_target ?? "—"}</div>
      </Card>
      <Card>
        <h2 className="mb-3 text-sm font-medium uppercase tracking-wide text-slate-500">Detector executions</h2>
        <ul className="space-y-2 font-mono text-sm text-slate-700">
          {data.detectors.map((item, index) => (
            <li key={`${item.detector}-${index}`}>
              {item.detector} score={item.score.toFixed(3)} injection={String(item.is_injection)}{" "}
              {formatMs(item.latency_ms)}
            </li>
          ))}
        </ul>
      </Card>
    </div>
  );
}
