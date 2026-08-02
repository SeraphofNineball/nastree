import { reactive } from 'vue'
import type { FileTypeStat } from '../types'

const CUSTOM_COLOR_STORAGE_KEY = 'nastree-ext-colors'

function loadCustomExtColors(): Record<string, string> {
  try {
    const raw = localStorage.getItem(CUSTOM_COLOR_STORAGE_KEY)
    return raw ? JSON.parse(raw) : {}
  } catch {
    return {}
  }
}

/** User-chosen overrides for per-extension colors, keyed by extension (e.g. ".mkv").
 *  Reactive so treemap/pie/table all repaint when a color is changed. */
export const customExtColors = reactive<Record<string, string>>(loadCustomExtColors())

function persistCustomExtColors() {
  localStorage.setItem(CUSTOM_COLOR_STORAGE_KEY, JSON.stringify(customExtColors))
}

export function setExtColor(ext: string, color: string) {
  customExtColors[ext] = color
  persistCustomExtColors()
}

export function resetExtColor(ext: string) {
  delete customExtColors[ext]
  persistCustomExtColors()
}

export function resetAllExtColors() {
  for (const key of Object.keys(customExtColors)) delete customExtColors[key]
  persistCustomExtColors()
}

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
  return customExtColors[ext] ?? map.get(ext) ?? otherColor()
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
