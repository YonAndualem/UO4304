"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Skeleton } from "@/components/ui/skeleton";
import { AppCard } from "@/components/AppCard";
import { EmptyState } from "@/components/EmptyState";
import { useIdentity } from "@/contexts/IdentityContext";
import { customerApi, ApiResponseError } from "@/lib/api";
import type { ApplicationDTO, ApplicationStatus } from "@/lib/types";

const ACTIVE_STATUSES: ApplicationStatus[] = ["PENDING", "SUBMITTED", "ACCEPTED", "ADJUSTED", "REREVIEW"];
const DONE_STATUSES: ApplicationStatus[] = ["APPROVED", "REJECTED", "CANCELLED"];

type Tab = "active" | "completed" | "all";

export default function CustomerApplicationsPage() {
  const { identity } = useIdentity();
  const router = useRouter();
  const [apps, setApps] = useState<ApplicationDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [tab, setTab] = useState<Tab>("active");

  useEffect(() => {
    if (!identity) { router.replace("/"); return; }
    if (identity.role !== "CUSTOMER") { router.replace("/"); return; }
    customerApi.list(identity)
      .then(setApps)
      .catch((e) => setError(e instanceof ApiResponseError ? e.body : e.message))
      .finally(() => setLoading(false));
  }, [identity, router]);

  const filtered = apps.filter((a) => {
    if (tab === "active") return ACTIVE_STATUSES.includes(a.status);
    if (tab === "completed") return DONE_STATUSES.includes(a.status);
    return true;
  });

  const needsAttentionCount = apps.filter((a) => a.status === "ADJUSTED").length;

  const tabClass = (t: Tab) =>
    `px-4 py-2 text-sm font-medium rounded-t-md border-b-2 transition-colors ${
      tab === t
        ? "border-primary text-primary"
        : "border-transparent text-gray-500 hover:text-gray-700"
    }`;

  return (
    <div className="max-w-3xl mx-auto py-8 px-4 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">My Applications</h1>
        <Link
          href="/customer/applications/new"
          className="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90 transition-colors"
        >
          + New Application
        </Link>
      </div>

      {needsAttentionCount > 0 && (
        <div className="bg-orange-50 border border-orange-300 rounded-md px-4 py-3 text-sm text-orange-800">
          <strong>{needsAttentionCount}</strong> application{needsAttentionCount > 1 ? "s" : ""} need{needsAttentionCount === 1 ? "s" : ""} your attention — adjustment requested by reviewer.
        </div>
      )}

      <div className="flex gap-1 border-b border-gray-200">
        <button className={tabClass("active")} onClick={() => setTab("active")}>
          Active {apps.filter((a) => ACTIVE_STATUSES.includes(a.status)).length > 0 && `(${apps.filter((a) => ACTIVE_STATUSES.includes(a.status)).length})`}
        </button>
        <button className={tabClass("completed")} onClick={() => setTab("completed")}>
          Completed {apps.filter((a) => DONE_STATUSES.includes(a.status)).length > 0 && `(${apps.filter((a) => DONE_STATUSES.includes(a.status)).length})`}
        </button>
        <button className={tabClass("all")} onClick={() => setTab("all")}>
          All {apps.length > 0 && `(${apps.length})`}
        </button>
      </div>

      {loading && (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-28 w-full rounded-xl" />)}
        </div>
      )}

      {error && <p className="text-red-600 text-sm">{error}</p>}

      {!loading && !error && filtered.length === 0 && (
        <EmptyState
          icon={tab === "completed" ? "✅" : "📄"}
          title={tab === "active" ? "No active applications" : tab === "completed" ? "No completed applications" : "No applications yet"}
          subtitle={tab === "active" ? "Applications in progress will appear here." : tab === "completed" ? "Approved, rejected, or cancelled applications will show here." : "Submit your first trade license application to get started."}
          demoHint="Want to see the full workflow? Use the Guided Demo to run through a complete scenario automatically."
          demoAction={{ label: "Open Guided Demo", href: "/guided-demo" }}
        />
      )}

      <div className="space-y-3">
        {filtered.map((app) => (
          <AppCard
            key={app.id}
            app={app}
            href={`/customer/applications/${app.id}`}
          />
        ))}
      </div>
    </div>
  );
}
