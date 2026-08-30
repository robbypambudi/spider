import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { api } from "@/services/api";
import type { CreatePolicyRequest, PolicyView } from "@/types/api";

const emptyForm: CreatePolicyRequest = {
  name: "",
  threshold: 0.5,
  action_on_detection: "block",
  chunker: "token",
  chunk_size: 256,
  chunk_overlap: 0,
  is_default: false,
};

const labelClass = "text-xs font-medium uppercase tracking-wide text-slate-500";
const fieldClass = "space-y-1";

export function SettingsPage() {
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState<PolicyView | null>(null);
  const [form, setForm] = useState<CreatePolicyRequest>(emptyForm);
  const [error, setError] = useState<string | null>(null);

  const runtime = useQuery({
    queryKey: ["settings", "runtime"],
    queryFn: api.runtimeSettings,
    staleTime: 60_000,
  });
  const policies = useQuery({
    queryKey: ["security", "policies"],
    queryFn: api.policies,
    staleTime: 10_000,
  });

  const resetForm = () => {
    setEditing(null);
    setForm(emptyForm);
    setError(null);
  };

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["security", "policies"] });
  };

  const create = useMutation({
    mutationFn: api.createPolicy,
    onSuccess: () => {
      invalidate();
      resetForm();
    },
    onError: (err: Error) => setError(err.message),
  });

  const update = useMutation({
    mutationFn: ({ id, body }: { id: string; body: CreatePolicyRequest }) =>
      api.updatePolicy(id, {
        name: body.name,
        threshold: body.threshold,
        action_on_detection: body.action_on_detection,
        chunker: body.chunker,
        chunk_size: body.chunk_size,
        chunk_overlap: body.chunk_overlap,
        is_default: body.is_default,
      }),
    onSuccess: () => {
      invalidate();
      resetForm();
    },
    onError: (err: Error) => setError(err.message),
  });

  const remove = useMutation({
    mutationFn: api.deletePolicy,
    onSuccess: () => {
      invalidate();
      resetForm();
    },
    onError: (err: Error) => setError(err.message),
  });

  const activate = useMutation({
    mutationFn: api.activatePolicy,
    onSuccess: () => invalidate(),
    onError: (err: Error) => setError(err.message),
  });

  const startEdit = (policy: PolicyView) => {
    setEditing(policy);
    setForm({
      name: policy.name,
      threshold: policy.threshold,
      action_on_detection: policy.action_on_detection,
      chunker: policy.chunker,
      chunk_size: policy.chunk_size,
      chunk_overlap: policy.chunk_overlap,
      is_default: policy.is_default,
    });
    setError(null);
  };

  const submit = () => {
    setError(null);
    if (editing?.id) {
      update.mutate({ id: editing.id, body: form });
      return;
    }
    create.mutate(form);
  };

  const pending = create.isPending || update.isPending;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-slate-900">Settings</h1>
        <p className="text-sm text-slate-500">
          Manage threshold policies and token chunking parameters. Runtime env defaults are read-only.
        </p>
      </div>

      <Card className="space-y-2 font-mono text-sm text-slate-700">
        <h2 className="font-sans text-sm font-medium uppercase tracking-wide text-slate-500">
          Runtime env (read-only)
        </h2>
        <div>SPIDER_DEFAULT_DETECTOR={runtime.data?.default_detector ?? "—"}</div>
        <div>SPIDER_FAIL_MODE={runtime.data?.fail_mode ?? "—"}</div>
        <div>SPIDER_CHUNKER={runtime.data?.chunker ?? "—"}</div>
        <div>SPIDER_CHUNK_SIZE={runtime.data?.chunk_size ?? "—"}</div>
        <div>SPIDER_CHUNK_OVERLAP={runtime.data?.chunk_overlap ?? "—"}</div>
        <p className="pt-2 font-sans text-slate-500">
          Active scan threshold and chunk params come from the default policy below. Change detector via{" "}
          <Link className="text-accent hover:underline" to="/models">
            Models
          </Link>
          .
        </p>
      </Card>

      <Card className="space-y-4">
        <h2 className="text-sm font-medium uppercase tracking-wide text-slate-500">
          {editing ? "Edit policy" : "New policy"}
        </h2>
        <div className="grid gap-4 md:grid-cols-2">
          <div className={fieldClass}>
            <label className={labelClass}>Name</label>
            <Input
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="production-threshold"
            />
          </div>
          <div className={fieldClass}>
            <label className={labelClass}>Threshold (0–1)</label>
            <Input
              type="number"
              min={0}
              max={1}
              step={0.001}
              value={form.threshold}
              onChange={(e) => setForm({ ...form, threshold: Number(e.target.value) })}
            />
          </div>
          <div className={fieldClass}>
            <label className={labelClass}>Action on detection</label>
            <select
              className="w-full rounded-lg border border-line bg-white px-3 py-2 text-sm"
              value={form.action_on_detection}
              onChange={(e) => setForm({ ...form, action_on_detection: e.target.value })}
            >
              <option value="block">block</option>
              <option value="review">review</option>
            </select>
          </div>
          <div className={fieldClass}>
            <label className={labelClass}>Chunker</label>
            <select
              className="w-full rounded-lg border border-line bg-white px-3 py-2 text-sm"
              value={form.chunker}
              onChange={(e) => setForm({ ...form, chunker: e.target.value })}
            >
              <option value="token">token (lab-aligned)</option>
              <option value="fixed">fixed (characters)</option>
            </select>
          </div>
          <div className={fieldClass}>
            <label className={labelClass}>Chunk size (tokens/chars)</label>
            <Input
              type="number"
              min={1}
              value={form.chunk_size}
              onChange={(e) => setForm({ ...form, chunk_size: Number(e.target.value) })}
            />
          </div>
          <div className={fieldClass}>
            <label className={labelClass}>Chunk overlap</label>
            <Input
              type="number"
              min={0}
              value={form.chunk_overlap}
              onChange={(e) => setForm({ ...form, chunk_overlap: Number(e.target.value) })}
            />
          </div>
        </div>
        <label className="flex items-center gap-2 text-sm text-slate-700">
          <input
            checked={form.is_default}
            type="checkbox"
            onChange={(e) => setForm({ ...form, is_default: e.target.checked })}
          />
          Set as default policy
        </label>
        {error ? <p className="text-sm text-red-600">{error}</p> : null}
        <div className="flex flex-wrap gap-2">
          <Button disabled={pending || !form.name} onClick={submit}>
            {editing ? "Save changes" : "Create policy"}
          </Button>
          {editing ? (
            <Button className="bg-white text-slate-700 ring-1 ring-line hover:bg-slate-50" onClick={resetForm}>
              Cancel
            </Button>
          ) : null}
        </div>
      </Card>

      <Card className="divide-y divide-line p-0">
        <div className="px-4 py-3 text-sm font-medium uppercase tracking-wide text-slate-500">
          Policies
        </div>
        {(policies.data ?? []).map((policy) => (
          <div
            key={policy.id ?? policy.name}
            className="flex flex-wrap items-center justify-between gap-3 px-4 py-4"
          >
            <div className="space-y-1">
              <div className="flex flex-wrap items-center gap-2">
                <span className="font-medium text-slate-900">{policy.name}</span>
                {policy.is_default ? <Badge value="DEFAULT" /> : null}
              </div>
              <p className="font-mono text-xs text-slate-600">
                τ={policy.threshold.toFixed(4)} · {policy.action_on_detection} · {policy.chunker}{" "}
                {policy.chunk_size}/{policy.chunk_overlap}
              </p>
            </div>
            <div className="flex flex-wrap gap-2">
              {!policy.is_default && policy.id ? (
                <Button
                  disabled={activate.isPending}
                  onClick={() => activate.mutate(policy.id!)}
                >
                  Set default
                </Button>
              ) : null}
              <Button className="bg-white text-slate-700 ring-1 ring-line hover:bg-slate-50" onClick={() => startEdit(policy)}>
                Edit
              </Button>
              {policy.id && !policy.is_default ? (
                <Button
                  className="bg-white text-slate-700 ring-1 ring-line hover:bg-slate-50"
                  disabled={remove.isPending}
                  onClick={() => {
                    if (window.confirm(`Delete policy "${policy.name}"?`)) {
                      remove.mutate(policy.id!);
                    }
                  }}
                >
                  Delete
                </Button>
              ) : null}
            </div>
          </div>
        ))}
        {!policies.isLoading && (policies.data ?? []).length === 0 ? (
          <p className="px-4 py-6 text-sm text-slate-500">No policies yet. Create one above.</p>
        ) : null}
      </Card>
    </div>
  );
}
