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

export function textSecondaryColor(): string {
  return cssVar('--text-secondary') || '#52514e'
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

/** Shifts a #rrggbb color toward white (positive) or black (negative), amount in 0..1. */
function shadeColor(hex: string, amount: number): string {
  const num = parseInt(hex.replace('#', ''), 16)
  const target = amount < 0 ? 0 : 255
  const p = Math.abs(amount)
  const channel = (shift: number) => {
    const c = (num >> shift) & 0xff
    return Math.round((target - c) * p) + c
  }
  const r = channel(16)
  const g = channel(8)
  const b = channel(0)
  return `#${((1 << 24) | (r << 16) | (g << 8) | b).toString(16).slice(1)}`
}

/** A subtle top-lighter/bottom-darker vertical gradient for treemap boxes,
 *  for the glossy embossed look WizTree uses instead of flat fills. */
export function treemapGradient(hex: string) {
  return {
    type: 'linear' as const,
    x: 0,
    y: 0,
    x2: 0,
    y2: 1,
    colorStops: [
      { offset: 0, color: shadeColor(hex, 0.22) },
      { offset: 1, color: shadeColor(hex, -0.18) },
    ],
  }
}
