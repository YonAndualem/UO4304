'use client'

import { useEffect, useMemo, useRef, useState } from 'react'
import dynamic from 'next/dynamic'
import { getFileContent } from '@/actions/file'

const PdfPreview = dynamic(
  () => import('@/components/PdfPreview').then((m) => m.PdfPreview),
  { ssr: false }
)

interface DocumentPreviewModalProps {
  fileKey: string
  token: string
  name: string
  contentType: string
  onClose: () => void
}

export function DocumentPreviewModal({ fileKey, token, name, contentType, onClose }: DocumentPreviewModalProps) {
  const overlayRef = useRef<HTMLDivElement>(null)
  // PDFs are passed as raw bytes so pdfjs reads from memory (no network stream, no fetch).
  // Images use a data URI directly since img-src allows data: in the CSP.
  const [pdfBytes, setPdfBytes] = useState<Uint8Array | null>(null)
  const [imageDataUri, setImageDataUri] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [numPages, setNumPages] = useState(0)
  const [pageNumber, setPageNumber] = useState(1)

  useEffect(() => {
    async function load() {
      setLoading(true)
      setError(null)
      setPdfBytes(null)
      setImageDataUri(null)
      setNumPages(0)
      try {
        const result = await getFileContent(fileKey, token)
        if (!result.success || !result.data) {
          setError(result.message ?? 'Failed to load file')
          return
        }
        const [meta, b64] = result.data.split(',')
        const mime = meta.split(':')[1].split(';')[0]
        const bytes = Uint8Array.from(atob(b64), (c) => c.charCodeAt(0))
        if (mime.startsWith('image/')) {
          setImageDataUri(result.data)
        } else {
          setPdfBytes(bytes)
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load file')
      } finally {
        setLoading(false)
      }
    }

    load()
  }, [fileKey, token])

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  const isImage = contentType.startsWith('image/')
  const pdfFile = useMemo(() => pdfBytes ? { data: pdfBytes } : null, [pdfBytes])

  return (
    <div
      ref={overlayRef}
      onClick={(e) => { if (e.target === overlayRef.current) onClose() }}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
    >
      <div className="relative flex flex-col bg-white rounded-xl shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200 shrink-0">
          <div className="flex items-center gap-2 min-w-0">
            <span className="text-gray-400 text-lg">{isImage ? '🖼' : '📄'}</span>
            <p className="font-medium text-sm truncate">{name}</p>
            <span className="text-xs text-gray-400 shrink-0">{contentType}</span>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-700 text-xl leading-none px-1"
            aria-label="Close preview"
          >
            ×
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto min-h-0 bg-gray-100" style={{ minHeight: '500px' }}>
          {loading && (
            <div className="flex items-center justify-center h-full py-16 text-sm text-gray-500">
              Loading preview…
            </div>
          )}
          {error && (
            <div className="flex items-center justify-center h-full py-16 text-sm text-red-500">
              {error}
            </div>
          )}
          {!loading && !error && imageDataUri && (
            <div className="flex items-center justify-center h-full p-4">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src={imageDataUri} alt={name} className="max-w-full max-h-full object-contain rounded" />
            </div>
          )}
          {!loading && !error && pdfFile && (
            <div className="flex flex-col items-center p-4 gap-3">
              <PdfPreview
                file={pdfFile}
                pageNumber={pageNumber}
                numPages={numPages}
                onLoadSuccess={(pages) => { setNumPages(pages); setPageNumber(1) }}
                onLoadError={(message) => setError(message)}
              />
              {numPages > 1 && (
                <div className="flex items-center gap-3 text-sm text-gray-600 shrink-0">
                  <button
                    onClick={() => setPageNumber(p => Math.max(1, p - 1))}
                    disabled={pageNumber <= 1}
                    className="px-2 py-1 rounded border disabled:opacity-40 hover:bg-gray-100"
                  >
                    ‹
                  </button>
                  <span>{pageNumber} / {numPages}</span>
                  <button
                    onClick={() => setPageNumber(p => Math.min(numPages, p + 1))}
                    disabled={pageNumber >= numPages}
                    className="px-2 py-1 rounded border disabled:opacity-40 hover:bg-gray-100"
                  >
                    ›
                  </button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
