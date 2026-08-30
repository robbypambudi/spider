import { cn } from "@/lib/utils";

const tones: Record<string, string> = {
  ALLOW: "bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200",
  BLOCK: "bg-red-50 text-red-700 ring-1 ring-red-200",
  REVIEW: "bg-amber-50 text-amber-700 ring-1 ring-amber-200",
  ERROR: "bg-slate-100 text-slate-700 ring-1 ring-slate-200",
  ONLINE: "bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200",
  OFFLINE: "bg-slate-100 text-slate-600 ring-1 ring-slate-200",
  BUSY: "bg-amber-50 text-amber-700 ring-1 ring-amber-200",
};

export function Badge({ value }: { value: string }) {
  return (
    <span
      className={cn(
        "inline-flex rounded-md px-2 py-0.5 font-mono text-xs uppercase tracking-wide",
        tones[value] ?? "bg-slate-100 text-slate-700 ring-1 ring-slate-200",
      )}
    >
      {value}
    </span>
  );
}
