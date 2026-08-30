import { cn } from "@/lib/utils";

interface ThreatScoreGaugeProps {
  score: number; // 0.0 to 1.0
  threshold?: number; // e.g. 0.5
  decision?: string; // ALLOW, BLOCK, REVIEW
  className?: string;
}

export function ThreatScoreGauge({
  score,
  threshold = 0.5,
  decision,
  className,
}: ThreatScoreGaugeProps) {
  const percentage = Math.min(Math.max(0, score * 100), 100);
  const thresholdPct = Math.min(Math.max(0, threshold * 100), 100);

  const isBlocked = decision === "BLOCK" || score >= threshold;

  return (
    <div
      className={cn(
        "rounded-xl border p-4 transition-all duration-200",
        isBlocked
          ? "border-rose-200 bg-rose-50/40"
          : "border-emerald-200 bg-emerald-50/40",
        className,
      )}
    >
      <div className="flex items-center justify-between">
        <div>
          <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
            Threat Score
          </span>
          <div className="flex items-baseline gap-2">
            <span
              className={cn(
                "font-mono text-3xl font-bold",
                isBlocked ? "text-rose-600" : "text-emerald-600",
              )}
            >
              {score.toFixed(3)}
            </span>
            <span className="font-mono text-xs text-slate-400">/ 1.000</span>
          </div>
        </div>

        <div className="text-right">
          <div className="text-xs text-slate-500">Policy Threshold (τ)</div>
          <span className="font-mono text-sm font-semibold text-slate-700">
            {threshold.toFixed(3)}
          </span>
        </div>
      </div>

      {/* Progress Bar with Threshold Marker */}
      <div className="relative mt-4">
        {/* Track */}
        <div className="relative h-3 w-full overflow-hidden rounded-full bg-slate-200/80">
          <div
            className={cn(
              "h-full transition-all duration-500 ease-out",
              isBlocked
                ? "bg-gradient-to-r from-amber-400 to-rose-600"
                : "bg-gradient-to-r from-emerald-400 to-teal-500",
            )}
            style={{ width: `${percentage}%` }}
          />
        </div>

        {/* Threshold indicator line */}
        <div
          className="absolute -top-1 bottom-0 z-10 flex flex-col items-center"
          style={{ left: `${thresholdPct}%`, transform: "translateX(-50%)" }}
        >
          <div className="h-5 w-0.5 rounded-full bg-slate-900 shadow-sm ring-2 ring-white" />
          <span className="mt-1 font-mono text-[9px] font-bold text-slate-700">τ={threshold}</span>
        </div>
      </div>

      <div className="mt-3 flex justify-between text-[11px] font-medium text-slate-500">
        <span className="text-emerald-600">0.0 (Benign)</span>
        <span className="text-rose-600">1.0 (Injection)</span>
      </div>
    </div>
  );
}
