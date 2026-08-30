import { cn } from "@/lib/utils";

interface SpiderLogoProps {
  size?: number;
  variant?: "icon" | "full" | "monochrome";
  glow?: boolean;
  className?: string;
}

export function SpiderLogo({
  size = 32,
  variant = "icon",
  glow = false,
  className,
}: SpiderLogoProps) {
  return (
    <div className={cn("inline-flex items-center gap-3 select-none", className)}>
      <div
        className={cn(
          "relative flex items-center justify-center rounded-xl transition-all duration-300",
          glow && "shadow-[0_0_20px_-3px_rgba(37,99,235,0.4)]",
        )}
        style={{ width: size, height: size }}
      >
        <svg
          viewBox="0 0 48 48"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className="w-full h-full"
        >
          {/* Cyber Shield Outer Contour */}
          <path
            d="M24 4L40 10V22C40 32.5 33.2 41.8 24 44C14.8 41.8 8 32.5 8 22V10L24 4Z"
            fill="url(#spiderGrad)"
            stroke="#1d4ed8"
            strokeWidth="1.5"
          />

          {/* Cybernetic Web Lines */}
          <path
            d="M24 12V36M14 20L34 28M14 28L34 20"
            stroke="#93c5fd"
            strokeWidth="1.2"
            strokeOpacity="0.4"
            strokeLinecap="round"
          />

          {/* Spider Head & Body Core */}
          <circle cx="24" cy="18" r="3.2" fill="#ffffff" />
          <ellipse cx="24" cy="26" rx="4.5" ry="6" fill="#ffffff" />

          {/* Glowing Spider Eyes */}
          <circle cx="22.5" cy="17.2" r="0.8" fill="#38bdf8" />
          <circle cx="25.5" cy="17.2" r="0.8" fill="#38bdf8" />

          {/* Spider Cyber Legs (Left 4 legs) */}
          <path
            d="M21 22L14 16M20 24L12 23M20 27L12 30M21 29L15 35"
            stroke="#ffffff"
            strokeWidth="1.8"
            strokeLinecap="round"
            strokeLinejoin="round"
          />

          {/* Spider Cyber Legs (Right 4 legs) */}
          <path
            d="M27 22L34 16M28 24L36 23M28 27L36 30M27 29L33 35"
            stroke="#ffffff"
            strokeWidth="1.8"
            strokeLinecap="round"
            strokeLinejoin="round"
          />

          {/* Linear Gradients */}
          <defs>
            <linearGradient id="spiderGrad" x1="8" y1="4" x2="40" y2="44" gradientUnits="userSpaceOnUse">
              <stop stopColor="#2563eb" />
              <stop offset="0.5" stopColor="#1d4ed8" />
              <stop offset="1" stopColor="#0f172a" />
            </linearGradient>
          </defs>
        </svg>
      </div>

      {variant === "full" && (
        <div className="flex flex-col">
          <div className="flex items-center gap-1.5 font-mono font-bold tracking-wider text-slate-900 text-base">
            <span>SPIDER</span>
            <span className="rounded bg-blue-50 px-1.5 py-0.2 text-[9px] font-semibold text-accent ring-1 ring-blue-200">
              v0.2
            </span>
          </div>
          <span className="text-[11px] font-medium text-slate-400 -mt-0.5">
            Runtime Defense Platform
          </span>
        </div>
      )}
    </div>
  );
}
