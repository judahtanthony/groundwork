export type Theme = 'light' | 'dark'

const storageKey = 'groundwork-theme'

export function preferredTheme(): Theme {
  const saved = localStorage.getItem(storageKey)
  if (saved === 'light' || saved === 'dark') return saved
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function applyTheme(theme: Theme) {
  document.documentElement.dataset.theme = theme
  document.documentElement.style.colorScheme = theme
  localStorage.setItem(storageKey, theme)
}

export function initializeTheme() {
  const theme = preferredTheme()
  applyTheme(theme)
  return theme
}
