import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { SpiderLogo } from "@/components/brand/SpiderLogo";
import { ItsCreditBadge } from "@/components/brand/ItsCreditBadge";
import { useAuth } from "@/hooks/useAuth";
import {
  ArrowRight,
  Eye,
  EyeOff,
  Lock,
  Mail,
  ShieldAlert,
  Sparkles,
  Zap,
} from "lucide-react";

export function LoginPage() {
  const login = useAuth((s) => s.login);
  const navigate = useNavigate();
  const [email, setEmail] = useState("admin@spider.local");
  const [password, setPassword] = useState("spider-admin");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setPending(true);
    setError(null);
    try {
      await login(email, password);
      navigate("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed. Please check your credentials.");
    } finally {
      setPending(false);
    }
  };

  return (
    <div className="flex min-h-screen w-full bg-slate-950 font-sans text-slate-100 selection:bg-accent selection:text-white">
      {/* Left Side: Cyber-Defense Hero Panel */}
      <div className="relative hidden lg:flex lg:w-3/5 flex-col justify-between overflow-hidden border-r border-slate-800/80 bg-gradient-to-br from-slate-950 via-slate-900 to-blue-950/40 p-12">
        {/* Background Cybernetic Grid & Glow */}
        <div className="absolute inset-0 bg-[linear-gradient(to_right,#1e293b18_1px,transparent_1px),linear-gradient(to_bottom,#1e293b18_1px,transparent_1px)] bg-[size:32px_32px] pointer-events-none" />
        <div className="absolute -left-32 top-1/4 h-96 w-96 rounded-full bg-blue-600/15 blur-3xl pointer-events-none" />
        <div className="absolute right-0 bottom-10 h-96 w-96 rounded-full bg-indigo-600/10 blur-3xl pointer-events-none" />

        {/* Top Header */}
        <div className="relative z-10 flex items-center justify-between">
          <SpiderLogo size={42} variant="full" glow className="[&_span]:text-white [&_div>span:last-child]:text-slate-400" />
          <div className="flex items-center gap-2 rounded-full border border-emerald-500/30 bg-emerald-950/40 px-3 py-1 text-xs font-mono text-emerald-400 backdrop-blur-sm">
            <span className="relative flex h-2 w-2">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75" />
              <span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-500" />
            </span>
            DEFENSE_ACTIVE
          </div>
        </div>

        {/* Center Content */}
        <div className="relative z-10 max-w-xl space-y-8 my-auto py-8">
          <div className="space-y-4">
            <div className="inline-flex items-center gap-2 rounded-lg border border-blue-500/30 bg-blue-950/50 px-3 py-1 text-xs font-semibold text-blue-300 backdrop-blur-sm">
              <Sparkles className="h-3.5 w-3.5 text-blue-400" />
              Next-Gen LLM Security & Serving Gateway
            </div>
            <h1 className="text-4xl font-extrabold tracking-tight text-white sm:text-5xl leading-tight">
              Runtime Defense Against{" "}
              <span className="bg-gradient-to-r from-blue-400 via-sky-300 to-indigo-300 bg-clip-text text-transparent">
                Prompt Injection
              </span>
            </h1>
            <p className="text-sm text-slate-300 leading-relaxed">
              SPIDER enforces pre-inference detection, chunk-level threat scoring with fine-tuned
              Flan-T5 models, and cluster-aware LLM serving isolation.
            </p>
          </div>

          {/* Feature Highlight Pills */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
            <div className="flex items-center gap-3 rounded-xl border border-slate-800 bg-slate-900/60 p-3.5 backdrop-blur-xs">
              <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-blue-500/10 text-blue-400 ring-1 ring-blue-500/20">
                <ShieldAlert className="h-4 w-4" />
              </div>
              <div className="text-xs">
                <div className="font-semibold text-white">Strict Invariant</div>
                <div className="text-slate-400 text-[11px]">BLOCK skips LLM completely</div>
              </div>
            </div>

            <div className="flex items-center gap-3 rounded-xl border border-slate-800 bg-slate-900/60 p-3.5 backdrop-blur-xs">
              <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-400 ring-1 ring-emerald-500/20">
                <Zap className="h-4 w-4" />
              </div>
              <div className="text-xs">
                <div className="font-semibold text-white">Sub-15ms Overhead</div>
                <div className="text-slate-400 text-[11px]">Optimized sliding chunker</div>
              </div>
            </div>
          </div>
        </div>

        {/* Bottom Academic Collaboration Credit */}
        <div className="relative z-10 pt-6 border-t border-slate-800/80">
          <ItsCreditBadge variant="hero" />
        </div>
      </div>

      {/* Right Side: Auth Form Panel */}
      <div className="flex w-full lg:w-2/5 flex-col justify-between bg-slate-900/90 p-8 sm:p-12 lg:p-14">
        {/* Mobile Header */}
        <div className="flex items-center justify-between lg:hidden pb-6 border-b border-slate-800">
          <SpiderLogo size={36} variant="full" glow className="[&_span]:text-white [&_div>span:last-child]:text-slate-400" />
        </div>

        {/* Center Form Container */}
        <div className="mx-auto w-full max-w-sm space-y-6 my-auto py-6">
          <div className="space-y-2">
            <h2 className="text-2xl font-bold tracking-tight text-white">Control Plane Login</h2>
            <p className="text-xs text-slate-400">
              Enter your administrative credentials to access real-time defense telemetry.
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Email Field */}
            <div className="space-y-1.5">
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-400">
                Email Address
              </label>
              <div className="relative">
                <Mail className="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="admin@spider.local"
                  className="w-full rounded-xl border border-slate-700 bg-slate-800/80 pl-10 pr-4 py-2.5 font-mono text-sm text-white placeholder:text-slate-500 outline-none transition focus:border-accent focus:bg-slate-800 focus:ring-2 focus:ring-accent/30"
                  required
                />
              </div>
            </div>

            {/* Password Field */}
            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <label className="block text-xs font-semibold uppercase tracking-wider text-slate-400">
                  Password
                </label>
              </div>
              <div className="relative">
                <Lock className="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
                <input
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••••••"
                  className="w-full rounded-xl border border-slate-700 bg-slate-800/80 pl-10 pr-10 py-2.5 font-mono text-sm text-white placeholder:text-slate-500 outline-none transition focus:border-accent focus:bg-slate-800 focus:ring-2 focus:ring-accent/30"
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3.5 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
            </div>

            {/* Error Message */}
            {error && (
              <div className="rounded-xl border border-rose-500/30 bg-rose-950/50 p-3 text-xs font-medium text-rose-300 backdrop-blur-sm">
                {error}
              </div>
            )}

            {/* Submit Button */}
            <Button
              type="submit"
              disabled={pending}
              className="w-full h-11 text-sm font-semibold shadow-lg shadow-accent/25 hover:shadow-accent/40"
              icon={!pending ? <ArrowRight className="h-4 w-4" /> : undefined}
            >
              {pending ? "Authenticating..." : "Sign In to Control Plane"}
            </Button>

            {/* Quick Autofill Lab Button */}
            <div className="rounded-xl border border-slate-800 bg-slate-950/60 p-3 text-center space-y-1.5">
              <div className="text-[11px] text-slate-400">Development Lab Defaults</div>
              <div className="flex items-center justify-center gap-2 font-mono text-xs text-slate-300">
                <span className="bg-slate-800 px-2 py-0.5 rounded text-[11px]">admin@spider.local</span>
                <span className="text-slate-500">/</span>
                <span className="bg-slate-800 px-2 py-0.5 rounded text-[11px]">spider-admin</span>
              </div>
            </div>
          </form>
        </div>

        {/* Mobile / Bottom Footer Credit */}
        <div className="text-center text-xs text-slate-500 pt-4 border-t border-slate-800/80">
          <div className="font-medium text-slate-400">Institut Teknologi Sepuluh Nopember (ITS) Surabaya</div>
          <div className="text-[10px] text-slate-500 mt-0.5">SPIDER Runtime Defense System · Open Source R&D</div>
        </div>
      </div>
    </div>
  );
}
