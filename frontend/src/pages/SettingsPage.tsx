import { Card } from "@/components/ui/card";

export function SettingsPage() {
  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-medium">Settings</h1>
      <Card className="space-y-2 font-mono text-sm text-zinc-300">
        <div>SPIDER_DEFAULT_DETECTOR=rule-based</div>
        <div>SPIDER_DEFAULT_THRESHOLD=0.5</div>
        <div>SPIDER_FAIL_MODE=closed</div>
        <div>SPIDER_LOG_PROMPT_CONTENT=false</div>
        <p className="pt-2 font-sans text-zinc-500">
          Thresholds are policy configuration, not detector internals. Change them via environment
          variables or the policy table — never inside detector code.
        </p>
      </Card>
    </div>
  );
}
