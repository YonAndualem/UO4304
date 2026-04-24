'use server'

import { Buffer } from 'node:buffer'

const API_ORIGIN = process.env.API_ORIGIN ?? 'http://localhost:8080'

export async function getFileContent(
  key: string,
  token: string,
): Promise<{ success: boolean; data: string | null; contentType: string | null; message?: string }> {
  try {
    const url = `${API_ORIGIN}/api/documents/view?key=${encodeURIComponent(key)}`
    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${token}` },
      cache: 'no-store',
    })

    if (!res.ok) {
      return { success: false, data: null, contentType: null, message: 'Failed to fetch file' }
    }

    const contentType = res.headers.get('content-type') ?? 'application/octet-stream'
    const buffer = Buffer.from(await res.arrayBuffer())
    const base64 = buffer.toString('base64')

    return { success: true, data: `data:${contentType};base64,${base64}`, contentType }
  } catch (err) {
    console.error('getFileContent error:', err)
    return { success: false, data: null, contentType: null, message: 'Failed to fetch file' }
  }
}
