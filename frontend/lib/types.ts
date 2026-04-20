export type ApplicationStatus =
  | "PENDING"
  | "SUBMITTED"
  | "CANCELLED"
  | "ACCEPTED"
  | "REJECTED"
  | "ADJUSTED"
  | "APPROVED"
  | "REREVIEW";

export type Role = "CUSTOMER" | "REVIEWER" | "APPROVER";

export interface Identity {
  userId: string;
  role: Role;
  token: string;
}

export interface CommodityDTO {
  id: string;
  name: string;
  description: string;
  category: string;
}

export interface DocumentDTO {
  id: string;
  name: string;
  url: string;
  content_type: string;
  uploaded_at: string;
}

export interface PaymentDTO {
  id: string;
  amount: number;
  currency: string;
  transaction_id: string;
  status: string;
}

export interface HistoryEntryDTO {
  id: string;
  actor_id: string;
  action: string;
  from_status: string;
  to_status: string;
  notes: string;
  occurred_at: string;
}

export interface ApplicationDTO {
  id: string;
  license_type: string;
  applicant_id: string;
  status: ApplicationStatus;
  notes: string;
  commodity: CommodityDTO | null;
  documents: DocumentDTO[];
  payment: PaymentDTO | null;
  history: HistoryEntryDTO[];
  created_at: string;
  updated_at: string;
}

export type ReviewAction = "ACCEPT" | "REJECT" | "ADJUST";
export type ApproveAction = "APPROVE" | "REJECT" | "REREVIEW";

export interface ApiError {
  error: string;
}
