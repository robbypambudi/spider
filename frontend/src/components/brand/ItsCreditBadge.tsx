import { cn } from "@/lib/utils";
import { GraduationCap } from "lucide-react";

interface ItsCreditBadgeProps {
  variant?: "compact" | "full" | "hero";
  className?: string;
}

export function ItsCreditBadge({ variant = "compact", className }: ItsCreditBadgeProps) {
  if (variant === "hero") {
    return (
      <div
        className={cn(
          "inline-flex items-center gap-3 rounded-xl border border-blue-200/40 bg-blue-950/40 p-3.5 backdrop-blur-md text-slate-200",
          className,
        )}
      >
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-blue-600/30 text-blue-400 ring-1 ring-blue-400/40 shadow-inner">
          <GraduationCap className="h-5 w-5" />
        </div>
        <div className="text-left">
          <div className="text-[11px] font-semibold tracking-wider text-blue-300 uppercase">
            Research & Development Collaboration
          </div>
          <div className="text-xs font-bold text-white">
            Institut Teknologi Sepuluh Nopember (ITS) Surabaya
          </div>
        </div>
      </div>
    );
  }

  if (variant === "full") {
    return (
      <div
        className={cn(
          "flex items-center gap-2.5 rounded-lg border border-line bg-slate-50 p-2.5 text-xs text-slate-600",
          className,
        )}
      >
        <GraduationCap className="h-4 w-4 text-accent shrink-0" />
        <div>
          <span className="font-semibold text-slate-800">In Partnership with:</span>{" "}
          <span className="font-medium">Institut Teknologi Sepuluh Nopember (ITS) Surabaya</span>
        </div>
      </div>
    );
  }

  return (
    <div
      className={cn(
        "inline-flex items-center gap-1.5 rounded-md bg-slate-50 px-2 py-1 text-[11px] font-medium text-slate-500 ring-1 ring-slate-200/70",
        className,
      )}
    >
      <GraduationCap className="h-3.5 w-3.5 text-accent" />
      <span>ITS Surabaya Collaboration</span>
    </div>
  );
}
