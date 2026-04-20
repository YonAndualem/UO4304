"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { useIdentity } from "@/contexts/IdentityContext";
import { authApi, ApiResponseError } from "@/lib/api";
import type { Role } from "@/lib/types";

const ROLES: { value: Role; label: string; description: string; selected: string }[] = [
  { value: "CUSTOMER", label: "Customer",  description: "Submit and track trade license applications",       selected: "border-blue-500 bg-blue-50" },
  { value: "REVIEWER", label: "Reviewer",  description: "Review submitted applications and take action",    selected: "border-green-500 bg-green-50" },
  { value: "APPROVER", label: "Approver",  description: "Approve reviewed applications for final decision", selected: "border-purple-500 bg-purple-50" },
];

const ROLE_HOME: Record<Role, string> = {
  CUSTOMER: "/customer/applications",
  REVIEWER: "/reviewer/queue",
  APPROVER: "/approver/queue",
};

const DEMO_USERS: { userId: string; role: Role; tag: string; tagColor: string }[] = [
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

type Tab = "signin" | "demo";
type Mode = "login" | "register";

export default function HomePage() {
  const { identity, setIdentity } = useIdentity();
  const router = useRouter();

  const [tab, setTab] = useState<Tab>("signin");
  const [mode, setMode] = useState<Mode>("login");

  const [userId, setUserId] = useState("");
  const [password, setPassword] = useState("");
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [demoLoading, setDemoLoading] = useState<string | null>(null);

  useEffect(() => {
    if (identity) router.replace(ROLE_HOME[identity.role]);
  }, [identity, router]);

  function reset() {
    setUserId(""); setPassword(""); setSelectedRole(null); setError("");
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!userId.trim() || !password) return;
    if (mode === "register" && !selectedRole) { setError("Please select a role."); return; }

    setError(""); setSubmitting(true);
    try {
      if (mode === "register") {
        await authApi.register(userId.trim(), password, selectedRole!);
        // Auto-login after registration
      }
      const res = await authApi.login(userId.trim(), password);
      setIdentity({ userId: res.user_id, role: res.role as Role, token: res.token });
    } catch (e) {
      setError(e instanceof ApiResponseError ? JSON.parse(e.body || "{}").error ?? e.body : "Unexpected error");
    } finally {
      setSubmitting(false);
    }
  }

  async function quickLogin(u: typeof DEMO_USERS[number]) {
    setDemoLoading(u.userId);
    try {
      const res = await authApi.login(u.userId, "demo");
      setIdentity({ userId: res.user_id, role: res.role as Role, token: res.token });
    } catch {
      // fallback error is silently swallowed; user sees nothing happens
    } finally {
      setDemoLoading(null);
    }
  }

  return (
    <div className="min-h-screen bg-linear-to-br from-slate-50 to-blue-50 flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md space-y-6">
        <div className="text-center space-y-1">
          <h1 className="text-3xl font-bold text-gray-900">Trade License Portal</h1>
          <p className="text-gray-500">Enterprise workflow management system</p>
        </div>

        <Card>
          {/* Tab bar */}
          <div className="flex border-b">
            <button
              type="button"
              onClick={() => { setTab("signin"); reset(); }}
              className={`flex-1 py-3 text-sm font-medium transition-colors ${
                tab === "signin" ? "border-b-2 border-blue-600 text-blue-600" : "text-gray-500 hover:text-gray-700"
              }`}
            >
              Sign In
            </button>
            <button
              type="button"
              onClick={() => { setTab("demo"); reset(); }}
              className={`flex-1 py-3 text-sm font-medium transition-colors ${
                tab === "demo" ? "border-b-2 border-amber-500 text-amber-600" : "text-gray-500 hover:text-gray-700"
              }`}
            >
              Demo Accounts
            </button>
          </div>

          <CardContent className="pt-5">
            {tab === "signin" ? (
              <form onSubmit={handleSubmit} className="space-y-4">
                {/* Login / Register toggle */}
                <div className="flex rounded-lg border border-gray-200 overflow-hidden text-sm">
                  <button
                    type="button"
                    onClick={() => { setMode("login"); setSelectedRole(null); setError(""); }}
                    className={`flex-1 py-2 font-medium transition-colors ${
                      mode === "login" ? "bg-gray-900 text-white" : "bg-white text-gray-500 hover:text-gray-700"
                    }`}
                  >
                    Sign In
                  </button>
                  <button
                    type="button"
                    onClick={() => { setMode("register"); setError(""); }}
                    className={`flex-1 py-2 font-medium transition-colors ${
                      mode === "register" ? "bg-gray-900 text-white" : "bg-white text-gray-500 hover:text-gray-700"
                    }`}
                  >
                    Create Account
                  </button>
                </div>

                <div>
                  <Label htmlFor="user-id">User ID</Label>
                  <Input
                    id="user-id"
                    value={userId}
                    onChange={(e) => { setUserId(e.target.value); setError(""); }}
                    placeholder="e.g. alice"
                    maxLength={64}
                    autoComplete="username"
                  />
                </div>

                <div>
                  <Label htmlFor="password">Password</Label>
                  <Input
                    id="password"
                    type="password"
                    value={password}
                    onChange={(e) => { setPassword(e.target.value); setError(""); }}
                    placeholder={mode === "register" ? "Min. 6 characters" : "Your password"}
                    autoComplete={mode === "register" ? "new-password" : "current-password"}
                  />
                </div>

                {mode === "register" && (
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
                )}

                {error && <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded px-3 py-2">{error}</p>}

                <Button
                  type="submit"
                  className="w-full"
                  disabled={submitting || !userId || !password || (mode === "register" && !selectedRole)}
                >
                  {submitting
                    ? mode === "register" ? "Creating account…" : "Signing in…"
                    : mode === "register" ? "Create Account" : "Sign In"}
                </Button>
              </form>
            ) : (
              <div className="space-y-3">
                <p className="text-xs text-gray-500 bg-amber-50 border border-amber-200 rounded px-3 py-2">
                  Pre-seeded accounts — password is <code className="font-mono font-bold">demo</code> for all.{" "}
                  <Link href="/guided-demo" className="text-amber-700 font-medium hover:underline">
                    Run Guided Demo →
                  </Link>
                </p>
                <div className="space-y-1.5">
                  {DEMO_USERS.map((u) => (
                    <button
                      key={u.userId}
                      type="button"
                      onClick={() => quickLogin(u)}
                      disabled={demoLoading !== null}
                      className="w-full flex items-center justify-between rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm hover:border-amber-300 hover:bg-amber-50 transition-colors group disabled:opacity-50"
                    >
                      <span className="font-mono text-gray-700 group-hover:text-amber-800">
                        {demoLoading === u.userId ? "Signing in…" : u.userId}
                      </span>
                      <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${u.tagColor}`}>{u.tag}</span>
                    </button>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
