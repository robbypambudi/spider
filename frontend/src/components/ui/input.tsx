import { cn } from "@/lib/utils";
import type { InputHTMLAttributes } from "react";

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={cn(
        "w-full rounded-md border border-line bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none ring-orange-500 focus:ring-1",
        className,
      )}
      {...props}
    />
  );
}
