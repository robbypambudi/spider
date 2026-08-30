import { Badge } from "@/components/ui/badge";
import type { ScanResponse } from "@/types/api";
import {
  ArrowRight,
  Bot,
  CheckCircle2,
  Cpu,
  Layers,
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
  const isBlock = scan.decision === "BLOCK";
  const threshold = scan.threshold ?? 0.5;

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
      detail: scan.detectors?.[0]?.detector ?? (scan.model || "prompt-shield"),
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

  return (
    <div className="space-y-4 rounded-xl border border-line bg-white p-5 shadow-card">
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
    </div>
  );
}
