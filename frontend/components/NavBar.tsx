"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { useIdentity } from "@/contexts/IdentityContext";
import { useDemoMode } from "@/contexts/DemoModeContext";
import { Separator } from "@/components/ui/separator";

const ROLE_HOME: Record<string, string> = {
  CUSTOMER: "/customer/applications",
  REVIEWER: "/reviewer/queue",
  APPROVER: "/approver/queue",
};

export function NavBar() {
  const { identity, logout } = useIdentity();
  const { isDemoMode, toggleDemoMode } = useDemoMode();
  const router = useRouter();

  function handleLogout() {
    logout();
    router.push("/");
  }

  return (
    <header
      className={`border-b px-4 py-3 flex items-center justify-between shadow-sm transition-colors duration-300 ${
        isDemoMode ? "bg-amber-50 border-amber-300" : "bg-white border-gray-200"
      }`}
    >
      <Link
        href={identity ? (ROLE_HOME[identity.role] ?? "/") : "/"}
        className="flex items-center gap-2"
      >
        <span className="font-bold text-gray-900">Trade License Portal</span>
        {identity && (
          <span className="text-xs bg-blue-100 text-blue-700 rounded-full px-2 py-0.5 font-medium">
            {identity.role}
          </span>
        )}
      </Link>

      <div className="flex items-center gap-3">
        {/* Live / Demo pill toggle */}
        <button
          type="button"
          onClick={toggleDemoMode}
          aria-pressed={isDemoMode}
          className={`flex items-center rounded-full border text-xs font-semibold overflow-hidden transition-colors duration-200 ${
            isDemoMode ? "border-amber-400 bg-amber-100" : "border-gray-300 bg-gray-100"
          }`}
        >
          <span
            className={`px-3 py-1 rounded-full transition-colors duration-200 ${
              !isDemoMode ? "bg-white text-gray-800 shadow-sm" : "text-gray-400"
            }`}
          >
            Live
          </span>
          <span
            className={`px-3 py-1 rounded-full transition-colors duration-200 ${
              isDemoMode ? "bg-amber-400 text-white shadow-sm" : "text-gray-400"
            }`}
          >
            Demo
          </span>
        </button>

        {isDemoMode && (
          <>
            <Link
              href="/guided-demo"
              className="text-xs font-semibold text-amber-700 bg-amber-100 border border-amber-300 rounded-full px-3 py-1 hover:bg-amber-200 transition-colors"
            >
              Guided Demo
            </Link>
            <Separator orientation="vertical" className="h-5" />
          </>
        )}

        {identity && (
          <>
            <span className="font-mono text-xs bg-gray-100 px-2 py-1 rounded text-gray-600">
              {identity.userId}
            </span>
            <Button variant="ghost" size="sm" onClick={handleLogout}>
              Sign out
            </Button>
          </>
        )}
      </div>
    </header>
  );
}
