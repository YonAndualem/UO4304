"use client";

import { useDemoMode } from "@/contexts/DemoModeContext";

export function DemoBanner() {
  const { isDemoMode, toggleDemoMode } = useDemoMode();

  if (!isDemoMode) return null;

  return (
    <div className="bg-amber-400 text-amber-950 text-sm font-medium px-4 py-2 flex items-center justify-between">
      <span>
        🎬 Demo Mode — you are viewing a live walkthrough of the workflow. All data is for demonstration purposes.
      </span>
      <button
        type="button"
        onClick={toggleDemoMode}
        className="text-xs underline underline-offset-2 hover:no-underline shrink-0 ml-4"
      >
        Exit demo
      </button>
    </div>
  );
}
