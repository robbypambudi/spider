import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Tabs } from "@/components/ui/tabs";
import { api } from "@/services/api";
import { formatMs } from "@/lib/utils";
import {
  ArrowUpRight,
  Database,
  RefreshCw,
  Search,
  ShieldAlert,
  ShieldCheck,
} from "lucide-react";

export function SecurityScansPage() {
  const [search, setSearch] = useState("");
  const [decisionFilter, setDecisionFilter] = useState("ALL");

  const { data, refetch, isFetching } = useQuery({
    queryKey: ["security", "scans"],
    queryFn: api.scans,
    staleTime: 10_000,
  });

  const scans = data ?? [];

  const filteredScans = scans.filter((s) => {
    const matchesSearch =
      s.id.toLowerCase().includes(search.toLowerCase()) ||
      (s.policy && s.policy.toLowerCase().includes(search.toLowerCase())) ||
      (s.detector && s.detector.toLowerCase().includes(search.toLowerCase()));

    const matchesDecision = decisionFilter === "ALL" || s.decision === decisionFilter;

    return matchesSearch && matchesDecision;
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Security Threat Log</h1>
            <span className="inline-flex items-center gap-1 rounded-full bg-blue-50 px-2.5 py-0.5 text-xs font-semibold text-accent ring-1 ring-blue-200">
              <Database className="h-3 w-3" />
              Storage Optimized (Threats Only)
            </span>
          </div>
          <p className="text-sm text-slate-500 mt-0.5">
            Forensic audit trail of detected prompt injections, threat chunk scores, and enforcement actions.
          </p>
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={() => refetch()}
          icon={<RefreshCw className={`h-3.5 w-3.5 ${isFetching ? "animate-spin" : ""}`} />}
        >
          Refresh
        </Button>
      </div>

      {/* Storage Optimization Notice Banner */}
      <div className="flex items-center gap-3 rounded-xl border border-blue-100 bg-blue-50/60 px-4 py-3 text-xs text-blue-900">
        <ShieldAlert className="h-4 w-4 text-accent shrink-0" />
        <div>
          <span className="font-semibold">Selective Audit Active:</span> Benign queries (<code className="font-mono bg-blue-100/80 px-1 py-0.5 rounded text-blue-800">ALLOW</code>) are processed with zero database pollution and aggregated into telemetry metrics, while dangerous attacks (<code className="font-mono bg-blue-100/80 px-1 py-0.5 rounded text-blue-800">BLOCK</code> / <code className="font-mono bg-blue-100/80 px-1 py-0.5 rounded text-blue-800">REVIEW</code>) are preserved for forensic review.
        </div>
      </div>

      {/* Filter and Search Bar */}
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-line bg-white p-3 shadow-card">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-400" />
          <input
            type="text"
            placeholder="Search by scan ID, policy, detector..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-lg border border-line bg-slate-50/50 pl-8 pr-3 py-1.5 text-xs text-slate-900 placeholder:text-slate-400 outline-none focus:border-accent focus:bg-white focus:ring-1 focus:ring-accent"
          />
        </div>

        <Tabs
          tabs={[
            { id: "ALL", label: "All Threats", count: scans.length },
            { id: "BLOCK", label: "Blocked", count: scans.filter((s) => s.decision === "BLOCK").length },
            { id: "REVIEW", label: "Review", count: scans.filter((s) => s.decision === "REVIEW").length },
            { id: "ALLOW", label: "Allowed", count: scans.filter((s) => s.decision === "ALLOW").length },
          ]}
          activeTab={decisionFilter}
          onChange={setDecisionFilter}
        />
      </div>

      {/* Table */}
      <Card className="overflow-x-auto p-0">
        <table className="w-full text-left text-xs text-slate-700">
          <thead className="border-b border-line bg-slate-50 text-[11px] uppercase tracking-wider text-slate-500 font-semibold">
            <tr>
              <th className="px-4 py-3">Timestamp</th>
              <th className="px-4 py-3">Decision</th>
              <th className="px-4 py-3">Threat Score</th>
              <th className="px-4 py-3">Detector</th>
              <th className="px-4 py-3">Policy</th>
              <th className="px-4 py-3">Chunks</th>
              <th className="px-4 py-3">Latency</th>
              <th className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-line-subtle font-mono">
            {filteredScans.map((scan) => (
              <tr key={scan.id} className="transition hover:bg-slate-50/70">
                <td className="px-4 py-3 text-slate-500 font-sans">
                  {new Date(scan.created_at).toLocaleString()}
                </td>
                <td className="px-4 py-3">
                  <Badge value={scan.decision} />
                </td>
                <td className="px-4 py-3 font-bold text-slate-900">
                  <span
                    className={
                      scan.score >= 0.75
                        ? "text-rose-600"
                        : scan.score >= 0.5
                        ? "text-amber-600"
                        : "text-emerald-600"
                    }
                  >
                    {scan.score.toFixed(4)}
                  </span>
                </td>
                <td className="px-4 py-3 text-slate-600 font-sans">{scan.detector}</td>
                <td className="px-4 py-3 text-slate-600 font-sans">{scan.policy}</td>
                <td className="px-4 py-3 text-slate-600">{scan.chunks_scanned}</td>
                <td className="px-4 py-3 text-slate-600">{formatMs(scan.latency_ms)}</td>
                <td className="px-4 py-3 text-right">
                  <Link
                    to={`/security/scans/${scan.id}`}
                    className="inline-flex items-center gap-1 font-sans text-xs font-semibold text-accent hover:underline"
                  >
                    Inspect
                    <ArrowUpRight className="h-3.5 w-3.5" />
                  </Link>
                </td>
              </tr>
            ))}
            {filteredScans.length === 0 && (
              <tr>
                <td colSpan={8} className="py-12 text-center text-slate-400 font-sans">
                  <ShieldCheck className="mx-auto h-8 w-8 text-emerald-400 mb-2" />
                  <p className="font-semibold text-slate-700">No threat incidents found</p>
                  <p className="text-xs text-slate-400 mt-1">
                    Clean benign prompts are automatically allowed without consuming database storage.
                  </p>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
