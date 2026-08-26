import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useAuth } from "@/hooks/useAuth";

export function LoginPage() {
  const login = useAuth((s) => s.login);
  const navigate = useNavigate();
  const [email, setEmail] = useState("admin@spider.local");
  const [password, setPassword] = useState("spider-admin");
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  return (
    <div className="flex min-h-screen items-center justify-center bg-canvas p-6">
      <Card className="w-full max-w-md space-y-4">
        <div>
          <div className="font-mono text-xs uppercase tracking-[0.25em] text-orange-400">SPIDER</div>
          <h1 className="mt-2 text-xl font-medium">Sign in</h1>
          <p className="mt-1 text-sm text-zinc-400">
            Runtime defense control plane. Default lab credentials are prefilled.
          </p>
        </div>
        <form
          className="space-y-3"
          onSubmit={async (event) => {
            event.preventDefault();
            setPending(true);
            setError(null);
            try {
              await login(email, password);
              navigate("/dashboard");
            } catch (err) {
              setError(err instanceof Error ? err.message : "Login failed");
            } finally {
              setPending(false);
            }
          }}
        >
          <label className="block text-xs text-zinc-400">
            Email
            <Input className="mt-1" value={email} onChange={(e) => setEmail(e.target.value)} />
          </label>
          <label className="block text-xs text-zinc-400">
            Password
            <Input
              className="mt-1"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </label>
          {error ? <p className="text-sm text-red-400">{error}</p> : null}
          <Button type="submit" disabled={pending} className="w-full">
            {pending ? "Signing in…" : "Continue"}
          </Button>
        </form>
      </Card>
    </div>
  );
}
