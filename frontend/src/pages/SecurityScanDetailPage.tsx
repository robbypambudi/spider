import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { ThreatScoreGauge } from "@/components/security/ThreatScoreGauge";
import { api } from "@/services/api";
import { formatMs } from "@/lib/utils";
import {
  ArrowLeft,
  Cpu,
  Fingerprint,
  Shield,
  ShieldAlert,
} from "lucide-react";

export function SecurityScanDetailPage() {
  const { scanId = "" } = useParams();
  const { data, isLoading } = useQuery({
    queryKey: ["security", "scan", scanId],
    queryFn: () => api.scan(scanId),
    enabled: Boolean(scanId),
    staleTime: 30_000,
  });

  if (isLoading) {
    return <div className="py-12 text-center text-sm text-slate-500">Loading scan audit log...</div>;
  }

  if (!data) {
    return (
      <div className="space-y-4">
        <Link to="/security/scans" className="inline-flex items-center gap-1.5 text-xs text-slate-500 hover:text-slate-900">
          <ArrowLeft className="h-3.5 w-3.5" /> Back to scans
        </Link>
        <Card className="py-12 text-center text-slate-500">Scan ID not found.</Card>
      </div>
    );
  }

  const isBlocked = data.decision === "BLOCK";

  return (
    <div className="space-y-6">
      {/* Navigation */}
      <Link
        to="/security/scans"
        className="inline-flex items-center gap-1.5 text-xs font-medium text-slate-500 hover:text-slate-900 transition"
      >
        <ArrowLeft className="h-3.5 w-3.5" />
        Back to scan history
      </Link>

      {/* Header Banner */}
      <Card className="flex flex-wrap items-center justify-between gap-4 p-6">
        <div className="flex items-center gap-4">
          <div
            className={`flex h-12 w-12 items-center justify-center rounded-xl ring-1 ${
              isBlocked ? "bg-rose-50 text-rose-600 ring-rose-200" : "bg-emerald-50 text-emerald-600 ring-emerald-200"
            }`}
          >
            {isBlocked ? <ShieldAlert className="h-6 w-6" /> : <Shield className="h-6 w-6" />}
          </div>
          <div>
            <div className="flex items-center gap-2.5">
              <h1 className="font-mono text-xl font-bold text-slate-900">Scan {data.id}</h1>
              <Badge value={data.decision} showDot pulse />
            </div>
            <div className="mt-1 text-xs text-slate-500 font-mono">
              Recorded at: {new Date(data.created_at).toLocaleString()}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-4 border-l border-line-subtle pl-4 font-mono text-xs">
          <div>
            <div className="text-[10px] uppercase text-slate-400 font-semibold">Total Latency</div>
            <div className="text-base font-bold text-slate-900">{formatMs(data.latency_ms)}</div>
          </div>
        </div>
      </Card>

      {/* Score and Metadata */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <ThreatScoreGauge
          score={data.score}
          threshold={data.threshold ?? 0.5}
          decision={data.decision}
          className="lg:col-span-1"
        />

        {/* Pipeline Parameters */}
        <Card className="space-y-3 p-5 lg:col-span-2">
          <div className="flex items-center gap-2 border-b border-line-subtle pb-2.5">
            <Fingerprint className="h-4 w-4 text-accent" />
            <h3 className="text-sm font-bold text-slate-900">Pipeline Execution Metadata</h3>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 font-mono text-xs">
            <div className="rounded-lg bg-slate-50 p-2.5 space-y-0.5">
              <div className="text-[10px] uppercase text-slate-400 font-semibold font-sans">Detector Engine</div>
              <div className="font-bold text-slate-900">{data.detector} @ {data.detector_version}</div>
            </div>

            <div className="rounded-lg bg-slate-50 p-2.5 space-y-0.5">
              <div className="text-[10px] uppercase text-slate-400 font-semibold font-sans">Active Policy</div>
              <div className="font-bold text-slate-900">{data.policy}</div>
            </div>

            <div className="rounded-lg bg-slate-50 p-2.5 space-y-0.5">
              <div className="text-[10px] uppercase text-slate-400 font-semibold font-sans">Chunking Strategy</div>
              <div className="font-bold text-slate-900">{data.chunking_strategy} ({data.chunks_scanned} chunks)</div>
            </div>

            <div className="rounded-lg bg-slate-50 p-2.5 space-y-0.5">
              <div className="text-[10px] uppercase text-slate-400 font-semibold font-sans">Target Model</div>
              <div className="font-bold text-slate-900">{data.model_target ?? "None (Scan Only)"}</div>
            </div>
          </div>

          <div className="border-t border-line-subtle pt-2 text-[11px] font-mono text-slate-500">
            Prompt Hash: <span className="text-slate-700">{data.prompt_hash}</span> (Length: {data.prompt_length} chars)
          </div>
        </Card>
      </div>

      {/* Individual Detector Results */}
      <Card className="space-y-4 p-5">
        <div className="flex items-center justify-between border-b border-line-subtle pb-3">
          <div className="flex items-center gap-2">
            <Cpu className="h-4 w-4 text-indigo-500" />
            <h3 className="text-sm font-bold text-slate-900">Detector Execution Breakdown</h3>
          </div>
          <span className="font-mono text-xs text-slate-500">{data.detectors.length} detector(s) executed</span>
        </div>

        <div className="space-y-2.5">
          {data.detectors.map((item, index) => (
            <div
              key={`${item.detector}-${index}`}
              className="flex items-center justify-between rounded-lg border border-line-subtle bg-slate-50/60 p-3.5 font-mono text-xs"
            >
              <div className="space-y-0.5">
                <div className="font-bold text-slate-900">{item.detector}</div>
                <div className="text-[11px] text-slate-500">
                  Latency: {formatMs(item.latency_ms)} · Injection Flag: {String(item.is_injection)}
                </div>
              </div>

              <div className="flex items-center gap-3">
                <div className="text-right">
                  <div className="text-[10px] uppercase text-slate-400 font-sans">Score</div>
                  <div className="font-bold text-slate-900">{item.score.toFixed(3)}</div>
                </div>
                <Badge value={item.is_injection ? "BLOCK" : "ALLOW"} />
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
