import { cn } from "@/lib/utils";

const tones: Record<string, string> = {
  ALLOW: "bg-emerald-500/15 text-emerald-300",
  BLOCK: "bg-red-500/15 text-red-300",
  REVIEW: "bg-amber-500/15 text-amber-300",
  ERROR: "bg-zinc-500/15 text-zinc-300",
  ONLINE: "bg-emerald-500/15 text-emerald-300",
  OFFLINE: "bg-zinc-500/15 text-zinc-400",
  BUSY: "bg-amber-500/15 text-amber-300",
};

export function Badge({ value }: { value: string }) {
  return (
    <span
      className={cn(
        "inline-flex rounded px-2 py-0.5 font-mono text-xs uppercase tracking-wide",
        tones[value] ?? "bg-zinc-800 text-zinc-300",
      )}
    >
      {value}
    </span>
  );
}
