'use client'

import { useMemo } from 'react'
import { Document, Page, pdfjs } from 'react-pdf'
import 'react-pdf/dist/Page/AnnotationLayer.css'
import 'react-pdf/dist/Page/TextLayer.css'

pdfjs.GlobalWorkerOptions.workerSrc = new URL(
    'pdfjs-dist/build/pdf.worker.min.mjs',
    import.meta.url,
).toString()

interface PdfPreviewProps {
    bytes: Uint8Array
    pageNumber: number
    numPages: number
    onLoadSuccess: (numPages: number) => void
    onLoadError: (message: string) => void
}

export function PdfPreview({ bytes, pageNumber, numPages, onLoadSuccess, onLoadError }: PdfPreviewProps) {
    const file = useMemo(() => ({ data: bytes }), [bytes])

    return (
        <Document
            file={file}
            onLoadSuccess={({ numPages }) => onLoadSuccess(numPages)}
            onLoadError={(err) => onLoadError(err.message)}
            loading={<div className="text-sm text-gray-500 py-8">Loading PDF...</div>}
        >
            {numPages > 0 && (
                <Page
                    pageNumber={pageNumber}
                    renderTextLayer
                    renderAnnotationLayer
                    className="shadow-md"
                />
            )}
        </Document>
    )
}
