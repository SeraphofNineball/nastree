function clean(path: string): string {
  return path.replace(/[\\/]+$/, '')
}

/** Parent directory of `path`, clamped so it never rises above `root`. */
export function parentPath(path: string, root: string): string {
  const cleanRoot = clean(root)
  const cleanPath = clean(path)
  if (cleanPath === cleanRoot || !cleanPath.startsWith(cleanRoot)) return cleanRoot
  const idx = Math.max(cleanPath.lastIndexOf('/'), cleanPath.lastIndexOf('\\'))
  if (idx <= cleanRoot.length) return cleanRoot
  return cleanPath.slice(0, idx)
}
