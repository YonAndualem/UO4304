"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { AppDetail } from "@/components/AppDetail";
import { ActionModal, type ActionOption } from "@/components/ActionModal";
import { useIdentity } from "@/contexts/IdentityContext";
import { reviewerApi, ApiResponseError } from "@/lib/api";
import type { ApplicationDTO, ReviewAction } from "@/lib/types";

const REVIEW_ACTIONS: ActionOption[] = [
  { value: "ACCEPT",  label: "Accept",  requiresNotes: false, variant: "default" },
  { value: "REJECT",  label: "Reject",  requiresNotes: true,  variant: "destructive" },
  { value: "ADJUST",  label: "Adjust",  requiresNotes: true,  variant: "outline" },
];

export default function ReviewerApplicationDetailPage() {
  const { identity } = useIdentity();
  const router = useRouter();
  const { id } = useParams<{ id: string }>();
  const [app, setApp] = useState<ApplicationDTO | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [modalOpen, setModalOpen] = useState(false);

  const canAct = app?.status === "SUBMITTED" || app?.status === "REREVIEW";

  useEffect(() => {
    if (!identity) { router.replace("/"); return; }
    reviewerApi.get(identity, id)
      .then(setApp)
      .catch((e) => setError(e instanceof ApiResponseError ? e.body : e.message))
      .finally(() => setLoading(false));
  }, [identity, id, router]);

  async function handleAction(action: string, notes: string) {
    if (!identity) return;
    await reviewerApi.takeAction(identity, id, { action: action as ReviewAction, notes });
    router.push("/reviewer/queue");
  }

  return (
    <div className="max-w-2xl mx-auto py-8 px-4 space-y-6">
      <Button variant="ghost" onClick={() => router.back()} className="text-sm">← Back to Queue</Button>

      {loading && <div className="space-y-4"><Skeleton className="h-8 w-48" /><Skeleton className="h-32 w-full" /></div>}
      {error && <p className="text-red-600 text-sm">{error}</p>}

      {app && (
        <>
          <AppDetail app={app} />
          {canAct && (
            <Button onClick={() => setModalOpen(true)} className="w-full">
              Take Action
            </Button>
          )}
        </>
      )}

      <ActionModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSubmit={handleAction}
        actions={REVIEW_ACTIONS}
        title="Review Decision"
      />
    </div>
  );
}
