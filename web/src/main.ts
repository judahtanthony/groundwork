import './style.css'

type State = {
  ok: boolean
  version: string
  total: number
  eligible: number
}

const root = document.querySelector<HTMLDivElement>('#app')

if (!root) {
  throw new Error('missing #app mount point')
}

root.innerHTML = `
  <main class="shell">
    <header>
      <p class="eyebrow">Local operator UI</p>
      <h1>Groundwork</h1>
      <p class="lede">The embedded SPA is ready for the root-centric operator surfaces.</p>
    </header>
    <section aria-labelledby="coordinator-heading">
      <div>
        <p class="eyebrow">Coordinator</p>
        <h2 id="coordinator-heading">Connecting…</h2>
        <p id="state-detail" class="detail" aria-live="polite">Reading same-origin coordinator state.</p>
      </div>
      <span id="state-dot" class="status" aria-hidden="true"></span>
    </section>
    <nav aria-label="Current operator surfaces">
      <a href="/">Dashboard</a>
      <a href="/tickets">Tickets</a>
      <a href="/approvals">Approvals</a>
    </nav>
  </main>
`

const heading = document.querySelector<HTMLHeadingElement>('#coordinator-heading')
const detail = document.querySelector<HTMLParagraphElement>('#state-detail')
const dot = document.querySelector<HTMLSpanElement>('#state-dot')

async function loadState() {
  try {
    const response = await fetch('/api/v1/state', { headers: { Accept: 'application/json' } })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const state = (await response.json()) as State
    if (heading) heading.textContent = state.ok ? 'Connected' : 'Unavailable'
    if (detail) detail.textContent = `${state.total} nodes · ${state.eligible} ready · gw ${state.version}`
    dot?.classList.add(state.ok ? 'status--ok' : 'status--bad')
  } catch (error) {
    if (heading) heading.textContent = 'Unavailable'
    if (detail) detail.textContent = error instanceof Error ? error.message : 'Could not read coordinator state.'
    dot?.classList.add('status--bad')
  }
}

void loadState()
