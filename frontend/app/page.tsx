"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { useIdentity } from "@/contexts/IdentityContext";
import { isValidUserId } from "@/lib/identity";
import type { Role } from "@/lib/types";

const ROLES: { value: Role; label: string; description: string; href: string; selected: string }[] = [
  { value: "CUSTOMER", label: "Customer",  description: "Submit and track trade license applications", href: "/customer/applications", selected: "border-blue-500 bg-blue-50" },
  { value: "REVIEWER", label: "Reviewer",  description: "Review submitted applications and take action",    href: "/reviewer/queue",         selected: "border-green-500 bg-green-50" },
  { value: "APPROVER", label: "Approver",  description: "Approve reviewed applications for final decision", href: "/approver/queue",         selected: "border-purple-500 bg-purple-50" },
];

// Seed credentials — matches cmd/seed/main.go exactly
const SEED_USERS: { userId: string; role: Role; tag: string; tagColor: string }[] = [
  { userId: "customer-seed-001", role: "CUSTOMER", tag: "PENDING",   tagColor: "bg-gray-100 text-gray-600" },
  { userId: "customer-seed-002", role: "CUSTOMER", tag: "SUBMITTED", tagColor: "bg-blue-100 text-blue-700" },
  { userId: "customer-seed-003", role: "CUSTOMER", tag: "ACCEPTED",  tagColor: "bg-green-100 text-green-700" },
  { userId: "customer-seed-004", role: "CUSTOMER", tag: "APPROVED",  tagColor: "bg-emerald-100 text-emerald-700" },
  { userId: "customer-seed-005", role: "CUSTOMER", tag: "REJECTED",  tagColor: "bg-red-100 text-red-700" },
  { userId: "customer-seed-006", role: "CUSTOMER", tag: "ADJUSTED",  tagColor: "bg-yellow-100 text-yellow-700" },
  { userId: "customer-seed-007", role: "CUSTOMER", tag: "REJECTED",  tagColor: "bg-red-100 text-red-700" },
  { userId: "customer-seed-008", role: "CUSTOMER", tag: "REREVIEW",  tagColor: "bg-orange-100 text-orange-700" },
  { userId: "reviewer-seed-001", role: "REVIEWER", tag: "REVIEWER",  tagColor: "bg-green-100 text-green-700" },
  { userId: "approver-seed-001", role: "APPROVER", tag: "APPROVER",  tagColor: "bg-purple-100 text-purple-700" },
];

export default function HomePage() {
  const { identity, setIdentity } = useIdentity();
  const router = useRouter();
  const [userId, setUserId] = useState("");
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (identity) {
      const cfg = ROLES.find((r) => r.value === identity.role);
      if (cfg) router.replace(cfg.href);
    }
  }, [identity, router]);

  function handleEnter() {
    if (!userId.trim()) { setError("User ID is required."); return; }
    if (!isValidUserId(userId.trim())) {
      setError("User ID may only contain letters, numbers, hyphens, and underscores (max 64 chars).");
      return;
    }
    if (!selectedRole) { setError("Please select a role."); return; }
    setError("");
    const cfg = ROLES.find((r) => r.value === selectedRole)!;
    setIdentity({ userId: userId.trim(), role: selectedRole });
    router.push(cfg.href);
  }

  function quickLogin(u: typeof SEED_USERS[number]) {
    setIdentity({ userId: u.userId, role: u.role });
    const cfg = ROLES.find((r) => r.value === u.role)!;
    router.push(cfg.href);
  }

  return (
    <div className="min-h-screen bg-linear-to-br from-slate-50 to-blue-50 flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-2xl space-y-6">
        <div className="text-center space-y-1">
          <h1 className="text-3xl font-bold text-gray-900">Trade License Portal</h1>
          <p className="text-gray-500">Select your role to continue</p>
        </div>

        <div className="grid md:grid-cols-2 gap-6">
          {/* Manual login */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Sign In</CardTitle>
              <CardDescription>Enter any user ID and pick a role.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <Label htmlFor="user-id">User ID</Label>
                <Input
                  id="user-id"
                  value={userId}
                  onChange={(e) => { setUserId(e.target.value); setError(""); }}
                  placeholder="e.g. customer-seed-001"
                  maxLength={64}
                  autoComplete="off"
                />
              </div>

              <div className="space-y-2">
                <Label>Role</Label>
                <div className="grid gap-2">
                  {ROLES.map((r) => (
                    <button
                      key={r.value}
                      type="button"
                      onClick={() => { setSelectedRole(r.value); setError(""); }}
                      className={[
                        "text-left rounded-lg border-2 px-3 py-2 transition-colors",
                        selectedRole === r.value ? r.selected : "bg-white border-gray-200 hover:border-gray-300",
                      ].join(" ")}
                    >
                      <p className="font-semibold text-sm">{r.label}</p>
                      <p className="text-xs text-gray-500">{r.description}</p>
                    </button>
                  ))}
                </div>
              </div>

              {error && <p className="text-sm text-red-600">{error}</p>}

              <Button className="w-full" onClick={handleEnter} disabled={!userId || !selectedRole}>
                Enter Portal
              </Button>
            </CardContent>
          </Card>

          {/* Seed quick-login */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Seed Credentials</CardTitle>
              <CardDescription>
                Pre-seeded users — one click to enter.{" "}
                <Link href="/test-flow" className="text-blue-600 hover:underline">Run full flow →</Link>
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-1.5">
                {SEED_USERS.map((u) => (
                  <button
                    key={u.userId}
                    type="button"
                    onClick={() => quickLogin(u)}
                    className="w-full flex items-center justify-between rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm hover:border-blue-300 hover:bg-blue-50 transition-colors group"
                  >
                    <span className="font-mono text-gray-700 group-hover:text-blue-700">{u.userId}</span>
                    <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${u.tagColor}`}>{u.tag}</span>
                  </button>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
