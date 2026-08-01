import type { TreeNode, FileTypeStat, ScanStatus } from './types'

async function getJSON<T>(url: string): Promise<T> {
  const res = await fetch(url)
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `${res.status} ${res.statusText}`)
  }
  return res.json()
}

export const api = {
  status: () => getJSON<ScanStatus | { running: boolean; scanId: null }>('/api/status'),
  node: (path: string) => getJSON<TreeNode>(`/api/node?path=${encodeURIComponent(path)}`),
  children: (path: string) => getJSON<TreeNode[]>(`/api/children?path=${encodeURIComponent(path)}`),
  tree: (path: string) => getJSON<TreeNode>(`/api/tree?path=${encodeURIComponent(path)}`),
  fileTypes: () => getJSON<FileTypeStat[]>('/api/filetypes'),
  triggerScan: async () => {
    const res = await fetch('/api/scan/trigger', { method: 'POST' })
    if (!res.ok && res.status !== 409) {
      throw new Error(`trigger failed: ${res.status}`)
    }
    return res.status === 202
  },
}
