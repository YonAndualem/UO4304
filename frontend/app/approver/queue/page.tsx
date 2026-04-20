"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Skeleton } from "@/components/ui/skeleton";
import { AppCard } from "@/components/AppCard";
import { EmptyState } from "@/components/EmptyState";
import { useIdentity } from "@/contexts/IdentityContext";
import { useDemoMode } from "@/contexts/DemoModeContext";
import { approverApi, ApiResponseError } from "@/lib/api";
import type { ApplicationDTO } from "@/lib/types";

const SEED_IDS = new Set([
  "customer-seed-001", "customer-seed-002", "customer-seed-003",
  "customer-seed-004", "customer-seed-005", "customer-seed-006",
  "customer-seed-007", "customer-seed-008",
]);

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

  useEffect(() => {
    if (!identity) { router.replace("/"); return; }
    if (identity.role !== "APPROVER") { router.replace("/"); return; }

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
  }, [identity, router]);

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
    <div className="max-w-3xl mx-auto py-8 px-4 space-y-6">
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
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-28 w-full rounded-xl" />)}
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

      <div className="space-y-3">
        {displayed.map((app) => (
          <AppCard key={app.id} app={app} href={`/approver/applications/${app.id}`} />
        ))}
      </div>
    </div>
  );
}
