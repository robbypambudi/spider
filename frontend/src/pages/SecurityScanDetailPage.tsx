import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { ThreatScoreGauge } from "@/components/security/ThreatScoreGauge";
import { api } from "@/services/api";
import { formatMs } from "@/lib/utils";
import {
  AlertTriangle,
  ArrowLeft,
  CheckCircle2,
  Cpu,
  Quote,
  RefreshCw,
  Search,
  Shield,
} from "lucide-react";

export function SecurityScanDetailPage() {
  const { scanId = "" } = useParams();
  const [filterFlaggedOnly, setFilterFlaggedOnly] = useState(false);
  const [searchChunk, setSearchChunk] = useState("");

  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ["security", "scans", scanId],
    queryFn: () => api.scan(scanId),
    enabled: Boolean(scanId),
  });

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center text-sm text-slate-500 font-sans">
        <RefreshCw className="mr-2 h-4 w-4 animate-spin text-accent" />
        Loading scan inspection telemetry...
      </div>
    );
  }

  if (!data) {
    return (
      <div className="space-y-4">
        <Link to="/security/scans" className="inline-flex items-center gap-1.5 text-xs text-slate-500 hover:text-slate-900 font-sans">
          <ArrowLeft className="h-3.5 w-3.5" /> Back to threat log
        </Link>
        <Card className="py-12 text-center text-slate-500 font-sans">
          <Shield className="mx-auto h-8 w-8 text-slate-300 mb-2" />
          <p className="font-semibold text-slate-800">Scan result not found</p>
          <p className="text-xs text-slate-400">The requested scan ID {scanId} does not exist.</p>
        </Card>
      </div>
    );
  }

  const threshold = data.threshold ?? 0.5;

  const filteredDetectors = data.detectors.filter((item) => {
    const matchesFlag = !filterFlaggedOnly || item.is_injection || item.score >= threshold;
    const snippetText = (item.snippet || item.chunk_text || (item.metadata?.snippet as string) || (item.metadata?.chunk_text as string) || "").toLowerCase();
    const detectorName = item.detector.toLowerCase();
    const matchesSearch = !searchChunk || snippetText.includes(searchChunk.toLowerCase()) || detectorName.includes(searchChunk.toLowerCase());
    return matchesFlag && matchesSearch;
  });

  const flaggedCount = data.detectors.filter((d) => d.is_injection || d.score >= threshold).length;

  return (
    <div className="space-y-6">
      {/* Breadcrumb Navigation */}
      <div className="flex items-center justify-between font-sans">
        <Link
          to="/security/scans"
          className="inline-flex items-center gap-1.5 text-xs font-medium text-slate-500 hover:text-slate-900 transition"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Back to security threat log
        </Link>
        <Button
          variant="outline"
          size="sm"
          onClick={() => refetch()}
          icon={<RefreshCw className={`h-3.5 w-3.5 ${isFetching ? "animate-spin" : ""}`} />}
        >
          Refresh
        </Button>
      </div>

      {/* Header Overview */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Left: Gauge */}
        <ThreatScoreGauge
          score={data.score}
          threshold={threshold}
          decision={data.decision}
          className="lg:col-span-1"
        />

        {/* Right: Telemetry & Enforcement Specs */}
        <Card className="flex flex-col justify-between space-y-4 p-5 lg:col-span-2">
          <div className="flex items-start justify-between border-b border-line-subtle pb-3">
            <div>
              <div className="flex items-center gap-2 font-mono text-base font-bold text-slate-900">
                <span>Scan #{data.id.slice(0, 13)}</span>
                <Badge value={data.decision} showDot pulse />
              </div>
              <div className="text-xs text-slate-500 font-sans mt-0.5">
                Timestamp: {new Date(data.created_at).toLocaleString()}
              </div>
            </div>
            <div className="text-right font-mono">
              <div className="text-[10px] uppercase text-slate-400 font-semibold font-sans">Latency</div>
              <div className="text-sm font-bold text-slate-900">{formatMs(data.latency_ms)}</div>
            </div>
          </div>

          {/* Details Grid */}
          <div className="grid grid-cols-2 gap-3 text-xs font-mono sm:grid-cols-4">
            <div className="rounded-lg bg-slate-50 p-2.5 space-y-0.5">
              <div className="text-[10px] uppercase text-slate-400 font-semibold font-sans">Detector Engine</div>
              <div className="font-bold text-slate-900">{data.detector}</div>
            </div>

            <div className="rounded-lg bg-slate-50 p-2.5 space-y-0.5">
              <div className="text-[10px] uppercase text-slate-400 font-semibold font-sans">Active Policy</div>
              <div className="font-bold text-slate-900">{data.policy}</div>
            </div>

            <div className="rounded-lg bg-slate-50 p-2.5 space-y-0.5">
              <div className="text-[10px] uppercase text-slate-400 font-semibold font-sans">Chunks Scanned</div>
              <div className="font-bold text-slate-900">{data.chunks_scanned} chunks</div>
            </div>

            <div className="rounded-lg bg-slate-50 p-2.5 space-y-0.5">
              <div className="text-[10px] uppercase text-slate-400 font-semibold font-sans">Flagged Threat Chunks</div>
              <div className={`font-bold ${flaggedCount > 0 ? "text-rose-600" : "text-emerald-600"}`}>
                {flaggedCount} flagged
              </div>
            </div>
          </div>

          <div className="border-t border-line-subtle pt-2 text-[11px] font-mono text-slate-500 flex flex-wrap justify-between gap-2">
            <span>Prompt Hash: <span className="text-slate-700">{data.prompt_hash}</span></span>
            <span>Length: {data.prompt_length} chars</span>
          </div>
        </Card>
      </div>

      {/* Individual Detector / Chunk Results Breakdown with Snippets */}
      <Card className="space-y-4 p-5">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line-subtle pb-3">
          <div className="flex items-center gap-2">
            <Cpu className="h-4 w-4 text-indigo-500" />
            <h3 className="text-sm font-bold text-slate-900">Detector Execution & Chunk Snippet Breakdown</h3>
          </div>

          {/* Filter & Search Bar */}
          <div className="flex flex-wrap items-center gap-3">
            {/* Search Chunk Text */}
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-400" />
              <input
                type="text"
                placeholder="Search chunk text..."
                value={searchChunk}
                onChange={(e) => setSearchChunk(e.target.value)}
                className="rounded-lg border border-line bg-slate-50 pl-8 pr-2.5 py-1 text-xs text-slate-900 placeholder:text-slate-400 outline-none focus:border-accent focus:bg-white"
              />
            </div>

            {/* Flagged Only Toggle */}
            <button
              type="button"
              onClick={() => setFilterFlaggedOnly(!filterFlaggedOnly)}
              className={`flex items-center gap-1.5 rounded-lg px-2.5 py-1 text-xs font-semibold transition ${
                filterFlaggedOnly
                  ? "bg-rose-50 text-rose-700 ring-1 ring-rose-300"
                  : "bg-slate-100 text-slate-600 hover:bg-slate-200"
              }`}
            >
              <AlertTriangle className="h-3.5 w-3.5" />
              <span>Flagged Attacks Only ({flaggedCount})</span>
            </button>

            <span className="font-mono text-xs text-slate-500">
              Showing {filteredDetectors.length} of {data.detectors.length}
            </span>
          </div>
        </div>

        {/* Chunk List */}
        <div className="space-y-3">
          {filteredDetectors.map((item, index) => {
            const isChunkInjection = item.is_injection || item.score >= threshold;
            const snippet =
              item.snippet ||
              item.chunk_text ||
              (item.metadata?.snippet as string) ||
              (item.metadata?.chunk_text as string) ||
              "";
            const matchedPatterns = (item.metadata?.matched_patterns as string[]) || [];

            return (
              <div
                key={`${item.detector}-${index}`}
                className={`rounded-xl border p-4 transition space-y-3 ${
                  isChunkInjection
                    ? "border-rose-200 bg-rose-50/40 ring-1 ring-rose-200/60"
                    : "border-line-subtle bg-slate-50/50"
                }`}
              >
                {/* Chunk Header */}
                <div className="flex flex-wrap items-center justify-between gap-2 border-b border-line-subtle/80 pb-2.5">
                  <div className="flex items-center gap-2">
                    <span
                      className={`flex h-6 w-6 items-center justify-center rounded-md font-mono text-[11px] font-bold ${
                        isChunkInjection
                          ? "bg-rose-600 text-white"
                          : "bg-slate-200 text-slate-700"
                      }`}
                    >
                      {item.chunk_index !== undefined ? item.chunk_index + 1 : index + 1}
                    </span>
                    <div>
                      <div className="font-mono text-xs font-bold text-slate-900">
                        {item.detector} {item.version ? `@ ${item.version}` : ""}
                      </div>
                      <div className="text-[11px] text-slate-500 font-mono">
                        Latency: {formatMs(item.latency_ms)} · Flag: {String(item.is_injection)}
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <div className="text-right font-mono">
                      <div className="text-[10px] uppercase text-slate-400 font-sans">Threat Score</div>
                      <div
                        className={`text-sm font-bold ${
                          item.score >= 0.75
                            ? "text-rose-600"
                            : item.score >= threshold
                            ? "text-amber-600"
                            : "text-emerald-600"
                        }`}
                      >
                        {item.score.toFixed(3)}
                      </div>
                    </div>
                    <Badge value={isChunkInjection ? "BLOCK" : "ALLOW"} />
                  </div>
                </div>

                {/* Matched Pattern Warning Pills */}
                {matchedPatterns.length > 0 && (
                  <div className="flex flex-wrap items-center gap-1.5 text-xs">
                    <span className="text-[11px] font-semibold text-rose-700">Matched Injections:</span>
                    {matchedPatterns.map((pat, pi) => (
                      <span
                        key={pi}
                        className="rounded bg-rose-100 px-2 py-0.5 font-mono text-[11px] font-bold text-rose-800"
                      >
                        {pat}
                      </span>
                    ))}
                  </div>
                )}

                {/* Extracted Sentence / Chunk Text Snippet */}
                {snippet ? (
                  <div className="rounded-lg bg-white p-3 border border-slate-200 shadow-xs space-y-1">
                    <div className="flex items-center gap-1 text-[10px] uppercase font-bold tracking-wider text-slate-400 font-sans">
                      <Quote className="h-3 w-3 text-accent" />
                      Extracted Sentence / Chunk Content
                    </div>
                    <div className="font-mono text-xs text-slate-800 whitespace-pre-wrap break-words leading-relaxed">
                      "{snippet}"
                    </div>
                  </div>
                ) : (
                  <div className="text-[11px] text-slate-400 italic font-sans">
                    No raw text snippet attached to this chunk evaluation.
                  </div>
                )}
              </div>
            );
          })}

          {filteredDetectors.length === 0 && (
            <div className="py-8 text-center text-slate-400 font-sans text-xs">
              <CheckCircle2 className="mx-auto h-8 w-8 text-emerald-400 mb-2" />
              <p className="font-semibold text-slate-700">No matching chunks found</p>
              <p className="text-slate-400 mt-0.5">
                {filterFlaggedOnly
                  ? "All chunks evaluated below threat threshold (clean benign)."
                  : "No chunks match your search query."}
              </p>
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}
