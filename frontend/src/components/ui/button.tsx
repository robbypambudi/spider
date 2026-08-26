import { cn } from "@/lib/utils";
import type { ButtonHTMLAttributes } from "react";

export function Button({ className, ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      className={cn(
        "inline-flex items-center justify-center rounded-md bg-orange-500 px-3 py-2 text-sm font-medium text-black transition hover:bg-orange-400 disabled:cursor-not-allowed disabled:opacity-50",
        className,
      )}
      {...props}
    />
  );
}
