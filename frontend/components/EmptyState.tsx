"use client";

import Link from "next/link";
import { useDemoMode } from "@/contexts/DemoModeContext";

interface EmptyStateProps {
  icon: string;
  title: string;
  subtitle: string;
  demoHint?: string;
  demoAction?: { label: string; href: string };
}

export function EmptyState({ icon, title, subtitle, demoHint, demoAction }: EmptyStateProps) {
  const { isDemoMode } = useDemoMode();

  return (
    <div className="flex flex-col items-center justify-center py-20 text-center space-y-3">
      <span className="text-5xl">{icon}</span>
      <p className="text-lg font-semibold text-gray-700">{title}</p>
      <p className="text-sm text-gray-400 max-w-xs">{subtitle}</p>

      {isDemoMode && demoHint && (
        <div className="mt-4 bg-amber-50 border border-amber-200 rounded-lg px-4 py-3 max-w-sm">
          <p className="text-xs text-amber-800">{demoHint}</p>
          {demoAction && (
            <Link
              href={demoAction.href}
              className="mt-2 inline-block text-xs font-semibold bg-amber-400 hover:bg-amber-500 text-amber-950 px-4 py-1.5 rounded-full transition-colors"
            >
              {demoAction.label}
            </Link>
          )}
        </div>
      )}
    </div>
  );
}
