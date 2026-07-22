import {
  ArrowLeft,
  ChevronRight,
  GitBranch,
  Home,
  Moon,
  Search,
  Sun,
} from 'lucide'
import './style.css'
import {
  ActorBadge,
  Badge,
  ChecklistItem,
  DependencyBadge,
  EmptyState,
  ErrorState,
  Icon,
  IconButton,
  NodeTypeBadge,
  Panel,
  type SemanticTone,
} from './design-system/components'
import { applyTheme, initializeTheme } from './design-system/theme'

type Ticket = {
  id: string
  parent_id?: string
  kind: string
  node_type?: 'leaf' | 'composite'
  work_type?: string
  title: string
  description?: string
  status: string
  assignee?: string
  acceptance: string[]
  labels: string[]
  created_at: string
  updated_at: string
}

type State = {
  ok: boolean
  version: string
  total: number
  eligible: number
  counts: Record<string, number>
}

type BriefNode = Pick<Ticket, 'id' | 'title' | 'status' | 'node_type'>
type Decision = {
  id?: string
  event_type?: string
  statement?: string
  status?: string
  created_at?: string
}
type Brief = {
  node: BriefNode
  acceptance: string[]
  ancestor_spine: BriefNode[]
  dependencies: BriefNode[]
  pending_blockers?: Decision[]
  recent_decisions?: Decision[]
  changed_files?: string[]
  completion_summary?: unknown
}
type Dependencies = { depends_on: string[]; dependents: string[] }
type Validation = {
  id: string
  name: string
  command?: string
  status: string
  started_at?: string
  completed_at?: string
}
type Run = {
  id: string
  ticket_id: string
  actor_id: string
  mode: string
  runtime: string
  model?: string
  status: string
  started_at: string
  updated_at: string
  completed_at?: string
  last_event?: string
  last_message?: string
}
type RunEvent = {
  id: number
  run_id: string
  event_type: string
  payload: string
  created_at: string
}
type LandPreview = { staged: boolean; diff: string }

type AppData = { state: State; tickets: Ticket[]; runs: Run[] }

const mountPoint = document.querySelector<HTMLDivElement>('#app')
if (!mountPoint) throw new Error('missing #app mount point')
const mount = mountPoint

let theme = initializeTheme()
let data: AppData | undefined
let searchTerm = ''

const settledStatuses = new Set(['done', 'cancelled', 'archived'])
const runningStatuses = new Set(['running', 'starting', 'paused'])

function el<K extends keyof HTMLElementTagNameMap>(tag: K, className?: string, text?: string) {
  const node = document.createElement(tag)
  if (className) node.className = className
  if (text !== undefined) node.textContent = text
  return node
}

function button(className: string, label: string, onClick: () => void) {
  const node = el('button', className, label)
  node.type = 'button'
  node.addEventListener('click', onClick)
  return node
}

function statusTone(status: string): SemanticTone {
  if (status === 'done' || status === 'approved' || status === 'pass') return 'ok'
  if (status === 'blocked' || status === 'failed' || status === 'fail') return 'bad'
  if (status === 'in_progress' || status === 'running' || status === 'review') return 'run'
  if (status === 'todo' || status === 'landing') return 'warn'
  return 'idle'
}

function pretty(value: string) {
  return value.replaceAll('_', ' ')
}

function timeLabel(value?: string) {
  if (!value) return 'time unavailable'
  const date = new Date(value)
  return Number.isNaN(date.valueOf()) ? value : date.toLocaleString()
}

async function getJSON<T>(path: string): Promise<T> {
  const response = await fetch(path, { headers: { Accept: 'application/json' } })
  if (!response.ok) {
    let message = `HTTP ${response.status}`
    try {
      const body = await response.json() as { error?: { message?: string } }
      message = body.error?.message ?? message
    } catch {
      // Keep the status when the server did not return its JSON error envelope.
    }
    throw new Error(message)
  }
  return response.json() as Promise<T>
}

function navigate(ticket: Ticket) {
  window.location.hash = ticket.parent_id ? `#/node/${ticket.id}` : `#/root/${ticket.id}`
}

function parseRoute(): { view: 'roots' } | { view: 'node'; id: string } {
  const match = window.location.hash.match(/^#\/(?:root|node)\/([^/]+)$/)
  return match ? { view: 'node', id: decodeURIComponent(match[1]!) } : { view: 'roots' }
}

function descendantsOf(id: string, tickets: Ticket[]) {
  const byParent = new Map<string, Ticket[]>()
  for (const ticket of tickets) {
    if (!ticket.parent_id) continue
    const siblings = byParent.get(ticket.parent_id) ?? []
    siblings.push(ticket)
    byParent.set(ticket.parent_id, siblings)
  }
  const descendants: Ticket[] = []
  const pending = [...(byParent.get(id) ?? [])]
  while (pending.length) {
    const current = pending.shift()!
    descendants.push(current)
    pending.push(...(byParent.get(current.id) ?? []))
  }
  return descendants
}

function rollup(ticket: Ticket, tickets: Ticket[]) {
  const descendants = descendantsOf(ticket.id, tickets)
  const settled = descendants.filter((child) => settledStatuses.has(child.status)).length
  return { settled, total: descendants.length }
}

function appHeader(title: string, subtitle: string, showHome = false) {
  const header = el('header', 'gw-shell-head')
  const brand = el('div', 'gw-brand')
  const eyebrow = el('p', 'gw-eyebrow', 'Groundwork · local operator')
  const heading = el('h1', undefined, title)
  const detail = el('p', 'gw-page-subtitle', subtitle)
  brand.append(eyebrow, heading, detail)

  const actions = el('div', 'gw-head-actions')
  if (showHome) {
    const home = IconButton('Back to roots board', Home, () => { window.location.hash = '#/' })
    actions.append(home)
  }
  const nextTheme = theme === 'light' ? 'dark' : 'light'
  const themeButton = IconButton(`Use ${nextTheme} theme`, theme === 'light' ? Moon : Sun, () => {
    theme = theme === 'light' ? 'dark' : 'light'
    applyTheme(theme)
    void render()
  })
  actions.append(themeButton)
  header.append(brand, actions)
  return header
}

function metric(label: string, value: string, tone?: string) {
  const cell = el('div', `gw-strip-cell${tone ? ` ${tone}` : ''}`)
  cell.append(el('span', 's-l', label), el('span', 's-v', value))
  return cell
}

function renderRootsBoard(appData: AppData) {
  const app = el('main', 'gw gw-app gw-board-page')
  const roots = appData.tickets.filter((ticket) => !ticket.parent_id)
  const activeRuns = appData.runs.filter((run) => runningStatuses.has(run.status)).length
  app.append(appHeader('Roots', 'Choose a human planning handle, then drill into its agent-created work tree.'))

  const strip = el('section', 'gw-strip', undefined)
  strip.setAttribute('aria-label', 'Coordinator summary')
  strip.append(
    metric('Root nodes', String(roots.length)),
    metric('All nodes', String(appData.state.total)),
    metric('Ready', String(appData.state.eligible), appData.state.eligible ? 'warn' : undefined),
    metric('Active runs', String(activeRuns), activeRuns ? 'run' : undefined),
  )
  app.append(strip)

  const controls = el('div', 'gw-ctlbar')
  const search = el('label', 'gw-search')
  search.append(Icon(Search))
  const input = el('input')
  input.type = 'search'
  input.placeholder = 'Search roots by id or title'
  input.setAttribute('aria-label', 'Search roots')
  input.value = searchTerm
  input.addEventListener('input', () => {
    searchTerm = input.value
    renderRootRows(rows, roots, appData)
  })
  search.append(input)
  controls.append(search, el('span', 'gw-count-copy', `${roots.length} root${roots.length === 1 ? '' : 's'}`))
  app.append(controls)

  const rows = el('section', 'gw-rows')
  rows.setAttribute('aria-label', 'Root nodes')
  renderRootRows(rows, roots, appData)
  app.append(rows)
  mount.replaceChildren(app)
}

function renderRootRows(rows: HTMLElement, roots: Ticket[], appData: AppData) {
  rows.replaceChildren()
  const query = searchTerm.trim().toLowerCase()
  const shown = roots.filter((ticket) => !query || ticket.id.toLowerCase().includes(query) || ticket.title.toLowerCase().includes(query))
  if (!shown.length) {
    rows.append(EmptyState(roots.length ? 'No matching roots' : 'No roots yet', roots.length ? 'Try another id or title.' : 'Create a parentless node to establish the first human planning handle.'))
    return
  }
  for (const rootTicket of shown) {
    const progress = rollup(rootTicket, appData.tickets)
    const percent = progress.total ? Math.round((progress.settled / progress.total) * 100) : 0
    const rootRuns = new Set([rootTicket.id, ...descendantsOf(rootTicket.id, appData.tickets).map((ticket) => ticket.id)])
    const activeRuns = appData.runs.filter((run) => rootRuns.has(run.ticket_id) && runningStatuses.has(run.status)).length
    const rootType = rootTicket.node_type ?? (progress.total ? 'composite' : 'leaf')
    const row = el('button', `gw-root-row${rootTicket.status === 'blocked' || rootTicket.status === 'review' ? ' attn' : ''}`)
    row.type = 'button'
    row.addEventListener('click', () => navigate(rootTicket))
    const status = el('span', `rr-status tone-${statusTone(rootTicket.status)}`)
    status.title = pretty(rootTicket.status)
    const progressCopy = el('span', 'rr-prog')
    const bar = el('span', 'rr-bar')
    const fill = el('i')
    fill.style.width = `${percent}%`
    bar.append(fill)
    progressCopy.append(bar, `${progress.settled}/${progress.total || 0} settled`)
    const attention = el('span', 'rr-attn')
    attention.append(NodeTypeBadge(rootType), Badge(pretty(rootTicket.status), statusTone(rootTicket.status)))
    if (activeRuns) attention.append(Badge(`${activeRuns} active`, 'run'))
    row.append(status, el('span', 'rr-id', rootTicket.id), el('span', 'rr-title', rootTicket.title), progressCopy, attention, Icon(ChevronRight))
    rows.append(row)
  }
}

async function renderNode(appData: AppData, id: string) {
  const ticket = appData.tickets.find((item) => item.id === id)
  if (!ticket) {
    renderFailure('Node not found', `${id} is not present in the coordinator ticket list.`)
    return
  }

  const app = el('main', 'gw gw-app gw-node-page')
  app.append(appHeader(ticket.id, ticket.title, true))
  const loading = el('div', 'gw-loading', 'Loading node context…')
  app.append(loading)
  mount.replaceChildren(app)

  try {
    const [children, brief, dependencies, validations, preview] = await Promise.all([
      getJSON<Ticket[]>(`/api/v1/tickets/${encodeURIComponent(id)}/children`),
      getJSON<Brief>(`/api/v1/tickets/${encodeURIComponent(id)}/context`),
      getJSON<Dependencies>(`/api/v1/tickets/${encodeURIComponent(id)}/dependencies`),
      getJSON<Validation[]>(`/api/v1/tickets/${encodeURIComponent(id)}/validations`),
      getJSON<LandPreview>(`/api/v1/tickets/${encodeURIComponent(id)}/land/preview`).catch(() => ({ staged: false, diff: '' })),
    ])
    const runs = appData.runs.filter((run) => run.ticket_id === id)
    const runEvents = (await Promise.all(runs.map(async (run) => {
      const events = await getJSON<RunEvent[]>(`/api/v1/runs/${encodeURIComponent(run.id)}/events`).catch(() => [])
      return events
    }))).flat()
    loading.remove()
    app.append(
      renderBreadcrumbs(brief, appData.tickets),
      renderSpine(ticket, children, brief, dependencies, appData),
      renderFocusedDetail(ticket, brief, validations, runs, runEvents, preview),
    )
  } catch (error) {
    loading.replaceWith(ErrorState('Unable to load node', error instanceof Error ? error.message : 'Unknown coordinator error'))
  }
}

function renderBreadcrumbs(brief: Brief, tickets: Ticket[]) {
  const nav = el('nav', 'gw-breadcrumbs')
  nav.setAttribute('aria-label', 'Node ancestors')
  const rootButton = button('gw-crumb home', 'Roots', () => { window.location.hash = '#/' })
  rootButton.prepend(Icon(Home))
  nav.append(rootButton)
  for (const ancestor of brief.ancestor_spine) {
    nav.append(Icon(ChevronRight))
    const full = tickets.find((ticket) => ticket.id === ancestor.id)
    nav.append(button('gw-crumb', ancestor.id, () => full ? navigate(full) : undefined))
  }
  nav.append(Icon(ChevronRight), el('span', 'gw-crumb current', brief.node.id))
  return nav
}

function renderSpine(ticket: Ticket, children: Ticket[], brief: Brief, dependencies: Dependencies, appData: AppData) {
  const spine = el('section', 'gw-spine')
  spine.setAttribute('aria-label', 'Node hierarchy')

  const ancestors = el('div', 'gw-anc')
  ancestors.append(el('p', 'gw-anc-label', 'Up · ancestors'))
  if (!brief.ancestor_spine.length) {
    ancestors.append(el('p', 'gw-quiet', 'This node is a root — the top-level human planning handle.'))
  } else {
    for (const ancestor of brief.ancestor_spine) {
      const full = appData.tickets.find((item) => item.id === ancestor.id)
      const item = button('gw-anc-item', ancestor.id, () => full ? navigate(full) : undefined)
      item.append(el('span', 'gw-anc-title', ancestor.title))
      ancestors.append(item)
    }
  }

  const here = el('article', 'gw-node-card')
  const head = el('div', 'gw-node-card-head')
  const type = ticket.node_type ?? (children.length ? 'composite' : 'leaf')
  head.append(el('span', 'gw-id', ticket.id), Badge(pretty(ticket.status), statusTone(ticket.status)), NodeTypeBadge(type))
  const title = el('h2', 'gw-node-title', ticket.title)
  const meta = el('div', 'gw-node-meta')
  meta.append(
    ActorBadge(ticket.parent_id ? 'agent branch / leaf' : 'human root handle', ticket.parent_id ? 'ai' : 'human'),
    el('span', 'gw-kind', ticket.work_type || ticket.kind),
  )
  const depRow = el('div', 'gw-node-deps')
  if (dependencies.depends_on.length) depRow.append(DependencyBadge(String(dependencies.depends_on.length), 'blocked-by'))
  if (dependencies.dependents.length) depRow.append(DependencyBadge(String(dependencies.dependents.length), 'blocks'))
  const progress = rollup(ticket, appData.tickets)
  if (type === 'composite') depRow.append(Badge(`${progress.settled}/${progress.total} descendants settled`, progress.total > 0 && progress.settled === progress.total ? 'ok' : 'idle', false))
  here.append(head, title, meta, depRow)
  if (dependencies.depends_on.length || dependencies.dependents.length) {
    const edges = el('div', 'gw-edge-list')
    const addEdges = (ids: string[], label: string) => {
      for (const depID of ids) {
        const target = appData.tickets.find((item) => item.id === depID)
        const edge = button('gw-edge', `${label} ${depID}${target ? ` · ${target.title}` : ''}`, () => target ? navigate(target) : undefined)
        edge.prepend(Icon(GitBranch))
        edges.append(edge)
      }
    }
    addEdges(dependencies.depends_on, 'Waits on')
    addEdges(dependencies.dependents, 'Unblocks')
    here.append(edges)
  }

  const down = el('div', 'gw-children')
  const columnHead = el('div', 'gw-col-head')
  columnHead.append(el('span', 'ch-name', 'Down · children'), el('span', 'ch-count', String(children.length)))
  down.append(columnHead)
  if (!children.length) {
    down.append(EmptyState('No child nodes', type === 'leaf' ? 'This leaf is one verifiable change.' : 'This composite has not been broken down yet.'))
  } else {
    for (const [index, child] of children.entries()) {
      const childDescendants = descendantsOf(child.id, appData.tickets)
      const childType = child.node_type ?? (childDescendants.length ? 'composite' : 'leaf')
      const item = button(`gw-child${child.parent_id ? ' prov' : ''}`, '', () => navigate(child))
      item.replaceChildren()
      const order = el('span', 'gw-child-order', String(index + 1))
      const body = el('span', 'gw-child-body')
      const top = el('span', 'gw-child-top')
      top.append(el('span', 'gw-child-id', child.id), Badge(pretty(child.status), statusTone(child.status)))
      body.append(top, el('span', 'gw-child-title', child.title))
      const childMeta = el('span', 'gw-child-meta')
      childMeta.append(NodeTypeBadge(childType), ActorBadge('agent-created work', 'ai'))
      if (childType === 'composite') {
        const done = childDescendants.filter((node) => settledStatuses.has(node.status)).length
        childMeta.append(Badge(`${done}/${childDescendants.length} settled`, done === childDescendants.length && done > 0 ? 'ok' : 'idle', false))
      }
      body.append(childMeta)
      item.append(order, body, Icon(ChevronRight))
      down.append(item)
    }
  }
  spine.append(ancestors, here, down)
  return spine
}

function renderFocusedDetail(ticket: Ticket, brief: Brief, validations: Validation[], runs: Run[], events: RunEvent[], preview: LandPreview) {
  const detail = el('section', 'gw-detail')
  const heading = el('div', 'gw-detail-head')
  heading.append(el('div', undefined, undefined), Badge('Focused node', 'idle', false))
  heading.firstElementChild!.append(el('p', 'gw-eyebrow', 'Detail follows focus'), el('h2', undefined, `${ticket.id} evidence`))
  detail.append(heading)

  const grid = el('div', 'gw-detail-grid')
  grid.append(
    Panel('Problem', [textBlock(ticket.description, 'No problem statement recorded.')]),
    Panel('Acceptance', [acceptanceList(ticket.acceptance.length ? ticket.acceptance : brief.acceptance)]),
    Panel('Validations', [validationList(validations)]),
    Panel('Runs', [runList(runs)]),
    Panel('Diff', [diffView(preview, brief.changed_files ?? [])]),
    Panel('Timeline', [timeline(ticket, validations, runs, events, brief)]),
  )
  detail.append(grid)
  return detail
}

function textBlock(value: string | undefined, fallback: string) {
  return el('p', value ? 'gw-copy' : 'gw-quiet', value || fallback)
}

function acceptanceList(items: string[]) {
  if (!items.length) return el('p', 'gw-quiet', 'No acceptance criteria recorded.')
  const list = el('div', 'gw-checklist')
  for (const item of items) list.append(ChecklistItem(item, false))
  return list
}

function validationList(validations: Validation[]) {
  if (!validations.length) return el('p', 'gw-quiet', 'No validation results recorded.')
  const list = el('div', 'gw-evidence-list')
  for (const validation of validations) {
    const row = el('div', 'gw-evidence-row')
    const copy = el('div')
    copy.append(el('strong', undefined, validation.name), el('span', 'gw-row-detail', validation.command || timeLabel(validation.completed_at || validation.started_at)))
    row.append(copy, Badge(pretty(validation.status), statusTone(validation.status)))
    list.append(row)
  }
  return list
}

function runList(runs: Run[]) {
  if (!runs.length) return el('p', 'gw-quiet', 'No runs recorded for this node.')
  const list = el('div', 'gw-evidence-list')
  for (const run of runs) {
    const row = el('div', 'gw-evidence-row')
    const copy = el('div')
    copy.append(el('strong', 'gw-mono', run.id), el('span', 'gw-row-detail', `${run.actor_id} · ${run.mode} · ${timeLabel(run.updated_at)}`))
    row.append(copy, Badge(pretty(run.status), statusTone(run.status)))
    list.append(row)
  }
  return list
}

function diffView(preview: LandPreview, changedFiles: string[]) {
  const wrap = el('div', 'gw-diff-view')
  if (changedFiles.length) {
    const files = el('div', 'gw-file-list')
    for (const file of changedFiles) files.append(el('div', 'gw-diff-file', file))
    wrap.append(files)
  }
  if (preview.staged && preview.diff) {
    const pre = el('pre', 'gw-diff-code', preview.diff)
    wrap.append(pre)
  } else if (!changedFiles.length) {
    wrap.append(el('p', 'gw-quiet', 'No changed files or staged landing diff recorded.'))
  } else {
    wrap.append(el('p', 'gw-quiet', 'Changed files are recorded; no staged landing diff is available.'))
  }
  return wrap
}

type TimelineItem = { time?: string; label: string; tone: SemanticTone }

function timeline(ticket: Ticket, validations: Validation[], runs: Run[], events: RunEvent[], brief: Brief) {
  const items: TimelineItem[] = [
    { time: ticket.created_at, label: 'Node created', tone: 'idle' },
    { time: ticket.updated_at, label: `Node updated · ${pretty(ticket.status)}`, tone: statusTone(ticket.status) },
  ]
  for (const run of runs) items.push({ time: run.started_at, label: `${run.id} started by ${run.actor_id}`, tone: 'run' })
  for (const event of events) items.push({ time: event.created_at, label: `${event.run_id} · ${pretty(event.event_type)}`, tone: event.event_type.includes('fail') ? 'bad' : 'run' })
  for (const validation of validations) items.push({ time: validation.completed_at || validation.started_at, label: `${validation.name} · ${pretty(validation.status)}`, tone: statusTone(validation.status) })
  for (const decision of [...(brief.pending_blockers ?? []), ...(brief.recent_decisions ?? [])]) {
    items.push({ time: decision.created_at, label: decision.statement || pretty(decision.event_type || 'Decision recorded'), tone: decision.status === 'pending' ? 'warn' : 'idle' })
  }
  items.sort((a, b) => (Date.parse(b.time || '') || 0) - (Date.parse(a.time || '') || 0))
  const list = el('div', 'gw-tl')
  for (const item of items) {
    const row = el('div', 'gw-tl-item')
    const rail = el('div', 'gw-tl-rail')
    rail.append(el('span', `gw-tl-node ${item.tone}`), el('span', 'gw-tl-line'))
    const body = el('div', 'gw-tl-body')
    body.append(el('p', 'gw-tl-text', item.label), el('p', 'gw-tl-time', timeLabel(item.time)))
    row.append(rail, body)
    list.append(row)
  }
  return list
}

function renderFailure(title: string, message: string) {
  const app = el('main', 'gw gw-app')
  app.append(appHeader(title, 'The embedded SPA could not render this route.', true), ErrorState(title, message))
  mount.replaceChildren(app)
}

async function loadAppData() {
  const [state, tickets, runs] = await Promise.all([
    getJSON<State>('/api/v1/state'),
    getJSON<Ticket[]>('/api/v1/tickets'),
    getJSON<Run[]>('/api/v1/runs'),
  ])
  return { state, tickets, runs }
}

async function render() {
  if (!data) {
    const app = el('main', 'gw gw-app')
    app.append(appHeader('Groundwork', 'Connecting to the coordinator…'), el('div', 'gw-loading', 'Loading roots and runs…'))
    mount.replaceChildren(app)
    try {
      data = await loadAppData()
    } catch (error) {
      renderFailure('Coordinator unavailable', error instanceof Error ? error.message : 'Could not load coordinator state.')
      return
    }
  }
  const route = parseRoute()
  if (route.view === 'node') await renderNode(data, route.id)
  else renderRootsBoard(data)
}

window.addEventListener('hashchange', () => { void render() })
void render()
