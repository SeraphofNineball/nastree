import type { TreeNode, FileTypeStat, ScanStatus } from './types'

async function getJSON<T>(url: string): Promise<T> {
  const res = await fetch(url)
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `${res.status} ${res.statusText}`)
  }
  return res.json()
}

export interface FileSearchParams {
  q?: string
  matchPath?: boolean
  foldersOnly?: boolean
  duplicatesOnly?: boolean
  dupMode?: 'name_size' | 'name_size_date'
  limit?: number
}

export const api = {
  status: () => getJSON<ScanStatus | { running: boolean; scanId: null }>('/api/status'),
  node: (path: string) => getJSON<TreeNode>(`/api/node?path=${encodeURIComponent(path)}`),
  children: (path: string) => getJSON<TreeNode[]>(`/api/children?path=${encodeURIComponent(path)}`),
  tree: (path: string) => getJSON<TreeNode>(`/api/tree?path=${encodeURIComponent(path)}`),
  fileTypes: () => getJSON<FileTypeStat[]>('/api/filetypes'),
  files: (params: FileSearchParams) => {
    const qs = new URLSearchParams()
    if (params.q) qs.set('q', params.q)
    if (params.matchPath) qs.set('matchPath', 'true')
    if (params.foldersOnly) qs.set('foldersOnly', 'true')
    if (params.duplicatesOnly) qs.set('duplicatesOnly', 'true')
    if (params.dupMode) qs.set('dupMode', params.dupMode)
    if (params.limit) qs.set('limit', String(params.limit))
    return getJSON<TreeNode[]>(`/api/files?${qs.toString()}`)
  },
  triggerScan: async () => {
    const res = await fetch('/api/scan/trigger', { method: 'POST' })
    if (!res.ok && res.status !== 409) {
      throw new Error(`trigger failed: ${res.status}`)
    }
    return res.status === 202
  },
}
