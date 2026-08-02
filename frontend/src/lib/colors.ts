import { reactive, ref } from 'vue'
import type { FileTypeStat } from '../types'
import { currentTheme } from './theme'

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

const FOLDER_COLOR_STORAGE_KEY = 'nastree-folder-color'

/** User-chosen override for the folder/directory color. Null means "use the theme default". */
export const customFolderColor = ref<string | null>(localStorage.getItem(FOLDER_COLOR_STORAGE_KEY))

export function setFolderColor(color: string) {
  customFolderColor.value = color
  localStorage.setItem(FOLDER_COLOR_STORAGE_KEY, color)
}

export function resetFolderColor() {
  customFolderColor.value = null
  localStorage.removeItem(FOLDER_COLOR_STORAGE_KEY)
}

const GRADIENT_TOP_STORAGE_KEY = 'nastree-gradient-top'
const GRADIENT_BOTTOM_STORAGE_KEY = 'nastree-gradient-bottom'

/** User-chosen overrides for the treemap box gradient's highlight (top) and
 *  shadow (bottom) stops. Null means "use the computed theme default". */
export const customGradientTop = ref<string | null>(localStorage.getItem(GRADIENT_TOP_STORAGE_KEY))
export const customGradientBottom = ref<string | null>(localStorage.getItem(GRADIENT_BOTTOM_STORAGE_KEY))

export function setGradientTop(color: string) {
  customGradientTop.value = color
  localStorage.setItem(GRADIENT_TOP_STORAGE_KEY, color)
}

export function resetGradientTop() {
  customGradientTop.value = null
  localStorage.removeItem(GRADIENT_TOP_STORAGE_KEY)
}

export function setGradientBottom(color: string) {
  customGradientBottom.value = color
  localStorage.setItem(GRADIENT_BOTTOM_STORAGE_KEY, color)
}

export function resetGradientBottom() {
  customGradientBottom.value = null
  localStorage.removeItem(GRADIENT_BOTTOM_STORAGE_KEY)
}

function cssVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

/** Slot 1 is reserved for directories; files use slots 2-8 by global rank. */
export function seriesColor(slot: number): string {
  return cssVar(`--series-${slot}`)
}

export function directoryColor(): string {
  return customFolderColor.value ?? seriesColor(1)
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

/** Blends a #rrggbb color toward a target grey level (0-255), by `amount` in 0..1. */
function blendToward(hex: string, target: number, amount: number): string {
  const num = parseInt(hex.replace('#', ''), 16)
  const channel = (shift: number) => {
    const c = (num >> shift) & 0xff
    return Math.round((target - c) * amount) + c
  }
  const r = channel(16)
  const g = channel(8)
  const b = channel(0)
  return `#${((1 << 24) | (r << 16) | (g << 8) | b).toString(16).slice(1)}`
}

/** A subtle top-lighter/bottom-darker vertical gradient for treemap boxes,
 *  for the glossy embossed look WizTree uses instead of flat fills. Defaults:
 *  the shadow stop is a flat mid-grey (#808080) on the light theme - black
 *  there reads as muddy against a light page - and a near-black blend on dark
 *  themes. Either end can be overridden with a flat user-chosen color instead. */
export function treemapGradient(hex: string) {
  const isLight = currentTheme.value === 'light'
  const top = customGradientTop.value ?? blendToward(hex, 255, 0.22)
  const bottom = customGradientBottom.value ?? (isLight ? '#808080' : blendToward(hex, 0, 0.18))
  return {
    type: 'linear' as const,
    x: 0,
    y: 0,
    x2: 0,
    y2: 1,
    colorStops: [
      { offset: 0, color: top },
      { offset: 1, color: bottom },
    ],
  }
}
