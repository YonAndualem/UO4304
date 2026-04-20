import Link from "next/link";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { StatusBadge } from "./StatusBadge";
import type { ApplicationDTO } from "@/lib/types";

interface Props {
  app: ApplicationDTO;
  href: string;
}

export function AppCard({ app, href }: Props) {
  const needsAttention = app.status === "ADJUSTED";
  return (
    <Link href={href}>
      <Card className={`hover:shadow-md transition-shadow cursor-pointer ${needsAttention ? "border-orange-400 bg-orange-50" : ""}`}>
        <CardHeader className="pb-2">
          <div className="flex items-start justify-between gap-2">
            <div>
              <p className="text-xs text-gray-500 font-mono">{app.id}</p>
              <p className="font-semibold mt-0.5">{app.license_type.replace(/_/g, " ")}</p>
              {needsAttention && (
                <p className="text-xs font-semibold text-orange-600 mt-0.5">⚠ Action required — adjustment requested</p>
              )}
            </div>
            <StatusBadge status={app.status} />
          </div>
        </CardHeader>
        <CardContent className="text-sm text-gray-600 space-y-1">
          {app.commodity && (
            <p><span className="font-medium">Commodity:</span> {app.commodity.name}</p>
          )}
          <p><span className="font-medium">Applicant:</span> {app.applicant_id}</p>
          {app.notes && (
            <p className="text-xs bg-yellow-50 border border-yellow-100 rounded px-2 py-1 text-yellow-800">
              {app.notes}
            </p>
          )}
          <p className="text-xs text-gray-400 pt-1">
            {new Date(app.created_at).toLocaleDateString()} · {app.documents.length} doc(s)
          </p>
        </CardContent>
      </Card>
    </Link>
  );
}
