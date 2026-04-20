import type { Identity, Role } from "./types";

const STORAGE_KEY = "trade_license_identity";

// Validate userId: alphanumeric, hyphens, underscores only — prevents header injection.
export function isValidUserId(id: string): boolean {
  return /^[a-zA-Z0-9_-]{1,64}$/.test(id);
}

export function saveIdentity(identity: Identity): void {
  if (typeof window === "undefined") return;
  localStorage.setItem(STORAGE_KEY, JSON.stringify(identity));
}

export function loadIdentity(): Identity | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as Identity;
    if (!isValidUserId(parsed.userId)) return null;
    const validRoles: Role[] = ["CUSTOMER", "REVIEWER", "APPROVER"];
    if (!validRoles.includes(parsed.role)) return null;
    if (!parsed.token) return null;
    return parsed;
  } catch {
    return null;
  }
}

export function clearIdentity(): void {
  if (typeof window === "undefined") return;
  localStorage.removeItem(STORAGE_KEY);
}
