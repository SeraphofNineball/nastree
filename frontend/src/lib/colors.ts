import type { FileTypeStat } from '../types'

function cssVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

/** Slot 1 is reserved for directories; files use slots 2-8 by global rank. */
export function seriesColor(slot: number): string {
  return cssVar(`--series-${slot}`)
}

export function directoryColor(): string {
  return seriesColor(1)
}

export function otherColor(): string {
  return cssVar('--text-muted') || '#898781'
}

export function surfaceColor(): string {
  return cssVar('--surface-2') || '#f9f9f7'
}

/** Stable ext -> color assignment: top 7 extensions by global size get their
 *  own hue, everything else folds into a shared "other" gray. */
export function buildExtColorMap(fileTypes: FileTypeStat[]): Map<string, string> {
  const map = new Map<string, string>()
  fileTypes.slice(0, 7).forEach((ft, i) => map.set(ft.ext, seriesColor(i + 2)))
  return map
}

export function colorForExt(ext: string | undefined, map: Map<string, string>): string {
  if (!ext) return otherColor()
  return map.get(ext) ?? otherColor()
}
