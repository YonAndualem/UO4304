"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { AppDetail } from "@/components/AppDetail";
import { useIdentity } from "@/contexts/IdentityContext";
import { customerApi, ApiResponseError } from "@/lib/api";
import type { ApplicationDTO } from "@/lib/types";

export default function CustomerApplicationDetailPage() {
  const { identity } = useIdentity();
  const router = useRouter();
  const { id } = useParams<{ id: string }>();
  const [app, setApp] = useState<ApplicationDTO | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [actionBusy, setActionBusy] = useState(false);

  useEffect(() => {
    if (!identity) { router.replace("/"); return; }
    customerApi.get(identity, id)
      .then(setApp)
      .catch((e) => setError(e instanceof ApiResponseError ? e.body : e.message))
      .finally(() => setLoading(false));
  }, [identity, id, router]);

  async function handleCancel() {
    if (!identity || !app) return;
    if (!confirm("Cancel this application?")) return;
    setActionBusy(true);
    setError("");
    try {
      await customerApi.cancel(identity, id);
      const updated = await customerApi.get(identity, id);
      setApp(updated);
    } catch (e) {
      setError(e instanceof ApiResponseError ? e.body : (e instanceof Error ? e.message : "Error"));
    } finally {
      setActionBusy(false);
    }
  }

  async function handleDelete() {
    if (!identity || !app) return;
    if (!confirm("Permanently delete this application? This cannot be undone.")) return;
    setActionBusy(true);
    setError("");
    try {
      await customerApi.delete(identity, id);
      router.push("/customer/applications");
    } catch (e) {
      setError(e instanceof ApiResponseError ? e.body : (e instanceof Error ? e.message : "Error"));
      setActionBusy(false);
    }
  }

  const canEdit = app?.status === "PENDING" || app?.status === "ADJUSTED";
  const canCancel = app?.status === "PENDING" || app?.status === "ADJUSTED";
  const canResubmit = app?.status === "ADJUSTED";
  const canDelete = app?.status === "PENDING" || app?.status === "CANCELLED" || app?.status === "REJECTED";

  return (
    <div className="max-w-2xl mx-auto py-8 px-4 space-y-6">
      <Button variant="ghost" onClick={() => router.back()} className="text-sm">← Back</Button>

      {loading && <div className="space-y-4"><Skeleton className="h-8 w-48" /><Skeleton className="h-32 w-full" /></div>}
      {error && <p className="text-red-600 text-sm bg-red-50 border border-red-200 rounded px-3 py-2">{error}</p>}

      {app && (
        <>
          {app.status === "ADJUSTED" && app.notes && (
            <div className="bg-orange-50 border border-orange-300 rounded-md px-4 py-3">
              <p className="text-sm font-semibold text-orange-800 mb-1">⚠ Reviewer Adjustment Notes</p>
              <p className="text-sm text-orange-900">{app.notes}</p>
              <p className="text-xs text-orange-600 mt-2">Please update your application and resubmit.</p>
            </div>
          )}

          <AppDetail app={app} />

          <div className="flex flex-wrap gap-3 pt-2">
            {canEdit && (
              <Link
                href={`/customer/applications/${id}/edit`}
                className="inline-flex items-center justify-center rounded-md bg-secondary px-4 py-2 text-sm font-medium text-secondary-foreground border border-input hover:bg-secondary/80 transition-colors"
              >
                Edit Details
              </Link>
            )}
            {canResubmit && (
              <Link
                href={`/customer/applications/${id}/edit?resubmit=1`}
                className="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90 transition-colors"
              >
                Resubmit
              </Link>
            )}
            {canCancel && (
              <Button variant="outline" onClick={handleCancel} disabled={actionBusy}>
                {actionBusy ? "Processing…" : "Cancel Application"}
              </Button>
            )}
            {canDelete && (
              <Button variant="destructive" onClick={handleDelete} disabled={actionBusy}>
                {actionBusy ? "Deleting…" : "Delete"}
              </Button>
            )}
          </div>
        </>
      )}
    </div>
  );
}
