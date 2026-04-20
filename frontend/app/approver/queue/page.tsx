"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Skeleton } from "@/components/ui/skeleton";
import { AppCard } from "@/components/AppCard";
import { EmptyState } from "@/components/EmptyState";
import { useIdentity } from "@/contexts/IdentityContext";
import { approverApi, ApiResponseError } from "@/lib/api";
import type { ApplicationDTO } from "@/lib/types";

export default function ApproverQueuePage() {
  const { identity } = useIdentity();
  const router = useRouter();
  const [apps, setApps] = useState<ApplicationDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!identity) { router.replace("/"); return; }
    if (identity.role !== "APPROVER") { router.replace("/"); return; }
    approverApi.list(identity)
      .then(setApps)
      .catch((e) => setError(e instanceof ApiResponseError ? e.body : e.message))
      .finally(() => setLoading(false));
  }, [identity, router]);

  return (
    <div className="max-w-3xl mx-auto py-8 px-4 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Approval Queue</h1>
          <p className="text-sm text-gray-500 mt-1">Reviewed applications awaiting your approval</p>
        </div>
        <span className="text-sm text-gray-400 bg-gray-100 rounded-full px-3 py-1">{apps.length} item(s)</span>
      </div>

      {loading && (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-28 w-full rounded-xl" />)}
        </div>
      )}

      {error && <p className="text-red-600 text-sm">{error}</p>}

      {!loading && !error && apps.length === 0 && (
        <EmptyState
          icon="📭"
          title="Queue is empty"
          subtitle="No applications have been accepted by a reviewer yet."
          demoHint="Run the Happy Path scenario — after the reviewer accepts, the application appears here."
          demoAction={{ label: "Open Guided Demo", href: "/guided-demo?scenario=happy" }}
        />
      )}

      <div className="space-y-3">
        {apps.map((app) => (
          <AppCard
            key={app.id}
            app={app}
            href={`/approver/applications/${app.id}`}
          />
        ))}
      </div>
    </div>
  );
}
