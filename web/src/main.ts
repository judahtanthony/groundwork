import { Moon, Sun } from 'lucide'
import './style.css'
import { Badge, Icon, IconButton, Panel } from './design-system/components'
import { applyTheme, initializeTheme } from './design-system/theme'

type State = {
  ok: boolean
  version: string
  total: number
  eligible: number
}

const root = document.querySelector<HTMLDivElement>('#app')

if (!root) throw new Error('missing #app mount point')

let theme = initializeTheme()

const app = document.createElement('main')
app.className = 'gw gw-app'

const pageHeader = document.createElement('header')
pageHeader.className = 'gw-app-head'
const eyebrow = document.createElement('p')
eyebrow.className = 'gw-eyebrow'
eyebrow.textContent = 'Local operator UI'
const title = document.createElement('h1')
title.textContent = 'Groundwork'
const lede = document.createElement('p')
lede.className = 'lede'
lede.textContent = 'The embedded SPA is ready for root-centric operator surfaces.'
pageHeader.append(eyebrow, title, lede)

const heading = document.createElement('h2')
heading.textContent = 'Connecting…'
const detail = document.createElement('p')
detail.className = 'detail'
detail.setAttribute('aria-live', 'polite')
detail.textContent = 'Reading same-origin coordinator state.'
const stateCopy = document.createElement('div')
const stateLabel = document.createElement('p')
stateLabel.className = 'gw-eyebrow'
stateLabel.textContent = 'Coordinator'
stateCopy.append(stateLabel, heading, detail)

let stateBadge = Badge('Connecting', 'idle')
const stateRow = document.createElement('div')
stateRow.className = 'gw-state'
stateRow.append(stateCopy, stateBadge)

const nextTheme = () => theme === 'light' ? 'dark' : 'light'
const themeButton = IconButton(`Use ${nextTheme()} theme`, theme === 'light' ? Moon : Sun, () => {
  theme = theme === 'light' ? 'dark' : 'light'
  applyTheme(theme)
  themeButton.replaceChildren(Icon(theme === 'light' ? Moon : Sun))
  themeButton.title = `Use ${theme === 'light' ? 'dark' : 'light'} theme`
  themeButton.setAttribute('aria-label', themeButton.title)
})

const statePanel = Panel('System state', [stateRow], [themeButton])

const nav = document.createElement('nav')
nav.className = 'gw-app-nav'
nav.setAttribute('aria-label', 'Current operator surfaces')
const routes: ReadonlyArray<readonly [string, string]> = [
  ['Dashboard', '/'],
  ['Tickets', '/tickets'],
  ['Approvals', '/approvals'],
]
for (const [label, href] of routes) {
  const link = document.createElement('a')
  link.href = href
  link.textContent = label
  nav.append(link)
}

app.append(pageHeader, statePanel, nav)
root.replaceChildren(app)

function showState(label: string, tone: Parameters<typeof Badge>[1]) {
  const replacement = Badge(label, tone)
  stateBadge.replaceWith(replacement)
  stateBadge = replacement
}

async function loadState() {
  try {
    const response = await fetch('/api/v1/state', { headers: { Accept: 'application/json' } })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const state = (await response.json()) as State
    heading.textContent = state.ok ? 'Connected' : 'Unavailable'
    detail.textContent = `${state.total} nodes · ${state.eligible} ready · gw ${state.version}`
    showState(state.ok ? 'Live' : 'Unavailable', state.ok ? 'ok' : 'bad')
  } catch (error) {
    heading.textContent = 'Unavailable'
    detail.textContent = error instanceof Error ? error.message : 'Could not read coordinator state.'
    showState('Offline', 'bad')
  }
}

void loadState()
