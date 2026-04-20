import { Separator } from "@/components/ui/separator";
import { StatusBadge } from "./StatusBadge";
import { WorkflowTimeline } from "./WorkflowTimeline";
import type { ApplicationDTO } from "@/lib/types";

const ACTION_LABELS: Record<string, string> = {
  SUBMIT: "Submitted",
  CANCEL: "Cancelled",
  ACCEPT: "Accepted",
  REJECT: "Rejected",
  ADJUST: "Adjustment Requested",
  APPROVE: "Approved",
  REREVIEW: "Sent for Re-review",
  RESUBMIT: "Resubmitted",
  UPDATE: "Updated",
  DELETE: "Deleted",
};

export function AppDetail({ app }: { app: ApplicationDTO }) {
  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start gap-4 justify-between">
        <div>
          <p className="text-xs text-gray-400 font-mono">{app.id}</p>
          <h2 className="text-xl font-bold mt-1">{app.license_type.replace(/_/g, " ")}</h2>
          <p className="text-sm text-gray-500">Applicant: {app.applicant_id}</p>
        </div>
        <StatusBadge status={app.status} />
      </div>

      <WorkflowTimeline status={app.status} />

      {app.notes && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-md px-4 py-3 text-sm text-yellow-900">
          <strong>Notes:</strong> {app.notes}
        </div>
      )}

      <Separator />

      {app.commodity && (
        <section>
          <h3 className="font-semibold text-gray-700 mb-2">Commodity</h3>
          <dl className="grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
            <dt className="text-gray-500">Name</dt>
            <dd>{app.commodity.name}</dd>
            <dt className="text-gray-500">Category</dt>
            <dd>{app.commodity.category}</dd>
            <dt className="text-gray-500">Description</dt>
            <dd className="col-span-1">{app.commodity.description}</dd>
          </dl>
        </section>
      )}

      {app.documents.length > 0 && (
        <section>
          <h3 className="font-semibold text-gray-700 mb-2">Documents</h3>
          <ul className="space-y-2">
            {app.documents.map((doc) => (
              <li key={doc.id} className="flex items-center justify-between text-sm bg-gray-50 rounded px-3 py-2">
                <div>
                  <p className="font-medium">{doc.name}</p>
                  <p className="text-xs text-gray-400">{doc.content_type}</p>
                </div>
                {/* Only render a link if the URL is http/https — prevents javascript: injection */}
                {/^https?:\/\//i.test(doc.url) ? (
                  <a
                    href={doc.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-600 hover:underline text-xs"
                  >
                    View
                  </a>
                ) : null}
              </li>
            ))}
          </ul>
        </section>
      )}

      {app.payment && (
        <section>
          <h3 className="font-semibold text-gray-700 mb-2">Payment</h3>
          <dl className="grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
            <dt className="text-gray-500">Amount</dt>
            <dd>{app.payment.amount} {app.payment.currency}</dd>
            <dt className="text-gray-500">Transaction ID</dt>
            <dd className="font-mono text-xs">{app.payment.transaction_id}</dd>
            <dt className="text-gray-500">Status</dt>
            <dd>{app.payment.status}</dd>
          </dl>
        </section>
      )}

      {app.history && app.history.length > 0 && (
        <section>
          <h3 className="font-semibold text-gray-700 mb-3">Audit Trail</h3>
          <ol className="relative border-l border-gray-200 space-y-4 ml-2">
            {app.history.map((entry) => (
              <li key={entry.id} className="ml-4">
                <div className="absolute -left-1.5 mt-1.5 h-3 w-3 rounded-full border border-white bg-gray-400" />
                <p className="text-sm font-medium text-gray-900">
                  {ACTION_LABELS[entry.action] ?? entry.action}
                  <span className="ml-2 text-xs font-normal text-gray-400">
                    {entry.from_status} → {entry.to_status}
                  </span>
                </p>
                <p className="text-xs text-gray-500">by {entry.actor_id} · {new Date(entry.occurred_at).toLocaleString()}</p>
                {entry.notes && (
                  <p className="text-xs text-gray-600 mt-1 italic">&ldquo;{entry.notes}&rdquo;</p>
                )}
              </li>
            ))}
          </ol>
        </section>
      )}

      <div className="text-xs text-gray-400">
        Created {new Date(app.created_at).toLocaleString()} ·
        Updated {new Date(app.updated_at).toLocaleString()}
      </div>
    </div>
  );
}
