"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useIdentity } from "@/contexts/IdentityContext";
import { customerApi, ApiResponseError } from "@/lib/api";

export default function NewApplicationPage() {
  const { identity } = useIdentity();
  const router = useRouter();

  const [commodity, setCommodity] = useState({ name: "", description: "", category: "" });
  const [doc, setDoc] = useState({ name: "", url: "", content_type: "application/pdf" });
  const [payment, setPayment] = useState({ amount: "", currency: "USD", transaction_id: "" });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    setPayment(p => ({ ...p, transaction_id: "TXN-" + crypto.randomUUID().replace(/-/g, "").slice(0, 12).toUpperCase() }));
  }, []);

  // Validate doc URL: must be http/https to prevent javascript: injection
  function isValidUrl(url: string): boolean {
    try {
      const u = new URL(url);
      return u.protocol === "http:" || u.protocol === "https:";
    } catch {
      return false;
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!identity) { router.replace("/"); return; }

    if (!commodity.name.trim()) { setError("Commodity name is required."); return; }
    if (!doc.name.trim()) { setError("Document name is required."); return; }
    if (!isValidUrl(doc.url)) { setError("Document URL must be a valid http/https URL."); return; }
    const amount = parseFloat(payment.amount);
    if (isNaN(amount) || amount <= 0) { setError("Payment amount must be a positive number."); return; }
    if (!payment.transaction_id.trim()) { setError("Transaction ID is required."); return; }

    setError("");
    setLoading(true);
    try {
      const created = await customerApi.submit(identity, {
        license_type: "TRADE_LICENSE",
        commodity: {
          name: commodity.name.trim(),
          description: commodity.description.trim(),
          category: commodity.category.trim(),
        },
        documents: [{ name: doc.name.trim(), url: doc.url.trim(), content_type: doc.content_type }],
        payment: { amount, currency: payment.currency, transaction_id: payment.transaction_id.trim() },
      });
      // Validate returned ID is a UUID before embedding in URL
      if (!/^[0-9a-f-]{36}$/i.test(created.id)) throw new Error("Invalid application ID returned.");
      router.push(`/customer/applications/${created.id}`);
    } catch (e) {
      setError(e instanceof ApiResponseError ? e.body : (e instanceof Error ? e.message : "Unexpected error"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="max-w-xl mx-auto py-8 px-4">
      <h1 className="text-2xl font-bold mb-6">New Trade License Application</h1>
      <form onSubmit={handleSubmit} className="space-y-6">

        <Card>
          <CardHeader><CardTitle className="text-base">Commodity</CardTitle></CardHeader>
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
          <CardHeader><CardTitle className="text-base">Document</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            <div>
              <Label>Name *</Label>
              <Input value={doc.name} onChange={(e) => setDoc({ ...doc, name: e.target.value })} placeholder="e.g. Passport Copy" />
            </div>
            <div>
              <Label>URL * (https://…)</Label>
              <Input value={doc.url} onChange={(e) => setDoc({ ...doc, url: e.target.value })} placeholder="https://storage.example.com/file.pdf" />
            </div>
            <div>
              <Label>Content Type</Label>
              <Input value={doc.content_type} onChange={(e) => setDoc({ ...doc, content_type: e.target.value })} placeholder="application/pdf" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle className="text-base">Payment</CardTitle></CardHeader>
          <CardContent className="space-y-3">
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
                  placeholder="Generating…"
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
          </CardContent>
        </Card>

        {error && <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">{error}</p>}

        <div className="flex gap-3">
          <Button type="button" variant="outline" onClick={() => router.back()}>Cancel</Button>
          <Button type="submit" disabled={loading} className="flex-1">
            {loading ? "Submitting…" : "Submit Application"}
          </Button>
        </div>
      </form>
    </div>
  );
}
