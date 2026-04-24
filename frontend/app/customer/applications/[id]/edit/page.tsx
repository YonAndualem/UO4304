"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { DocumentUpload, type UploadedDoc } from "@/components/DocumentUpload";
import { useIdentity } from "@/contexts/IdentityContext";
import { customerApi, ApiResponseError } from "@/lib/api";
import type { ApplicationDTO } from "@/lib/types";

export default function EditApplicationPage() {
  const { identity } = useIdentity();
  const router = useRouter();
  const { id } = useParams<{ id: string }>();
  const searchParams = useSearchParams();
  const isResubmit = searchParams.get("resubmit") === "1";

  const [original, setOriginal] = useState<ApplicationDTO | null>(null);
  const [loadError, setLoadError] = useState("");

  const [commodity, setCommodity] = useState({ name: "", description: "", category: "" });
  const [doc, setDoc] = useState<UploadedDoc | null>(null);
  const [updatePayment, setUpdatePayment] = useState(false);
  const [payment, setPayment] = useState({ amount: "", currency: "USD", transaction_id: "" });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    setPayment(p => ({
      ...p,
      transaction_id: "TXN-" + crypto.randomUUID().replace(/-/g, "").slice(0, 12).toUpperCase(),
    }));
  }, []);

  useEffect(() => {
    if (!identity) { router.replace("/"); return; }
    customerApi.get(identity, id)
      .then((app) => {
        setOriginal(app);
        if (app.commodity) {
          setCommodity({ name: app.commodity.name, description: app.commodity.description, category: app.commodity.category });
        }
        if (app.documents.length > 0) {
          const d = app.documents[0];
          setDoc({ key: d.url, name: d.name, content_type: d.content_type });
        }
        if (app.payment) {
          setPayment({ amount: String(app.payment.amount), currency: app.payment.currency, transaction_id: app.payment.transaction_id });
        }
      })
      .catch((e) => setLoadError(e instanceof ApiResponseError ? e.body : e.message));
  }, [identity, id, router]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!identity || !original) return;

    if (!commodity.name.trim()) { setError("Commodity name is required."); return; }
    if (!doc) { setError("Please upload a document."); return; }

    if (updatePayment) {
      const amt = parseFloat(payment.amount);
      if (isNaN(amt) || amt <= 0) { setError("Payment amount must be a positive number."); return; }
      if (!payment.transaction_id.trim()) { setError("Transaction ID is required."); return; }
    }

    setError("");
    setSubmitting(true);

    const paymentPayload = updatePayment ? {
      amount: parseFloat(payment.amount),
      currency: payment.currency,
      transaction_id: payment.transaction_id.trim(),
    } : undefined;

    const payload = {
      commodity: { name: commodity.name.trim(), description: commodity.description.trim(), category: commodity.category.trim() },
      documents: [{ name: doc.name, url: doc.key, content_type: doc.content_type }],
      payment: paymentPayload,
    };

    try {
      if (isResubmit) {
        await customerApi.resubmit(identity, id, payload);
      } else {
        await customerApi.update(identity, id, payload);
      }
      router.push(`/customer/applications/${id}`);
    } catch (e) {
      setError(e instanceof ApiResponseError ? e.body : (e instanceof Error ? e.message : "Unexpected error"));
    } finally {
      setSubmitting(false);
    }
  }

  if (loadError) {
    return <div className="max-w-xl mx-auto py-8 px-4"><p className="text-red-600">{loadError}</p></div>;
  }

  if (!original) {
    return (
      <div className="max-w-xl mx-auto py-8 px-4 space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-48 w-full" />
      </div>
    );
  }

  return (
    <div className="max-w-xl mx-auto py-8 px-4">
      <h1 className="text-2xl font-bold mb-2">
        {isResubmit ? "Resubmit Application" : "Edit Application"}
      </h1>

      {isResubmit && original.notes && (
        <div className="bg-orange-50 border border-orange-300 rounded-md px-4 py-3 mb-6 text-sm">
          <p className="font-semibold text-orange-800 mb-1">Reviewer notes — please address these before resubmitting:</p>
          <p className="text-orange-900">{original.notes}</p>
        </div>
      )}

      <p className="text-xs text-gray-400 font-mono mb-6">{id}</p>

      <form onSubmit={handleSubmit} className="space-y-6">

        <Card>
          <CardHeader><CardTitle className="text-base">Step 1 — Commodity</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            <div>
              <Label>Name *</Label>
              <Input value={commodity.name} onChange={(e) => setCommodity({ ...commodity, name: e.target.value })} placeholder="e.g. General Trading" />
            </div>
            <div>
              <Label>Category</Label>
              <Input value={commodity.category} onChange={(e) => setCommodity({ ...commodity, category: e.target.value })} placeholder="e.g. Commerce" />
            </div>
            <div>
              <Label>Description</Label>
              <Input value={commodity.description} onChange={(e) => setCommodity({ ...commodity, description: e.target.value })} placeholder="Short description" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle className="text-base">Step 2 — Document</CardTitle></CardHeader>
          <CardContent>
            {identity && (
              <DocumentUpload
                identity={identity}
                value={doc}
                onChange={setDoc}
              />
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-base">Step 3 — Payment</CardTitle>
              <button
                type="button"
                onClick={() => setUpdatePayment(!updatePayment)}
                className={`text-xs font-semibold px-3 py-1 rounded-full border transition-colors ${
                  updatePayment
                    ? "bg-blue-50 border-blue-300 text-blue-700"
                    : "bg-gray-50 border-gray-200 text-gray-500 hover:border-gray-400"
                }`}
              >
                {updatePayment ? "✓ Updating payment" : "Update payment"}
              </button>
            </div>
          </CardHeader>
          <CardContent>
            {!updatePayment && original.payment ? (
              <div className="text-sm text-gray-600 bg-gray-50 rounded px-3 py-2 border border-gray-200">
                <p>Current settlement: <strong>{original.payment.amount} {original.payment.currency}</strong></p>
                <p className="text-xs text-gray-400 font-mono mt-1">TXN: {original.payment.transaction_id}</p>
                <p className="text-xs text-gray-400 mt-1">Click &ldquo;Update payment&rdquo; above to change these details.</p>
              </div>
            ) : (
              <div className="space-y-3">
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <Label>Amount *</Label>
                    <Input type="number" min="0.01" step="0.01" value={payment.amount} onChange={(e) => setPayment({ ...payment, amount: e.target.value })} placeholder="500.00" />
                  </div>
                  <div>
                    <Label>Currency</Label>
                    <Input value={payment.currency} onChange={(e) => setPayment({ ...payment, currency: e.target.value })} placeholder="USD" maxLength={3} />
                  </div>
                </div>
                <div>
                  <Label>Transaction ID *</Label>
                  <div className="flex gap-2 mt-1">
                    <Input
                      value={payment.transaction_id}
                      onChange={(e) => setPayment({ ...payment, transaction_id: e.target.value })}
                      className="font-mono text-sm"
                    />
                    <Button
                      type="button"
                      variant="outline"
                      className="shrink-0"
                      onClick={() => setPayment(p => ({ ...p, transaction_id: "TXN-" + crypto.randomUUID().replace(/-/g, "").slice(0, 12).toUpperCase() }))}
                    >
                      ↻
                    </Button>
                  </div>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {error && <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">{error}</p>}

        <div className="flex gap-3">
          <Button type="button" variant="outline" onClick={() => router.back()}>Cancel</Button>
          <Button type="submit" disabled={submitting} className="flex-1">
            {submitting ? "Saving…" : isResubmit ? "Submit for Review" : "Save Changes"}
          </Button>
        </div>
      </form>
    </div>
  );
}
