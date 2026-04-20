"use client";

import { useState } from "react";
import Link from "next/link";
import { useDemoMode } from "@/contexts/DemoModeContext";

const SCENARIOS = [
  { id: "happy",           label: "Happy Path",         description: "Submit → Reviewer Accepts → Approver Approves" },
  { id: "adjust_resubmit", label: "Adjust + Resubmit",  description: "Reviewer requests changes → Customer resubmits → Approved" },
  { id: "reviewer_reject", label: "Reviewer Rejects",   description: "Submit → Reviewer Rejects" },
  { id: "approver_reject", label: "Approver Rejects",   description: "Submit → Reviewer Accepts → Approver Rejects" },
  { id: "rereview",        label: "Re-review Cycle",    description: "Approver sends back to Reviewer → second pass → Approved" },
  { id: "cancel",          label: "Customer Cancels",   description: "Customer submits then cancels before review" },
];

export function DemoDrawer() {
  const { isDemoMode } = useDemoMode();
  const [open, setOpen] = useState(false);

  if (!isDemoMode) return null;

  return (
    <>
      {/* Floating trigger tab on the right edge */}
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="fixed right-0 top-1/2 -translate-y-1/2 z-40 bg-amber-400 text-amber-950 font-semibold text-xs writing-mode-vertical px-2 py-4 rounded-l-lg shadow-lg hover:bg-amber-500 transition-colors"
        style={{ writingMode: "vertical-rl", textOrientation: "mixed" }}
      >
        🎬 Scenarios
      </button>

      {/* Backdrop */}
      {open && (
        <div
          className="fixed inset-0 z-40 bg-black/30"
          onClick={() => setOpen(false)}
        />
      )}

      {/* Drawer panel */}
      <div
        className={`fixed top-0 right-0 h-full w-80 z-50 bg-white shadow-2xl flex flex-col transform transition-transform duration-300 ${
          open ? "translate-x-0" : "translate-x-full"
        }`}
      >
        <div className="bg-amber-400 text-amber-950 px-4 py-3 flex items-center justify-between">
          <span className="font-bold text-sm">🎬 Demo Scenarios</span>
          <button
            type="button"
            onClick={() => setOpen(false)}
            className="text-amber-800 hover:text-amber-950 text-lg leading-none"
          >
            ✕
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-4 space-y-2">
          <p className="text-xs text-gray-500 mb-3">
            Each scenario walks through a different workflow path. Click to open the full step-by-step runner.
          </p>
          {SCENARIOS.map((s) => (
            <Link
              key={s.id}
              href={`/guided-demo?scenario=${s.id}`}
              onClick={() => setOpen(false)}
              className="block rounded-lg border border-gray-200 bg-gray-50 hover:border-amber-300 hover:bg-amber-50 px-3 py-3 transition-colors group"
            >
              <p className="font-semibold text-sm text-gray-900 group-hover:text-amber-800">{s.label}</p>
              <p className="text-xs text-gray-500 mt-0.5">{s.description}</p>
            </Link>
          ))}
        </div>

        <div className="border-t p-4">
          <Link
            href="/guided-demo"
            onClick={() => setOpen(false)}
            className="block w-full text-center bg-amber-400 hover:bg-amber-500 text-amber-950 font-semibold text-sm py-2 rounded-lg transition-colors"
          >
            Open Guided Demo →
          </Link>
        </div>
      </div>
    </>
  );
}
