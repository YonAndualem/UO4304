"use client";

import { useState, useEffect, createContext, useContext } from "react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { StatusBadge } from "@/components/StatusBadge";
import { WorkflowTimeline } from "@/components/WorkflowTimeline";
import { customerApi, reviewerApi, approverApi, authApi, ApiResponseError } from "@/lib/api";
import type { ApplicationDTO, Identity, Role } from "@/lib/types";

// ── Demo identity context ─────────────────────────────────────────────────────
interface DemoIds { customer: Identity; reviewer: Identity; approver: Identity }
const DemoIdentityContext = createContext<DemoIds | null>(null);
function useDemoIds() {
  const ctx = useContext(DemoIdentityContext);
  if (!ctx) throw new Error("useDemoIds outside provider");
  return ctx;
}

// ── Scenario definitions ──────────────────────────────────────────────────────
interface Scenario {
  id: string;
  label: string;
  description: string;
  steps: string[];
}

const SCENARIOS: Scenario[] = [
  {
    id: "happy",
    label: "Happy Path",
    description: "Submit → Reviewer Accepts → Approver Approves",
    steps: ["submit", "review_accept", "approve_approve", "done"],
  },
  {
    id: "adjust_resubmit",
    label: "Adjust + Resubmit",
    description: "Submit → Reviewer Requests Adjustment → Customer Resubmits → Reviewer Accepts → Approver Approves",
    steps: ["submit", "review_adjust", "customer_resubmit", "review_accept", "approve_approve", "done"],
  },
  {
    id: "reviewer_reject",
    label: "Reviewer Rejects",
    description: "Submit → Reviewer Rejects",
    steps: ["submit", "review_reject", "done"],
  },
  {
    id: "approver_reject",
    label: "Approver Rejects",
    description: "Submit → Reviewer Accepts → Approver Rejects",
    steps: ["submit", "review_accept", "approve_reject", "done"],
  },
  {
    id: "rereview",
    label: "REREVIEW Cycle",
    description: "Submit → Reviewer Accepts → Approver Sends for Re-review → Reviewer Accepts → Approver Approves",
    steps: ["submit", "review_accept", "approve_rereview", "review_accept", "approve_approve", "done"],
  },
  {
    id: "cancel",
    label: "Customer Cancels",
    description: "Submit → Customer Cancels application",
    steps: ["submit", "customer_cancel", "done"],
  },
];

const STEP_INFO: Record<string, { label: string; role: string; color: string; description: string }> = {
  submit:           { label: "Submit",             role: "CUSTOMER",  color: "bg-blue-100 text-blue-700",   description: "Customer submits a new trade license application." },
  review_accept:    { label: "Reviewer: Accept",   role: "REVIEWER",  color: "bg-green-100 text-green-700",  description: "Reviewer accepts the application, forwarding it to the approver." },
  review_adjust:    { label: "Reviewer: Adjust",   role: "REVIEWER",  color: "bg-yellow-100 text-yellow-700", description: "Reviewer requests adjustment — application returns to customer." },
  review_reject:    { label: "Reviewer: Reject",   role: "REVIEWER",  color: "bg-red-100 text-red-700",    description: "Reviewer rejects the application." },
  customer_resubmit:{ label: "Customer: Resubmit", role: "CUSTOMER",  color: "bg-blue-100 text-blue-700",   description: "Customer updates details and resubmits for review." },
  customer_cancel:  { label: "Customer: Cancel",   role: "CUSTOMER",  color: "bg-gray-100 text-gray-700",   description: "Customer cancels the application." },
  approve_approve:  { label: "Approver: Approve",  role: "APPROVER",  color: "bg-purple-100 text-purple-700", description: "Approver grants final approval." },
  approve_reject:   { label: "Approver: Reject",   role: "APPROVER",  color: "bg-red-100 text-red-700",    description: "Approver rejects the application." },
  approve_rereview: { label: "Approver: Re-review",role: "APPROVER",  color: "bg-orange-100 text-orange-700", description: "Approver sends application back to reviewer." },
  done:             { label: "Done",               role: "",           color: "",                              description: "Flow complete." },
};

// ── Shared app summary ────────────────────────────────────────────────────────
function AppSummary({ app }: { app: ApplicationDTO }) {
  return (
    <div className="rounded-lg border border-gray-200 bg-gray-50 px-4 py-3 space-y-2">
      <div className="flex items-center justify-between">
        <p className="font-mono text-xs text-gray-400">{app.id}</p>
        <StatusBadge status={app.status} />
      </div>
      <WorkflowTimeline status={app.status} />
      {app.commodity && <p className="text-sm"><span className="text-gray-500">Commodity:</span> {app.commodity.name}</p>}
      {app.notes && <p className="text-xs bg-yellow-50 border border-yellow-100 rounded px-2 py-1 text-yellow-700">{app.notes}</p>}
    </div>
  );
}

// ── Step progress bar ─────────────────────────────────────────────────────────
function StepBar({ steps, current }: { steps: string[]; current: string }) {
  return (
    <div className="flex items-center gap-1 mb-6 overflow-x-auto pb-1">
      {steps.map((s, i) => {
        const ci = steps.indexOf(current);
        const done = i < ci;
        const active = i === ci;
        const info = STEP_INFO[s];
        return (
          <div key={`${s}-${i}`} className="flex items-center shrink-0">
            <div className="flex flex-col items-center min-w-[60px]">
              <div className={[
                "w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold border-2",
                done   ? "bg-emerald-500 border-emerald-500 text-white" : "",
                active ? "bg-blue-600 border-blue-600 text-white" : "",
                !done && !active ? "bg-white border-gray-300 text-gray-400" : "",
              ].join(" ")}>
                {done ? "✓" : i + 1}
              </div>
              <span className={[
                "text-[10px] mt-1 font-medium text-center leading-tight",
                done ? "text-emerald-600" : active ? "text-blue-700" : "text-gray-400",
              ].join(" ")}>{info?.label ?? s}</span>
            </div>
            {i < steps.length - 1 && (
              <div className={["h-0.5 w-6 mb-4 shrink-0", done ? "bg-emerald-400" : "bg-gray-200"].join(" ")} />
            )}
          </div>
        );
      })}
    </div>
  );
}

// ── Submit step ───────────────────────────────────────────────────────────────
function SubmitStep({ onDone }: { onDone: (app: ApplicationDTO) => void }) {
  const { customer } = useDemoIds();
  const [commodity, setCommodity] = useState("General Trading");
  const [category,  setCategory]  = useState("Commerce");
  const [docName,   setDocName]   = useState("Passport Copy");
  const [docUrl,    setDocUrl]    = useState("https://storage.example.com/test/passport.pdf");
  const [amount,    setAmount]    = useState("500");
  const [txnId,     setTxnId]     = useState("");
  const [loading,   setLoading]   = useState(false);
  const [error,     setError]     = useState("");

  useEffect(() => {
    setTxnId("TXN-" + crypto.randomUUID().replace(/-/g, "").slice(0, 12).toUpperCase());
  }, []);

  async function handle() {
    if (!txnId.trim()) { setError("Transaction ID is required."); return; }
    setLoading(true); setError("");
    try {
      const app = await customerApi.submit(customer, {
        license_type: "TRADE_LICENSE",
        commodity: { name: commodity, description: "Test flow commodity", category },
        documents: [{ name: docName, url: docUrl, content_type: "application/pdf" }],
        payment: { amount: parseFloat(amount), currency: "USD", transaction_id: txnId },
      });
      onDone(app);
    } catch (e) {
      setError(e instanceof ApiResponseError ? e.body : (e instanceof Error ? e.message : "Error"));
    } finally { setLoading(false); }
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-gray-500">Acting as <span className="font-mono font-medium text-blue-700">{customer.userId}</span></p>
      <div className="grid grid-cols-2 gap-3">
        <div><Label>Commodity name</Label><Input value={commodity} onChange={e => setCommodity(e.target.value)} /></div>
        <div><Label>Category</Label><Input value={category} onChange={e => setCategory(e.target.value)} /></div>
        <div><Label>Document name</Label><Input value={docName} onChange={e => setDocName(e.target.value)} /></div>
        <div><Label>Document URL</Label><Input value={docUrl} onChange={e => setDocUrl(e.target.value)} /></div>
        <div><Label>Amount (USD)</Label><Input type="number" value={amount} onChange={e => setAmount(e.target.value)} /></div>
        <div className="col-span-2">
          <Label>Transaction ID</Label>
          <div className="flex gap-2 mt-1">
            <Input value={txnId} onChange={e => setTxnId(e.target.value)} className="font-mono text-sm" placeholder="Generating…" />
            <Button type="button" variant="outline" onClick={() => setTxnId("TXN-" + crypto.randomUUID().replace(/-/g,"").slice(0,12).toUpperCase())} className="shrink-0">↻</Button>
          </div>
        </div>
      </div>
      {error && <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">{error}</p>}
      <Button onClick={handle} disabled={loading} className="w-full">{loading ? "Submitting…" : "Submit Application →"}</Button>
    </div>
  );
}

// ── Reviewer action step ──────────────────────────────────────────────────────
function ReviewerStep({ app, action, onDone }: { app: ApplicationDTO; action: "ACCEPT" | "REJECT" | "ADJUST"; onDone: (updated: ApplicationDTO) => void }) {
  const { reviewer } = useDemoIds();
  const [notes, setNotes] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const needsNotes = action !== "ACCEPT";

  async function handle() {
    if (needsNotes && !notes.trim()) { setError("Notes are required."); return; }
    setLoading(true); setError("");
    try {
      await reviewerApi.takeAction(reviewer, app.id, { action, notes: notes.trim() });
      const updated = await reviewerApi.get(reviewer, app.id);
      onDone(updated);
    } catch (e) {
      setError(e instanceof ApiResponseError ? e.body : (e instanceof Error ? e.message : "Error"));
    } finally { setLoading(false); }
  }

  const labels = { ACCEPT: "Accept →", REJECT: "Reject →", ADJUST: "Request Adjustment →" };

  return (
    <div className="space-y-4">
      <p className="text-sm text-gray-500">Acting as <span className="font-mono font-medium text-green-700">{reviewer.userId}</span></p>
      <AppSummary app={app} />
      {needsNotes && (
        <div>
          <Label>Notes <span className="text-red-500">*</span></Label>
          <Textarea value={notes} onChange={e => setNotes(e.target.value)} placeholder={action === "ADJUST" ? "What does the customer need to fix?" : "Reason for rejection…"} rows={2} />
        </div>
      )}
      {error && <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">{error}</p>}
      <Button onClick={handle} disabled={loading} className="w-full">{loading ? "Processing…" : labels[action]}</Button>
    </div>
  );
}

// ── Customer resubmit step ────────────────────────────────────────────────────
function CustomerResubmitStep({ app, onDone }: { app: ApplicationDTO; onDone: (updated: ApplicationDTO) => void }) {
  const { customer } = useDemoIds();
  const [commodity, setCommodity] = useState(app.commodity?.name ?? "");
  const [category,  setCategory]  = useState(app.commodity?.category ?? "");
  const [description, setDescription] = useState(app.commodity?.description ?? "");
  const [docName,   setDocName]   = useState(app.documents[0]?.name ?? "");
  const [docUrl,    setDocUrl]    = useState(app.documents[0]?.url ?? "");
  const [loading,   setLoading]   = useState(false);
  const [error,     setError]     = useState("");

  async function handle() {
    setLoading(true); setError("");
    try {
      await customerApi.resubmit(customer, app.id, {
        commodity: { name: commodity, description, category },
        documents: [{ name: docName, url: docUrl, content_type: "application/pdf" }],
      });
      const updated = await customerApi.get(customer, app.id);
      onDone(updated);
    } catch (e) {
      setError(e instanceof ApiResponseError ? e.body : (e instanceof Error ? e.message : "Error"));
    } finally { setLoading(false); }
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-gray-500">Acting as <span className="font-mono font-medium text-blue-700">{customer.userId}</span></p>
      {app.notes && (
        <div className="bg-orange-50 border border-orange-300 rounded px-3 py-2 text-sm text-orange-900">
          <strong>Reviewer notes:</strong> {app.notes}
        </div>
      )}
      <div className="grid grid-cols-2 gap-3">
        <div><Label>Commodity name</Label><Input value={commodity} onChange={e => setCommodity(e.target.value)} /></div>
        <div><Label>Category</Label><Input value={category} onChange={e => setCategory(e.target.value)} /></div>
        <div className="col-span-2"><Label>Description</Label><Input value={description} onChange={e => setDescription(e.target.value)} /></div>
        <div><Label>Document name</Label><Input value={docName} onChange={e => setDocName(e.target.value)} /></div>
        <div><Label>Document URL</Label><Input value={docUrl} onChange={e => setDocUrl(e.target.value)} /></div>
      </div>
      {error && <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">{error}</p>}
      <Button onClick={handle} disabled={loading} className="w-full">{loading ? "Resubmitting…" : "Resubmit for Review →"}</Button>
    </div>
  );
}

// ── Customer cancel step ──────────────────────────────────────────────────────
function CustomerCancelStep({ app, onDone }: { app: ApplicationDTO; onDone: (updated: ApplicationDTO) => void }) {
  const { customer } = useDemoIds();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handle() {
    setLoading(true); setError("");
    try {
      await customerApi.cancel(customer, app.id);
      const updated = await customerApi.get(customer, app.id);
      onDone(updated);
    } catch (e) {
      setError(e instanceof ApiResponseError ? e.body : (e instanceof Error ? e.message : "Error"));
    } finally { setLoading(false); }
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-gray-500">Acting as <span className="font-mono font-medium text-blue-700">{customer.userId}</span></p>
      <AppSummary app={app} />
      <p className="text-sm text-gray-600">Cancel this application. It will move to CANCELLED status and can be deleted.</p>
      {error && <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">{error}</p>}
      <Button variant="destructive" onClick={handle} disabled={loading} className="w-full">{loading ? "Cancelling…" : "Cancel Application →"}</Button>
    </div>
  );
}

// ── Approver action step ──────────────────────────────────────────────────────
function ApproverStep({ app, action, onDone }: { app: ApplicationDTO; action: "APPROVE" | "REJECT" | "REREVIEW"; onDone: (updated: ApplicationDTO) => void }) {
  const { approver } = useDemoIds();
  const [notes, setNotes] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const needsNotes = action !== "APPROVE";

  async function handle() {
    if (needsNotes && !notes.trim()) { setError("Notes are required."); return; }
    setLoading(true); setError("");
    try {
      await approverApi.takeAction(approver, app.id, { action, notes: notes.trim() });
      const updated = await approverApi.get(approver, app.id);
      onDone(updated);
    } catch (e) {
      setError(e instanceof ApiResponseError ? e.body : (e instanceof Error ? e.message : "Error"));
    } finally { setLoading(false); }
  }

  const labels = { APPROVE: "Approve →", REJECT: "Reject →", REREVIEW: "Send for Re-review →" };

  return (
    <div className="space-y-4">
      <p className="text-sm text-gray-500">Acting as <span className="font-mono font-medium text-purple-700">{approver.userId}</span></p>
      <AppSummary app={app} />
      {needsNotes && (
        <div>
          <Label>Notes <span className="text-red-500">*</span></Label>
          <Textarea value={notes} onChange={e => setNotes(e.target.value)} placeholder={action === "REREVIEW" ? "What needs re-review?" : "Reason for rejection…"} rows={2} />
        </div>
      )}
      {error && <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">{error}</p>}
      <Button onClick={handle} disabled={loading} className="w-full">{loading ? "Processing…" : labels[action]}</Button>
    </div>
  );
}

// ── Done ──────────────────────────────────────────────────────────────────────
function DoneStep({ app, onRestart }: { app: ApplicationDTO; onRestart: () => void }) {
  return (
    <div className="space-y-4 text-center">
      <div className="text-5xl">{app.status === "APPROVED" ? "🎉" : app.status === "CANCELLED" ? "🚫" : "❌"}</div>
      <p className="text-xl font-bold text-gray-900">Flow complete!</p>
      <div className="flex justify-center"><StatusBadge status={app.status} /></div>
      {app.notes && (
        <p className="text-sm bg-yellow-50 border border-yellow-200 rounded px-3 py-2 text-yellow-800">{app.notes}</p>
      )}
      <p className="text-xs text-gray-400 font-mono">{app.id}</p>
      <Link href={`/customer/applications/${app.id}`} className="block text-sm text-blue-600 hover:underline">View full detail with audit trail →</Link>
      <Button variant="outline" onClick={onRestart} className="w-full">Run another flow →</Button>
    </div>
  );
}

// ── Main page ─────────────────────────────────────────────────────────────────
export default function TestFlowPage() {
  const [demoIds, setDemoIds] = useState<DemoIds | null>(null);
  const [authError, setAuthError] = useState("");
  const [selectedScenario, setSelectedScenario] = useState<Scenario | null>(null);
  const [stepIndex, setStepIndex] = useState(0);
  const [app, setApp] = useState<ApplicationDTO | null>(null);

  useEffect(() => {
    async function resolve() {
      try {
        const [c, r, a] = await Promise.all([
          authApi.login("customer-seed-001", "demo"),
          authApi.login("reviewer-seed-001", "demo"),
          authApi.login("approver-seed-001", "demo"),
        ]);
        setDemoIds({
          customer: { userId: c.user_id, role: c.role as Role, token: c.token },
          reviewer: { userId: r.user_id, role: r.role as Role, token: r.token },
          approver: { userId: a.user_id, role: a.role as Role, token: a.token },
        });
      } catch {
        setAuthError("Could not authenticate demo accounts. Run: docker compose run --rm seed");
      }
    }
    resolve();
  }, []);

  const currentStepKey = selectedScenario ? selectedScenario.steps[stepIndex] : null;
  const stepInfo = currentStepKey ? STEP_INFO[currentStepKey] : null;

  function advance(updated: ApplicationDTO) {
    setApp(updated);
    setStepIndex(i => i + 1);
  }

  function restart() {
    setSelectedScenario(null);
    setStepIndex(0);
    setApp(null);
  }

  if (authError) {
    return <div className="max-w-2xl mx-auto py-10 px-4 text-red-600 text-sm">{authError}</div>;
  }

  if (!demoIds) {
    return <div className="max-w-2xl mx-auto py-10 px-4 text-gray-400 text-sm">Authenticating demo accounts…</div>;
  }

  if (!selectedScenario) {
    return (
      <DemoIdentityContext.Provider value={demoIds}>
      <div className="max-w-2xl mx-auto py-10 px-4">
        <h1 className="text-2xl font-bold text-gray-900 mb-1">Guided Demo</h1>
        <p className="text-sm text-gray-500 mb-8">Choose a scenario to walk through the full workflow without switching users manually.</p>

        <div className="grid gap-3">
          {SCENARIOS.map((s) => (
            <button
              key={s.id}
              onClick={() => setSelectedScenario(s)}
              className="text-left rounded-xl border border-gray-200 bg-white px-5 py-4 hover:border-blue-400 hover:shadow-sm transition-all"
            >
              <p className="font-semibold text-gray-900">{s.label}</p>
              <p className="text-sm text-gray-500 mt-1">{s.description}</p>
              <div className="flex flex-wrap gap-1 mt-2">
                {s.steps.filter(k => k !== "done").map((k, i) => {
                  const info = STEP_INFO[k];
                  return (
                    <span key={i} className={`text-[10px] font-semibold px-2 py-0.5 rounded-full ${info?.color || ""}`}>
                      {info?.label ?? k}
                    </span>
                  );
                })}
              </div>
            </button>
          ))}
        </div>

        <SeedTable />
      </div>
      </DemoIdentityContext.Provider>
    );
  }

  return (
    <DemoIdentityContext.Provider value={demoIds}>
    <div className="max-w-2xl mx-auto py-10 px-4">
      <div className="flex items-center gap-3 mb-6">
        <button onClick={restart} className="text-sm text-gray-400 hover:text-gray-600">← Scenarios</button>
        <span className="text-gray-300">|</span>
        <h1 className="text-lg font-bold text-gray-900">{selectedScenario.label}</h1>
      </div>

      <StepBar steps={selectedScenario.steps} current={currentStepKey ?? "done"} />

      {stepInfo && (
        <Card>
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle className="text-base">{stepInfo.label}</CardTitle>
              {stepInfo.role && (
                <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${stepInfo.color}`}>{stepInfo.role}</span>
              )}
            </div>
            <p className="text-sm text-gray-500">{stepInfo.description}</p>
          </CardHeader>
          <CardContent>
            {currentStepKey === "submit" && <SubmitStep onDone={advance} />}
            {currentStepKey === "review_accept"  && app && <ReviewerStep app={app} action="ACCEPT" onDone={advance} />}
            {currentStepKey === "review_adjust"  && app && <ReviewerStep app={app} action="ADJUST" onDone={advance} />}
            {currentStepKey === "review_reject"  && app && <ReviewerStep app={app} action="REJECT" onDone={advance} />}
            {currentStepKey === "customer_resubmit" && app && <CustomerResubmitStep app={app} onDone={advance} />}
            {currentStepKey === "customer_cancel"   && app && <CustomerCancelStep   app={app} onDone={advance} />}
            {currentStepKey === "approve_approve" && app && <ApproverStep app={app} action="APPROVE"  onDone={advance} />}
            {currentStepKey === "approve_reject"  && app && <ApproverStep app={app} action="REJECT"   onDone={advance} />}
            {currentStepKey === "approve_rereview"&& app && <ApproverStep app={app} action="REREVIEW" onDone={advance} />}
            {currentStepKey === "done" && app && <DoneStep app={app} onRestart={restart} />}
          </CardContent>
        </Card>
      )}
    </div>
    </DemoIdentityContext.Provider>
  );
}

// ── Seed credentials table ────────────────────────────────────────────────────
function SeedTable() {
  const users = [
    { userId: "customer-seed-001", role: "CUSTOMER", tag: "PENDING",   color: "bg-gray-100 text-gray-600",     note: "Used by test flow — can submit" },
    { userId: "customer-seed-002", role: "CUSTOMER", tag: "SUBMITTED", color: "bg-blue-100 text-blue-700",     note: "Awaiting reviewer" },
    { userId: "customer-seed-003", role: "CUSTOMER", tag: "ACCEPTED",  color: "bg-green-100 text-green-700",   note: "Awaiting approver" },
    { userId: "customer-seed-004", role: "CUSTOMER", tag: "APPROVED",  color: "bg-emerald-100 text-emerald-700", note: "Fully approved" },
    { userId: "customer-seed-005", role: "CUSTOMER", tag: "REJECTED",  color: "bg-red-100 text-red-700",       note: "Rejected by reviewer" },
    { userId: "customer-seed-006", role: "CUSTOMER", tag: "ADJUSTED",  color: "bg-yellow-100 text-yellow-700", note: "Needs customer correction" },
    { userId: "customer-seed-007", role: "CUSTOMER", tag: "REJECTED",  color: "bg-red-100 text-red-700",       note: "Rejected by approver" },
    { userId: "customer-seed-008", role: "CUSTOMER", tag: "REREVIEW",  color: "bg-orange-100 text-orange-700", note: "Sent back to reviewer" },
    { userId: "reviewer-seed-001", role: "REVIEWER", tag: "REVIEWER",  color: "bg-green-100 text-green-700",   note: "Sees SUBMITTED + REREVIEW queue" },
    { userId: "approver-seed-001", role: "APPROVER", tag: "APPROVER",  color: "bg-purple-100 text-purple-700", note: "Sees ACCEPTED queue" },
  ];

  return (
    <div className="mt-10 rounded-xl border border-gray-200 bg-white overflow-hidden">
      <div className="px-4 py-3 bg-gray-50 border-b border-gray-200">
        <p className="text-sm font-semibold text-gray-700">Seed Credentials</p>
        <p className="text-xs text-gray-500 mt-0.5">
          Quick-login from the <Link href="/" className="text-blue-600 hover:underline">home page</Link>.
        </p>
      </div>
      <div className="divide-y divide-gray-100">
        {users.map((u) => (
          <div key={u.userId} className="flex items-center gap-3 px-4 py-2.5">
            <span className="font-mono text-sm text-gray-800 w-44 shrink-0">{u.userId}</span>
            <span className={`text-xs font-semibold px-2 py-0.5 rounded-full shrink-0 ${u.color}`}>{u.tag}</span>
            <span className="text-xs text-gray-400">{u.note}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
