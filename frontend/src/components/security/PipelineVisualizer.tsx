import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import type { ScanResponse } from "@/types/api";
import {
  ArrowRight,
  Bot,
  CheckCircle2,
  ChevronDown,
  ChevronUp,
  Cpu,
  Layers,
  Quote,
  Scale,
  Shield,
  ShieldAlert,
  Sliders,
} from "lucide-react";

interface PipelineVisualizerProps {
  scan: ScanResponse;
  promptText?: string;
}

export function PipelineVisualizer({ scan, promptText }: PipelineVisualizerProps) {
  const [showAllChunks, setShowAllChunks] = useState(false);
  const isBlock = scan.decision === "BLOCK";
  const threshold = scan.threshold ?? 0.5;

  const detectors = scan.detectors ?? [];
  const flaggedChunks = detectors.filter((d) => d.is_injection || d.score >= threshold);

  const steps = [
    {
      step: 1,
      title: "Input & Preprocess",
      icon: <Bot className="h-4 w-4 text-blue-500" />,
      detail: `${promptText ? promptText.length : "—"} chars`,
      desc: "Normalized Unicode & whitespace",
      status: "COMPLETED",
    },
    {
      step: 2,
      title: "Token Chunker",
      icon: <Layers className="h-4 w-4 text-indigo-500" />,
      detail: `${scan.chunks_scanned} chunks`,
      desc: "Sliding-window overlap",
      status: "COMPLETED",
    },
    {
      step: 3,
      title: "ML / Rule Detector",
      icon: <Cpu className="h-4 w-4 text-purple-500" />,
      detail: detectors?.[0]?.detector ?? (scan.model || "prompt-shield"),
      desc: `${scan.latency_ms.toFixed(1)}ms latency`,
      status: "COMPLETED",
    },
    {
      step: 4,
      title: "Aggregator",
      icon: <Sliders className="h-4 w-4 text-amber-500" />,
      detail: `Max score: ${scan.score.toFixed(3)}`,
      desc: "Peak chunk threat aggregation",
      status: "COMPLETED",
    },
    {
      step: 5,
      title: "Policy Evaluation",
      icon: <Scale className="h-4 w-4 text-teal-500" />,
      detail: `Policy: ${scan.policy || "default"}`,
      desc: `τ threshold = ${threshold.toFixed(2)}`,
      status: "COMPLETED",
    },
  ];

  const displayChunks = showAllChunks ? detectors : flaggedChunks.length > 0 ? flaggedChunks : detectors.slice(0, 3);

  return (
    <div className="space-y-4 rounded-xl border border-line bg-white p-5 shadow-card">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-line-subtle pb-3">
        <div className="flex items-center gap-2">
          <Shield className="h-5 w-5 text-accent" />
          <h3 className="text-sm font-bold text-slate-900">Runtime Defense Pipeline Execution</h3>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-slate-500">Result:</span>
          <Badge value={scan.decision} showDot pulse />
        </div>
      </div>

      {/* Steps Flowchart */}
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-5">
        {steps.map((item, idx) => (
          <div key={item.step} className="relative flex flex-col justify-between rounded-lg border border-line-subtle bg-slate-50/70 p-3">
            <div className="space-y-1.5">
              <div className="flex items-center justify-between text-xs">
                <span className="flex items-center gap-1.5 font-semibold text-slate-700">
                  {item.icon}
                  {item.title}
                </span>
                <span className="font-mono text-[10px] text-slate-400">0{item.step}</span>
              </div>
              <div className="font-mono text-xs font-bold text-slate-900">{item.detail}</div>
              <p className="text-[11px] text-slate-500">{item.desc}</p>
            </div>

            {idx < steps.length - 1 && (
              <div className="hidden lg:block absolute -right-2 top-1/2 -translate-y-1/2 z-10">
                <div className="flex h-4 w-4 items-center justify-center rounded-full bg-white text-slate-400 ring-1 ring-slate-200">
                  <ArrowRight className="h-2.5 w-2.5" />
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Enforcement Decision Banner */}
      <div
        className={`flex items-center justify-between rounded-lg p-4 font-mono text-xs ${
          isBlock
            ? "border border-rose-200 bg-rose-50 text-rose-800"
            : "border border-emerald-200 bg-emerald-50 text-emerald-800"
        }`}
      >
        <div className="flex items-center gap-2.5">
          {isBlock ? (
            <ShieldAlert className="h-5 w-5 text-rose-600 shrink-0" />
          ) : (
            <CheckCircle2 className="h-5 w-5 text-emerald-600 shrink-0" />
          )}
          <div>
            <div className="font-bold uppercase tracking-wide">
              {isBlock ? "Attack Blocked — Invariant Enforced" : "Prompt Passed Clean — Safe to Serve"}
            </div>
            <div className="text-[11px] opacity-90">
              {isBlock
                ? "Threat score exceeds threshold. LLMProvider.infer() was strictly intercepted and NOT called."
                : "No prompt injection detected. Prompt securely forwarded to the serving router & cluster."}
            </div>
          </div>
        </div>
        <div className="hidden sm:block text-right">
          <div className="text-[10px] opacity-75">Pipeline Latency</div>
          <div className="font-bold text-sm">{scan.latency_ms.toFixed(2)} ms</div>
        </div>
      </div>

      {/* Extracted Chunk Snippets & Dangerous Sentences */}
      {detectors.length > 0 && (
        <div className="space-y-3 border-t border-line-subtle pt-3 font-sans">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Quote className="h-4 w-4 text-accent" />
              <h4 className="text-xs font-bold text-slate-900 uppercase tracking-wider">
                Evaluated Chunk Snippets & Threat Attribution
              </h4>
              {flaggedChunks.length > 0 && (
                <span className="rounded-full bg-rose-100 px-2 py-0.5 text-[10px] font-bold text-rose-800">
                  {flaggedChunks.length} Threat Chunk(s)
                </span>
              )}
            </div>

            {detectors.length > 3 && (
              <button
                type="button"
                onClick={() => setShowAllChunks(!showAllChunks)}
                className="flex items-center gap-1 text-xs font-semibold text-accent hover:underline"
              >
                <span>{showAllChunks ? "Collapse" : `Show All ${detectors.length} Chunks`}</span>
                {showAllChunks ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
              </button>
            )}
          </div>

          <div className="space-y-2">
            {displayChunks.map((chunk, cIdx) => {
              const isFlagged = chunk.is_injection || chunk.score >= threshold;
              const snippet =
                chunk.snippet ||
                chunk.chunk_text ||
                (chunk.metadata?.snippet as string) ||
                (chunk.metadata?.chunk_text as string) ||
                (promptText && promptText.length <= 300 ? promptText : "");
              const matchedPatterns = (chunk.metadata?.matched_patterns as string[]) || [];

              return (
                <div
                  key={cIdx}
                  className={`rounded-lg border p-3 text-xs transition space-y-2 ${
                    isFlagged
                      ? "border-rose-200 bg-rose-50/50 ring-1 ring-rose-200/50"
                      : "border-line-subtle bg-slate-50/50"
                  }`}
                >
                  <div className="flex items-center justify-between font-mono">
                    <div className="flex items-center gap-2">
                      <span
                        className={`flex h-5 w-5 items-center justify-center rounded text-[10px] font-bold ${
                          isFlagged ? "bg-rose-600 text-white" : "bg-slate-200 text-slate-700"
                        }`}
                      >
                        {chunk.chunk_index !== undefined ? chunk.chunk_index + 1 : cIdx + 1}
                      </span>
                      <span className="font-semibold text-slate-800">{chunk.detector}</span>
                    </div>

                    <div className="flex items-center gap-2">
                      <span className="text-[11px] text-slate-500">Score:</span>
                      <span
                        className={`font-bold ${
                          chunk.score >= 0.75
                            ? "text-rose-600"
                            : chunk.score >= threshold
                            ? "text-amber-600"
                            : "text-emerald-600"
                        }`}
                      >
                        {chunk.score.toFixed(3)}
                      </span>
                      <Badge value={isFlagged ? "BLOCK" : "ALLOW"} />
                    </div>
                  </div>

                  {matchedPatterns.length > 0 && (
                    <div className="flex flex-wrap items-center gap-1 font-mono text-[10px]">
                      <span className="font-semibold text-rose-700">Matched Pattern:</span>
                      {matchedPatterns.map((pat, pi) => (
                        <span key={pi} className="rounded bg-rose-100 px-1.5 py-0.5 font-bold text-rose-800">
                          {pat}
                        </span>
                      ))}
                    </div>
                  )}

                  {snippet && (
                    <div className="rounded bg-white p-2.5 border border-slate-200/80 font-mono text-[11px] text-slate-800 leading-relaxed break-words whitespace-pre-wrap">
                      "{snippet}"
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
