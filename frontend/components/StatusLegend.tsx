"use client";

import { useState } from "react";
import { useDemoMode } from "@/contexts/DemoModeContext";

const STATUSES = [
  { status: "PENDING",   color: "bg-gray-100 text-gray-700",     dot: "bg-gray-400",    meaning: "Saved by customer, not yet submitted for review." },
  { status: "SUBMITTED", color: "bg-blue-100 text-blue-700",     dot: "bg-blue-500",    meaning: "Submitted and waiting in the reviewer's queue." },
  { status: "ACCEPTED",  color: "bg-green-100 text-green-700",   dot: "bg-green-500",   meaning: "Reviewer approved it — now in the approver's queue." },
  { status: "ADJUSTED",  color: "bg-yellow-100 text-yellow-700", dot: "bg-yellow-500",  meaning: "Reviewer requested changes — customer must resubmit." },
  { status: "REREVIEW",  color: "bg-orange-100 text-orange-700", dot: "bg-orange-500",  meaning: "Approver sent it back to reviewer for a second look." },
  { status: "APPROVED",  color: "bg-emerald-100 text-emerald-700", dot: "bg-emerald-500", meaning: "Final approval granted — license issued." },
  { status: "REJECTED",  color: "bg-red-100 text-red-700",       dot: "bg-red-500",     meaning: "Rejected by reviewer or approver — application closed." },
  { status: "CANCELLED", color: "bg-gray-100 text-gray-500",     dot: "bg-gray-300",    meaning: "Cancelled by the customer before a final decision." },
];

export function StatusLegend() {
  const { isDemoMode } = useDemoMode();
  const [open, setOpen] = useState(false);

  if (!isDemoMode) return null;

  return (
    <div className="fixed bottom-4 left-4 z-40 w-72 shadow-xl rounded-xl overflow-hidden border border-amber-300">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="w-full bg-amber-100 hover:bg-amber-200 text-amber-900 text-xs font-semibold px-4 py-2.5 flex items-center justify-between transition-colors"
      >
        <span>📋 Status Legend</span>
        <span className="text-amber-600">{open ? "▲ hide" : "▼ show"}</span>
      </button>

      {open && (
        <div className="bg-white divide-y divide-gray-100 max-h-80 overflow-y-auto">
          {STATUSES.map((s) => (
            <div key={s.status} className="flex items-start gap-3 px-3 py-2">
              <span className={`mt-1 w-2 h-2 rounded-full shrink-0 ${s.dot}`} />
              <div>
                <span className={`inline-block text-xs font-bold px-1.5 py-0.5 rounded ${s.color}`}>
                  {s.status}
                </span>
                <p className="text-xs text-gray-500 mt-0.5">{s.meaning}</p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
