"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/StatusBadge";
import { EmptyState } from "@/components/EmptyState";
import { ActionModal, type ActionOption } from "@/components/ActionModal";
import { useIdentity } from "@/contexts/IdentityContext";
import { useDemoMode } from "@/contexts/DemoModeContext";
import { approverApi, ApiResponseError } from "@/lib/api";
import type { ApplicationDTO, ApproveAction } from "@/lib/types";

const SEED_IDS = new Set([
  "customer-seed-001", "customer-seed-002", "customer-seed-003",
  "customer-seed-004", "customer-seed-005", "customer-seed-006",
  "customer-seed-007", "customer-seed-008",
]);

const APPROVE_ACTIONS: ActionOption[] = [
  { value: "APPROVE",  label: "Approve",   requiresNotes: false, variant: "default"     },
  { value: "REREVIEW", label: "Re-Review", requiresNotes: true,  variant: "outline"     },
  { value: "REJECT",   label: "Reject",    requiresNotes: true,  variant: "destructive" },
];

type Tab = "active" | "processed" | "all";

export default function ApproverQueuePage() {
  const { identity } = useIdentity();
  const { isDemoMode } = useDemoMode();
  const router = useRouter();

  const [active, setActive]       = useState<ApplicationDTO[]>([]);
  const [processed, setProcessed] = useState<ApplicationDTO[]>([]);
  const [loading, setLoading]     = useState(true);
  const [error, setError]         = useState("");
  const [tab, setTab]             = useState<Tab>("active");

  const [selectedApp, setSelectedApp]       = useState<ApplicationDTO | null>(null);
  const [preSelectedAction, setPreSelected] = useState<string | undefined>(undefined);
  const [modalOpen, setModalOpen]           = useState(false);

  const loadData = useCallback(() => {
    if (!identity) return;
    setLoading(true);
    Promise.all([
      approverApi.list(identity),
      approverApi.list(identity, "APPROVED"),
      approverApi.list(identity, "REJECTED"),
    ])
      .then(([activeApps, approved, rejected]) => {
        setActive(activeApps);
        setProcessed([...approved, ...rejected].sort(
          (a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
        ));
      })
      .catch((e) => setError(e instanceof ApiResponseError ? e.body : e.message))
      .finally(() => setLoading(false));
  }, [identity]);

  useEffect(() => {
    if (!identity) { router.replace("/"); return; }
    if (identity.role !== "APPROVER") { router.replace("/"); return; }
    loadData();
  }, [identity, router, loadData]);

  function filterDemo(apps: ApplicationDTO[]) {
    if (isDemoMode) return apps;
    return apps.filter((a) => !SEED_IDS.has(a.applicant_id));
  }

  const allApps = filterDemo([
    ...active,
    ...processed.filter((a) => !active.some((x) => x.id === a.id)),
  ]);

  const displayed =
    tab === "active"    ? filterDemo(active) :
    tab === "processed" ? filterDemo(processed) :
    allApps;

  function openModal(app: ApplicationDTO, actionValue: string) {
    setSelectedApp(app);
    setPreSelected(actionValue);
    setModalOpen(true);
  }

  function closeModal() {
    setModalOpen(false);
    setSelectedApp(null);
    setPreSelected(undefined);
  }

  async function handleAction(action: string, notes: string) {
    if (!identity || !selectedApp) return;
    await approverApi.takeAction(identity, selectedApp.id, { action: action as ApproveAction, notes });
    closeModal();
    loadData();
  }

  const tabClass = (t: Tab) =>
    `px-4 py-2 text-sm font-medium rounded-t-md border-b-2 transition-colors ${
      tab === t
        ? "border-purple-600 text-purple-700"
        : "border-transparent text-gray-500 hover:text-gray-700"
    }`;

  function count(apps: ApplicationDTO[]) {
    const n = filterDemo(apps).length;
    return n > 0 ? ` (${n})` : "";
  }

  const emptyMessages: Record<Tab, { icon: string; title: string; subtitle: string }> = {
    active:    { icon: "📭", title: "Queue is empty",        subtitle: "No reviewed applications are awaiting your approval right now." },
    processed: { icon: "✅", title: "Nothing processed yet", subtitle: "Applications you have approved or rejected will appear here." },
    all:       { icon: "📋", title: "No applications",       subtitle: "Applications will appear here once accepted by a reviewer." },
  };

  const empty = emptyMessages[tab];

  return (
    <div className="max-w-6xl mx-auto py-8 px-4 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Approval Queue</h1>
          <p className="text-sm text-gray-500 mt-1">Reviewed applications awaiting your approval</p>
        </div>
        <span className="text-sm text-gray-400 bg-gray-100 rounded-full px-3 py-1">
          {displayed.length} item(s)
        </span>
      </div>

      <div className="flex gap-1 border-b border-gray-200">
        <button className={tabClass("active")}    onClick={() => setTab("active")}>
          Active{count(active)}
        </button>
        <button className={tabClass("processed")} onClick={() => setTab("processed")}>
          Processed{count(processed)}
        </button>
        <button className={tabClass("all")}       onClick={() => setTab("all")}>
          All{count(allApps)}
        </button>
      </div>

      {loading && (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-12 w-full rounded-md" />)}
        </div>
      )}

      {error && <p className="text-red-600 text-sm">{error}</p>}

      {!loading && !error && displayed.length === 0 && (
        <EmptyState
          icon={empty.icon}
          title={empty.title}
          subtitle={empty.subtitle}
          demoHint="Run the Happy Path scenario — after the reviewer accepts, the application appears here."
          demoAction={{ label: "Open Guided Demo", href: "/guided-demo?scenario=happy" }}
        />
      )}

      {!loading && !error && displayed.length > 0 && (
        <div className="overflow-x-auto rounded-xl border border-gray-200 shadow-sm">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 text-gray-500 uppercase text-xs tracking-wide">
              <tr>
                <th className="px-4 py-3 text-left font-medium">ID</th>
                <th className="px-4 py-3 text-left font-medium">License Type</th>
                <th className="px-4 py-3 text-left font-medium">Applicant</th>
                <th className="px-4 py-3 text-left font-medium">Commodity</th>
                <th className="px-4 py-3 text-left font-medium">Status</th>
                <th className="px-4 py-3 text-left font-medium">Submitted</th>
                <th className="px-4 py-3 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {displayed.map((app) => {
                const isActive = app.status === "ACCEPTED";
                return (
                  <tr
                    key={app.id}
                    className="hover:bg-gray-50 cursor-pointer transition-colors"
                    onClick={() => router.push(`/approver/applications/${app.id}`)}
                  >
                    <td className="px-4 py-3 font-mono text-xs text-gray-500">
                      {app.id.slice(0, 8)}…
                    </td>
                    <td className="px-4 py-3 text-gray-800">{app.license_type}</td>
                    <td className="px-4 py-3 text-gray-600 font-mono text-xs">{app.applicant_id}</td>
                    <td className="px-4 py-3 text-gray-700">{app.commodity?.name ?? "—"}</td>
                    <td className="px-4 py-3">
                      <StatusBadge status={app.status} />
                    </td>
                    <td className="px-4 py-3 text-gray-500 whitespace-nowrap">
                      {new Date(app.created_at).toLocaleDateString()}
                    </td>
                    <td
                      className="px-4 py-3 text-right"
                      onClick={(e) => e.stopPropagation()}
                    >
                      {isActive ? (
                        <div className="flex gap-1 justify-end">
                          <Button
                            size="sm"
                            variant="outline"
                            className="text-emerald-700 border-emerald-300 hover:bg-emerald-50"
                            onClick={() => openModal(app, "APPROVE")}
                          >
                            Approve
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            className="text-orange-700 border-orange-300 hover:bg-orange-50"
                            onClick={() => openModal(app, "REREVIEW")}
                          >
                            Re-Review
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            className="text-red-700 border-red-300 hover:bg-red-50"
                            onClick={() => openModal(app, "REJECT")}
                          >
                            Reject
                          </Button>
                        </div>
                      ) : (
                        <span className="text-gray-400 text-xs">—</span>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      <ActionModal
        open={modalOpen}
        onClose={closeModal}
        onSubmit={handleAction}
        actions={APPROVE_ACTIONS}
        title={`Approval Decision · ${selectedApp?.id.slice(0, 8) ?? ""}…`}
        preSelected={preSelectedAction}
      />
    </div>
  );
}
