import { reactive } from 'vue'

/** Paths currently expanded in the DirectoryTree sidebar, shared so that
 *  navigation from other views (breadcrumb, table, treemap) keeps the tree in sync. */
export const expandedPaths = reactive(new Set<string>())

function clean(path: string): string {
  return path.replace(/[\\/]+$/, '')
}

/** Marks every ancestor of `path` (down to `root`) as expanded. */
export function expandAncestors(path: string, root: string) {
  const cleanRoot = clean(root)
  let p = clean(path)
  while (p.length >= cleanRoot.length) {
    expandedPaths.add(p)
    if (p === cleanRoot) break
    const idx = Math.max(p.lastIndexOf('/'), p.lastIndexOf('\\'))
    if (idx <= cleanRoot.length) {
      expandedPaths.add(cleanRoot)
      break
    }
    p = p.slice(0, idx)
  }
}
