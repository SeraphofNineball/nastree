import { ref } from 'vue'

export interface ThemeOption {
  value: string
  label: string
}

export const THEMES: ThemeOption[] = [
  { value: 'dark', label: 'Dark' },
  { value: 'light', label: 'Light' },
  { value: 'high-contrast', label: 'High Contrast' },
  { value: 'vscode', label: 'VS Code' },
  { value: 'monokai', label: 'Monokai' },
  { value: 'solarized', label: 'Solarized' },
]

const STORAGE_KEY = 'nastree-theme'

export const currentTheme = ref(localStorage.getItem(STORAGE_KEY) || 'dark')

export function setTheme(value: string) {
  currentTheme.value = value
  document.documentElement.setAttribute('data-theme', value)
  localStorage.setItem(STORAGE_KEY, value)
}

// Apply immediately on module load so the correct theme is set before first paint.
setTheme(currentTheme.value)
