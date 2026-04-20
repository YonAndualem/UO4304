"use client";

import { useIdentity } from "@/contexts/IdentityContext";

const ROLE_COLORS: Record<string, string> = {
  CUSTOMER: "bg-blue-600",
  REVIEWER: "bg-green-600",
  APPROVER: "bg-purple-600",
};

const ROLE_LABELS: Record<string, string> = {
  CUSTOMER: "Customer — submitting and managing applications",
  REVIEWER: "Reviewer — reviewing submitted applications",
  APPROVER: "Approver — making final approval decisions",
};

export function RoleBar() {
  const { identity } = useIdentity();

  if (!identity) return null;

  const color = ROLE_COLORS[identity.role] ?? "bg-gray-600";
  const label = ROLE_LABELS[identity.role] ?? identity.role;

  return (
    <div className={`${color} text-white text-xs font-medium px-4 py-1.5 flex items-center gap-3`}>
      <span className="uppercase tracking-wide font-bold">{identity.role}</span>
      <span className="opacity-75">·</span>
      <span className="opacity-90">{label}</span>
      <span className="opacity-75">·</span>
      <span className="font-mono opacity-75">{identity.userId}</span>
    </div>
  );
}
