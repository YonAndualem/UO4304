import type {
  ApplicationDTO,
  Identity,
  ReviewAction,
  ApproveAction,
} from "./types";

class ApiResponseError extends Error {
  constructor(public status: number, public body: string) {
    super(`API error ${status}: ${body}`);
  }
}

export { ApiResponseError };

async function request<T>(
  path: string,
  identity: Identity,
  options: RequestInit = {}
): Promise<T> {
  const res = await fetch(path, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${identity.token}`,
      ...options.headers,
    },
  });

  if (res.status === 204) return undefined as T;

  const text = await res.text();
  if (!res.ok) throw new ApiResponseError(res.status, text);

  return text ? (JSON.parse(text) as T) : (undefined as T);
}

// ── Auth (public — no identity required) ─────────────────────────────────────

export interface AuthResponse {
  token: string;
  user_id: string;
  role: string;
}

async function authRequest<T>(path: string, body: object): Promise<T> {
  const res = await fetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const text = await res.text();
  if (!res.ok) throw new ApiResponseError(res.status, text);
  return JSON.parse(text) as T;
}

export const authApi = {
  register(userId: string, password: string, role: string): Promise<{ user_id: string; role: string }> {
    return authRequest("/api/auth/register", { user_id: userId, password, role });
  },
  login(userId: string, password: string): Promise<AuthResponse> {
    return authRequest("/api/auth/login", { user_id: userId, password });
  },
};

// ── File upload / preview ────────────────────────────────────────────────────

export interface UploadedFile {
  key: string;
  name: string;
  content_type: string;
}

export const storageApi = {
  async upload(identity: Identity, file: File): Promise<UploadedFile> {
    const form = new FormData();
    form.append("file", file);
    const res = await fetch("/api/customer/upload", {
      method: "POST",
      headers: { "Authorization": `Bearer ${identity.token}` },
      body: form,
    });
    const text = await res.text();
    if (!res.ok) throw new ApiResponseError(res.status, text);
    return JSON.parse(text) as UploadedFile;
  },

};

// ── Customer ──────────────────────────────────────────────────────────────────

export interface SubmitPayload {
  license_type: string;
  commodity: { name: string; description: string; category: string };
  documents: { name: string; url: string; content_type: string }[];
  payment: { amount: number; currency: string; transaction_id: string };
}

export interface UpdatePayload {
  commodity: { name: string; description: string; category: string };
  documents: { name: string; url: string; content_type: string }[];
  /** Optional — omit to keep the existing payment unchanged. */
  payment?: { amount: number; currency: string; transaction_id: string };
}

export const customerApi = {
  async submit(identity: Identity, payload: SubmitPayload): Promise<ApplicationDTO> {
    // Backend returns {application_id: "..."} — fetch the full DTO afterwards
    const { application_id } = await request<{ application_id: string }>(
      "/api/customer/applications",
      identity,
      { method: "POST", body: JSON.stringify(payload) }
    );
    return request<ApplicationDTO>(
      `/api/customer/applications/${encodeURIComponent(application_id)}`,
      identity
    );
  },

  list(identity: Identity): Promise<ApplicationDTO[]> {
    return request<ApplicationDTO[]>("/api/customer/applications", identity);
  },

  get(identity: Identity, id: string): Promise<ApplicationDTO> {
    return request<ApplicationDTO>(`/api/customer/applications/${encodeURIComponent(id)}`, identity);
  },

  update(identity: Identity, id: string, payload: UpdatePayload): Promise<ApplicationDTO> {
    return request<ApplicationDTO>(`/api/customer/applications/${encodeURIComponent(id)}`, identity, {
      method: "PUT",
      body: JSON.stringify(payload),
    });
  },

  resubmit(identity: Identity, id: string, payload: UpdatePayload): Promise<ApplicationDTO> {
    return request<ApplicationDTO>(`/api/customer/applications/${encodeURIComponent(id)}/resubmit`, identity, {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  cancel(identity: Identity, id: string): Promise<void> {
    return request<void>(`/api/customer/applications/${encodeURIComponent(id)}/cancel`, identity, {
      method: "POST",
    });
  },

  delete(identity: Identity, id: string): Promise<void> {
    return request<void>(`/api/customer/applications/${encodeURIComponent(id)}`, identity, {
      method: "DELETE",
    });
  },
};

// ── Reviewer ─────────────────────────────────────────────────────────────────

export interface ReviewPayload {
  action: ReviewAction;
  notes: string;
}

export const reviewerApi = {
  list(identity: Identity, status?: string): Promise<ApplicationDTO[]> {
    const qs = status ? `?status=${encodeURIComponent(status)}` : "";
    return request<ApplicationDTO[]>(`/api/reviewer/applications${qs}`, identity);
  },

  get(identity: Identity, id: string): Promise<ApplicationDTO> {
    return request<ApplicationDTO>(`/api/reviewer/applications/${encodeURIComponent(id)}`, identity);
  },

  takeAction(identity: Identity, id: string, payload: ReviewPayload): Promise<void> {
    return request<void>(`/api/reviewer/applications/${encodeURIComponent(id)}/action`, identity, {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },
};

// ── Approver ──────────────────────────────────────────────────────────────────

export interface ApprovePayload {
  action: ApproveAction;
  notes: string;
}

export const approverApi = {
  list(identity: Identity, status?: string): Promise<ApplicationDTO[]> {
    const qs = status ? `?status=${encodeURIComponent(status)}` : "";
    return request<ApplicationDTO[]>(`/api/approver/applications${qs}`, identity);
  },

  get(identity: Identity, id: string): Promise<ApplicationDTO> {
    return request<ApplicationDTO>(`/api/approver/applications/${encodeURIComponent(id)}`, identity);
  },

  takeAction(identity: Identity, id: string, payload: ApprovePayload): Promise<void> {
    return request<void>(`/api/approver/applications/${encodeURIComponent(id)}/action`, identity, {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },
};
