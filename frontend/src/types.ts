export interface TreeNode {
  path: string
  name: string
  isDir: boolean
  size: number
  files: number
  dirs: number
  ext?: string
  modTime: number
}

export interface FileTypeStat {
  ext: string
  size: number
  count: number
}

export interface ScanStatus {
  scanId: number | null
  rootPath: string
  startedAt: string
  finishedAt: string
  durationMs: number
  totalSize: number
  totalFiles: number
  totalDirs: number
  diskTotal: number
  diskFree: number
  running: boolean
  error?: string
}
