"use client";

import { useRef, useState } from "react";
import dynamic from "next/dynamic";
import { storageApi, ApiResponseError } from "@/lib/api";
import type { Identity } from "@/lib/types";

const DocumentPreviewModal = dynamic(
  () => import("@/components/DocumentPreviewModal").then((m) => m.DocumentPreviewModal),
  { ssr: false }
);

export interface UploadedDoc {
  key: string;
  name: string;
  content_type: string;
}

interface DocumentUploadProps {
  identity: Identity;
  value: UploadedDoc | null;
  onChange: (doc: UploadedDoc) => void;
  error?: string;
}

const ACCEPTED = ".pdf,.jpg,.jpeg,.png,.webp";

export function DocumentUpload({ identity, value, onChange, error }: DocumentUploadProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState("");
  const [showPreview, setShowPreview] = useState(false);

  async function handleFile(file: File) {
    setUploadError("");
    setUploading(true);
    try {
      const uploaded = await storageApi.upload(identity, file);
      onChange(uploaded);
    } catch (e) {
      setUploadError(e instanceof ApiResponseError ? e.body : (e instanceof Error ? e.message : "Upload failed"));
    } finally {
      setUploading(false);
    }
  }

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (file) handleFile(file);
    e.target.value = "";
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault();
    const file = e.dataTransfer.files?.[0];
    if (file) handleFile(file);
  }

  function handlePreview(e: React.MouseEvent) {
    e.stopPropagation();
    if (!value) return;
    setShowPreview(true);
  }

  const displayError = uploadError || error;

  return (
    <>
      {showPreview && value && (
        <DocumentPreviewModal
          fileKey={value.key}
          token={identity.token}
          name={value.name}
          contentType={value.content_type}
          onClose={() => setShowPreview(false)}
        />
      )}

      <div className="space-y-2">
        <div
          onDrop={handleDrop}
          onDragOver={(e) => e.preventDefault()}
          onClick={() => !value && inputRef.current?.click()}
          className={`relative flex flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed px-4 py-5 transition-colors
            ${uploading ? "opacity-60 pointer-events-none" : ""}
            ${value ? "cursor-default" : "cursor-pointer hover:border-blue-400 hover:bg-blue-50/40"}
            ${displayError ? "border-red-300 bg-red-50/30" : "border-gray-300 bg-gray-50"}
          `}
        >
          <input
            ref={inputRef}
            type="file"
            accept={ACCEPTED}
            className="sr-only"
            onChange={handleChange}
          />

          {uploading ? (
            <p className="text-sm text-gray-500 animate-pulse">Uploading…</p>
          ) : value ? (
            <div className="flex items-center gap-3 w-full">
              <span className="text-2xl shrink-0">
                {value.content_type.startsWith("image/") ? "🖼" : "📄"}
              </span>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">{value.name}</p>
                <p className="text-xs text-gray-400">{value.content_type}</p>
              </div>
              <div className="flex items-center gap-2 shrink-0">
                <button
                  type="button"
                  onClick={handlePreview}
                  className="text-xs px-3 py-1 rounded-md bg-blue-50 text-blue-700 border border-blue-200 hover:bg-blue-100 transition-colors"
                >
                  Preview
                </button>
                <button
                  type="button"
                  onClick={(e) => { e.stopPropagation(); inputRef.current?.click(); }}
                  className="text-xs px-3 py-1 rounded-md bg-gray-100 text-gray-600 border border-gray-200 hover:bg-gray-200 transition-colors"
                >
                  Replace
                </button>
              </div>
            </div>
          ) : (
            <>
              <span className="text-3xl">📎</span>
              <p className="text-sm font-medium text-gray-600">Click or drag a file here</p>
              <p className="text-xs text-gray-400">PDF, JPG, PNG, WEBP — max 20 MB</p>
            </>
          )}
        </div>

        {displayError && (
          <p className="text-xs text-red-600">{displayError}</p>
        )}
      </div>
    </>
  );
}
