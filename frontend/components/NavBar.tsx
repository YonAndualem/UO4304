"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { useIdentity } from "@/contexts/IdentityContext";

const ROLE_HOME: Record<string, string> = {
  CUSTOMER: "/customer/applications",
  REVIEWER: "/reviewer/queue",
  APPROVER: "/approver/queue",
};

export function NavBar() {
  const { identity, logout } = useIdentity();
  const router = useRouter();

  function handleLogout() {
    logout();
    router.push("/");
  }

  if (!identity) return null;

  return (
    <header className="border-b bg-white px-4 py-3 flex items-center justify-between shadow-sm">
      <Link href={ROLE_HOME[identity.role] ?? "/"} className="flex items-center gap-2">
        <span className="font-bold text-gray-900">Trade License Portal</span>
        <span className="text-xs bg-blue-100 text-blue-700 rounded-full px-2 py-0.5 font-medium">
          {identity.role}
        </span>
      </Link>
      <div className="flex items-center gap-3 text-sm text-gray-600">
        <Link href="/test-flow" className="text-sm text-blue-600 hover:underline font-medium">
          Test Flow
        </Link>
        <span className="font-mono text-xs bg-gray-100 px-2 py-1 rounded">{identity.userId}</span>
        <Button variant="ghost" size="sm" onClick={handleLogout}>Sign out</Button>
      </div>
    </header>
  );
}
