import { Badge } from "@/components/ui/badge";
import type { ApplicationStatus } from "@/lib/types";

const statusConfig: Record<
  ApplicationStatus,
  { label: string; className: string }
> = {
  PENDING:   { label: "Pending",   className: "bg-gray-100 text-gray-700 border-gray-200" },
  SUBMITTED: { label: "Submitted", className: "bg-blue-100 text-blue-700 border-blue-200" },
  CANCELLED: { label: "Cancelled", className: "bg-gray-100 text-gray-500 border-gray-200" },
  ACCEPTED:  { label: "Accepted",  className: "bg-green-100 text-green-700 border-green-200" },
  REJECTED:  { label: "Rejected",  className: "bg-red-100 text-red-700 border-red-200" },
  ADJUSTED:  { label: "Adjusted",  className: "bg-yellow-100 text-yellow-700 border-yellow-200" },
  APPROVED:  { label: "Approved",  className: "bg-emerald-100 text-emerald-700 border-emerald-200" },
  REREVIEW:  { label: "Re-Review", className: "bg-orange-100 text-orange-700 border-orange-200" },
};

interface Props {
  status: ApplicationStatus;
}

export function StatusBadge({ status }: Props) {
  const cfg = statusConfig[status] ?? { label: status, className: "bg-gray-100 text-gray-600" };
  return (
    <Badge variant="outline" className={cfg.className}>
      {cfg.label}
    </Badge>
  );
}
