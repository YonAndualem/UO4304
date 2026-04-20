import type { ApplicationStatus } from "@/lib/types";

const STEPS: { status: ApplicationStatus; label: string }[] = [
  { status: "PENDING",   label: "Draft" },
  { status: "SUBMITTED", label: "Submitted" },
  { status: "ACCEPTED",  label: "Accepted" },
  { status: "APPROVED",  label: "Approved" },
];

const TERMINAL: ApplicationStatus[] = ["CANCELLED", "REJECTED", "ADJUSTED", "REREVIEW"];

const ORDER: ApplicationStatus[] = [
  "PENDING", "SUBMITTED", "ACCEPTED", "APPROVED",
];

function stepState(
  stepStatus: ApplicationStatus,
  current: ApplicationStatus
): "done" | "active" | "idle" {
  if (TERMINAL.includes(current)) {
    const idx = ORDER.indexOf(stepStatus);
    const curIdx = ORDER.indexOf(current);
    if (idx < curIdx) return "done";
    if (idx === curIdx) return "active";
    return "idle";
  }
  const idx = ORDER.indexOf(stepStatus);
  const curIdx = ORDER.indexOf(current);
  if (idx < curIdx) return "done";
  if (idx === curIdx) return "active";
  return "idle";
}

export function WorkflowTimeline({ status }: { status: ApplicationStatus }) {
  const isTerminalSidetrack = TERMINAL.includes(status);

  return (
    <div className="flex items-center gap-0">
      {STEPS.map((step, i) => {
        const state = stepState(step.status, status);
        return (
          <div key={step.status} className="flex items-center">
            <div className="flex flex-col items-center">
              <div
                className={[
                  "w-8 h-8 rounded-full flex items-center justify-center text-xs font-semibold border-2",
                  state === "done"   ? "bg-emerald-500 border-emerald-500 text-white" : "",
                  state === "active" ? "bg-blue-500 border-blue-500 text-white" : "",
                  state === "idle"   ? "bg-white border-gray-300 text-gray-400" : "",
                ].join(" ")}
              >
                {state === "done" ? "✓" : i + 1}
              </div>
              <span className={[
                "text-[10px] mt-1 font-medium",
                state === "done"   ? "text-emerald-600" : "",
                state === "active" ? "text-blue-600" : "",
                state === "idle"   ? "text-gray-400" : "",
              ].join(" ")}>
                {step.label}
              </span>
            </div>
            {i < STEPS.length - 1 && (
              <div
                className={[
                  "h-0.5 w-10 mx-1 mb-4",
                  ORDER.indexOf(step.status) < ORDER.indexOf(status) && !isTerminalSidetrack
                    ? "bg-emerald-400"
                    : "bg-gray-200",
                ].join(" ")}
              />
            )}
          </div>
        );
      })}
      {isTerminalSidetrack && (
        <div className="ml-3 text-xs font-medium px-2 py-0.5 rounded-full
          bg-red-100 text-red-700 border border-red-200">
          {status}
        </div>
      )}
    </div>
  );
}
