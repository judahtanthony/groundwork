import {
  Activity,
  ArrowLeft,
  ChevronRight,
  FolderGit2,
  GitBranch,
  Home,
  Link,
  ListTodo,
  LockKeyhole,
  Moon,
  Pencil,
  Plus,
  Search,
  Settings2,
  ShieldCheck,
  Split,
  Sun,
} from 'lucide'
import './style.css'
import {
  ActorBadge,
  Badge,
  Button,
  ChecklistItem,
  DependencyBadge,
  EmptyState,
  ErrorState,
  Icon,
  IconButton,
  NodeTypeBadge,
  Panel,
  TextInput,
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
  requested_actor?: string
  priority?: number
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
type CompletionSummary = { outcome: string; changed?: string[] }
type Brief = {
  node: BriefNode
  acceptance: string[]
  ancestor_spine: BriefNode[]
  dependencies: BriefNode[]
  parent_contract?: string
  sops: string[]
  open_escalations: string[]
  pending_blockers?: Decision[]
  recent_decisions?: Decision[]
  changed_files?: string[]
  completion_summary?: CompletionSummary
  summary_stale?: boolean
  summary_stale_reason?: string
  summary_missing?: boolean
}
type Blocker = { id: string; status: string }
type BlockedTicket = Ticket & { blocked_by: Blocker[] }
type Readiness = {
  next: { ticket: Ticket; brief: Brief } | null
  ready: Ticket[]
  blocked: BlockedTicket[]
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
  workspace_path?: string
  base_commit?: string
  status: string
  started_at: string
  updated_at: string
  completed_at?: string
  last_event?: string
  last_message?: string

  input_tokens: number
  output_tokens: number
  total_tokens: number
}
type RunEvent = {
  id: number
  run_id: string
  event_type: string
  message?: string
  payload: string
  created_at: string
}
type Approval = {
  id: string
  run_id?: string
  ticket_id: string
  type: string
  risk_class: string
  risk_score?: number
  reversible?: boolean
  summary: string
  action_json: string
  status: string
  requested_by_actor: string
  decided_by_actor?: string
  required_actors: string[]
  required_roles: string[]
  decision_reason?: string
  created_at: string
  decided_at?: string
}
type RunDetail = Run & {
  plan: RunEvent[]
  changed_files: string[]
  validations: Validation[]
  approval?: Approval
  cost?: number
}
type LandPreview = { id: string; staged: boolean; diff: string }
type TrustMatch = Record<string, unknown>
type TrustRule = { id: string; description?: string; when: TrustMatch; actions?: string[]; require_roles?: string[] }
type TrustPolicy = {
  schema: string
  require_human: TrustRule[]
  auto_approve: TrustRule[]
  allow_claim: TrustRule[]
}
type TrustGroup = 'require_human' | 'auto_approve' | 'allow_claim'
type PolicyRuleView = { group: TrustGroup; order: number; rule: TrustRule }
type ValidationTemplate = {
  name: string
  template: { match: { files?: string[] }; required: Array<{ name: string; command?: string }>; landing_risk_floor?: string }
}
type Policies = { trust?: TrustPolicy; rules: PolicyRuleView[]; validation_templates: ValidationTemplate[]; warnings: string[] }
type PolicySuggestion = {
  id: string; kind: string; action_type: string; work_type: string; level: string
  rationale: string; status: string; created_at: string
}
type AgentsMDStatus = { path: string; state: 'missing' | 'out_of_sync' | 'synced'; detail: string }
type Settings = {
  repository_path: string
  sqlite_path: string
  config_path: string
  server: { address: string; bind: string; port: string }
  agent: { engine: string; model?: string; sandbox: string }
  concurrency: { max: number; lease_ttl: string; lease_heartbeat: string }
  agents_md: AgentsMDStatus
}
type DoctorCheck = { name: string; status: 'ok' | 'warn' | 'error'; detail: string }
type DoctorReport = { healthy: boolean; checks: DoctorCheck[] }

type AppData = { state: State; tickets: Ticket[]; runs: Run[]; readiness: Readiness; approvals: Approval[] }

const mountPoint = document.querySelector<HTMLDivElement>('#app')
if (!mountPoint) throw new Error('missing #app mount point')
const mount = mountPoint

let theme = initializeTheme()
let data: AppData | undefined
let searchTerm = ''
let selectedPolicyRuleID = ''

const settledStatuses = new Set(['done', 'cancelled', 'archived'])
const runningStatuses = new Set(['running', 'starting', 'paused'])
const transitions: Record<string, string[]> = {
  backlog: ['todo', 'cancelled'],
  todo: ['backlog', 'in_progress', 'blocked', 'cancelled'],
  in_progress: ['blocked', 'review', 'done', 'cancelled'],
  blocked: ['todo', 'in_progress', 'cancelled'],
  review: ['approved', 'rework', 'cancelled'],
  rework: ['in_progress', 'review', 'cancelled'],
  approved: ['landing', 'cancelled'],
  landing: ['done', 'blocked', 'cancelled'],
  done: [],
  cancelled: [],
}

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

async function requestJSON<T>(path: string, method: string, body?: unknown): Promise<T> {
  const response = await fetch(path, {
    method,
    headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  if (!response.ok) {
    let message = `HTTP ${response.status}`
    try {
      const payload = await response.json() as { error?: { message?: string } }
      message = payload.error?.message ?? message
    } catch {
      // Preserve the HTTP status for non-JSON failures.
    }
    throw new Error(message)
  }
  return response.json() as Promise<T>
}

async function refresh() {
  data = await loadAppData()
  await render()
}

function field(label: string, control: HTMLElement, hint?: string) {
  const wrap = el('label', 'gw-field')
  wrap.append(el('span', 'gw-field-label', label), control)
  if (hint) wrap.append(el('span', 'gw-field-hint', hint))
  return wrap
}

function selectInput(label: string, values: string[], selected = '', includeBlank?: string) {
  const select = el('select', 'gw-input')
  select.setAttribute('aria-label', label)
  if (includeBlank !== undefined) {
    const option = el('option', undefined, includeBlank)
    option.value = ''
    select.append(option)
  }
  for (const value of values) {
    const option = el('option', undefined, pretty(value))
    option.value = value
    option.selected = value === selected
    select.append(option)
  }
  return select
}

function openForm(
  title: string,
  description: string,
  build: (form: HTMLFormElement) => void,
  submit: (values: FormData) => Promise<void>,
) {
  const dialog = el('dialog', 'gw-dialog')
  const form = el('form', 'gw-form')
  form.method = 'dialog'
  form.append(el('p', 'gw-eyebrow', 'Focused node action'), el('h2', undefined, title), el('p', 'gw-page-subtitle', description))
  build(form)
  const error = el('p', 'gw-form-error')
  error.setAttribute('role', 'alert')
  const actions = el('div', 'gw-form-actions')
  const cancel = Button('Cancel', { variant: 'ghost', onClick: () => dialog.close() })
  const save = Button('Apply', { variant: 'primary', type: 'submit' })
  actions.append(cancel, save)
  form.append(error, actions)
  form.addEventListener('submit', (event) => {
    event.preventDefault()
    save.disabled = true
    error.textContent = ''
    void submit(new FormData(form)).then(async () => {
      dialog.close()
      await refresh()
    }).catch((cause) => {
      error.textContent = cause instanceof Error ? cause.message : 'The coordinator rejected the action.'
      save.disabled = false
    })
  })
  dialog.append(form)
  dialog.addEventListener('close', () => dialog.remove())
  document.body.append(dialog)
  dialog.showModal()
  const first = dialog.querySelector<HTMLElement>('input, textarea, select')
  first?.focus()
}

function namedInput(name: string, options: Parameters<typeof TextInput>[0]) {
  const input = TextInput(options)
  input.name = name
  return input
}

function values(value: FormDataEntryValue | null) {
  return String(value ?? '').trim()
}

function lines(value: FormDataEntryValue | null) {
  return String(value ?? '').split('\n').map((item) => item.trim()).filter(Boolean)
}

function navigate(ticket: Ticket) {
  window.location.hash = ticket.parent_id ? `#/node/${ticket.id}` : `#/root/${ticket.id}`
}

function parseRoute(): { view: 'roots' } | { view: 'readiness' } | { view: 'policies' } | { view: 'settings' } | { view: 'approvals' } | { view: 'approval'; id: string } | { view: 'node'; id: string } | { view: 'run'; id: string } {
  if (window.location.hash === '#/ready') return { view: 'readiness' }
  if (window.location.hash === '#/approvals') return { view: 'approvals' }
  if (window.location.hash === '#/policies') return { view: 'policies' }
  if (window.location.hash === '#/settings') return { view: 'settings' }
  const approvalMatch = window.location.hash.match(/^#\/approval\/([^/]+)$/)
  if (approvalMatch) return { view: 'approval', id: decodeURIComponent(approvalMatch[1]!) }
  const runMatch = window.location.hash.match(/^#\/run\/([^/]+)$/)
  if (runMatch) return { view: 'run', id: decodeURIComponent(runMatch[1]!) }
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
  actions.append(Button('Next & ready', {
    variant: parseRoute().view === 'readiness' ? 'primary' : 'ghost',
    icon: ListTodo,
    onClick: () => { window.location.hash = '#/ready' },
  }))
  const pending = data?.approvals.length ?? 0
  actions.append(Button(`Approvals${pending ? ` (${pending})` : ''}`, {
    variant: parseRoute().view === 'approvals' || parseRoute().view === 'approval' ? 'primary' : 'ghost',
    icon: LockKeyhole,
    onClick: () => { window.location.hash = '#/approvals' },
  }))
  actions.append(Button('Policies', {
    variant: parseRoute().view === 'policies' ? 'primary' : 'ghost',
    icon: ShieldCheck,
    onClick: () => { window.location.hash = '#/policies' },
  }))
  actions.append(Button('Settings', {
    variant: parseRoute().view === 'settings' ? 'primary' : 'ghost',
    icon: Settings2,
    onClick: () => { window.location.hash = '#/settings' },
  }))
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
  controls.append(
    search,
    el('span', 'gw-count-copy', `${roots.length} root${roots.length === 1 ? '' : 's'}`),
    Button('New root', { variant: 'primary', icon: Plus, onClick: () => openCreateForm() }),
  )
  app.append(controls)

  const rows = el('section', 'gw-rows')
  rows.setAttribute('aria-label', 'Root nodes')
  renderRootRows(rows, roots, appData)
  app.append(rows)
  mount.replaceChildren(app)
}

function openCreateForm(parent?: Ticket) {
  openForm(
    parent ? `New child of ${parent.id}` : 'New root',
    parent ? 'Create one scoped work node below the current focus.' : 'Create a parentless human planning handle.',
    (form) => {
      form.append(
        field('Title', namedInput('title', { label: 'Title', placeholder: 'One clear outcome' })),
        field('Description', namedInput('description', { label: 'Description', multiline: true, placeholder: 'Problem or scope' })),
      )
      const row = el('div', 'gw-form-row')
      const kind = namedInput('kind', { label: 'Kind', value: 'ticket' })
      const workType = namedInput('work_type', { label: 'Work type', placeholder: 'technical_implementation' })
      row.append(field('Kind', kind), field('Work type', workType))
      const status = selectInput('Initial status', ['backlog', 'todo'], 'backlog')
      status.name = 'status'
      const priority = namedInput('priority', { label: 'Priority', placeholder: '0.0 – 1.0' })
      form.append(row, field('Initial status', status), field('Priority', priority, 'Optional value from 0 to 1.'),
        field('Labels', namedInput('labels', { label: 'Labels', multiline: true, placeholder: 'One label per line' })),
        field('Acceptance', namedInput('acceptance', { label: 'Acceptance', multiline: true, placeholder: 'One criterion per line' })))
    },
    async (form) => {
      const priorityText = values(form.get('priority'))
      const payload: Partial<Ticket> = {
        title: values(form.get('title')),
        kind: values(form.get('kind')) || 'ticket',
        parent_id: parent?.id,
        status: values(form.get('status')) || 'backlog',
        description: values(form.get('description')),
        work_type: values(form.get('work_type')),
        labels: lines(form.get('labels')),
        acceptance: lines(form.get('acceptance')),
      }
      if (priorityText) payload.priority = Number(priorityText)
      const created = await requestJSON<Ticket>('/api/v1/tickets', 'POST', payload)
      navigate(created)
    },
  )
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

function renderReadiness(appData: AppData) {
  const app = el('main', 'gw gw-app gw-readiness-page')
  const snapshot = appData.readiness
  app.append(appHeader('Next & ready', 'Take the highest-value eligible node or inspect what is waiting on dependencies.', true))

  const strip = el('section', 'gw-strip')
  strip.setAttribute('aria-label', 'Readiness summary')
  strip.append(
    metric('Recommended', snapshot.next ? snapshot.next.ticket.id : 'None'),
    metric('Ready', String(snapshot.ready.length), snapshot.ready.length ? 'warn' : undefined),
    metric('Blocked', String(snapshot.blocked.length), snapshot.blocked.length ? 'bad' : undefined),
  )
  app.append(strip)

  if (snapshot.next) app.append(renderNextRecommendation(snapshot.next, appData))
  else app.append(EmptyState('Nothing is ready', 'No todo nodes currently have all dependencies satisfied.'))

  const columns = el('div', 'gw-readiness-columns')
  columns.append(renderReadyList(snapshot.ready), renderBlockedList(snapshot.blocked, appData.tickets))
  app.append(columns)
  mount.replaceChildren(app)
}

function renderNextRecommendation(next: NonNullable<Readiness['next']>, appData: AppData) {
  const section = el('section', 'gw-next')
  const head = el('div', 'gw-next-head')
  const title = el('div')
  title.append(el('p', 'gw-eyebrow', 'Top recommendation'), el('h2', undefined, `${next.ticket.id} · ${next.ticket.title}`))
  const actions = el('div', 'gw-next-actions')
  const claim = Button('Claim as human.owner', {
    variant: 'primary',
    onClick: () => void claimTicket(next.ticket, section, claim),
  })
  actions.append(
    Badge(pretty(next.ticket.status), statusTone(next.ticket.status)),
    NodeTypeBadge(next.ticket.node_type ?? 'leaf'),
    Button('Open node', { onClick: () => navigate(next.ticket) }),
    claim,
  )
  head.append(title, actions)

  const brief = next.brief
  const grid = el('div', 'gw-next-brief')
  grid.append(
    Panel('Acceptance criteria', [acceptanceList(brief.acceptance)]),
    Panel('Ancestor spine', [briefNodes(brief.ancestor_spine, appData.tickets, 'Root node — no ancestors.')]),
    Panel('Parent contract', [textBlock(brief.parent_contract, 'No parent contract recorded.')]),
    Panel('Dependencies', [briefNodes(brief.dependencies, appData.tickets, 'No dependencies.')]),
    Panel('SOPs', [stringList(brief.sops, 'No matching SOPs.')]),
    Panel('Open escalations', [stringList(brief.open_escalations, 'No open escalations.')]),
  )
  if (brief.pending_blockers?.length) {
    grid.append(Panel('Pending blockers', [stringList(brief.pending_blockers.map((item) => item.statement || pretty(item.event_type || 'Pending decision')), 'No pending blockers.')]))
  }
  if (brief.completion_summary) {
    const summary = brief.completion_summary
    const summaryCopy = `${summary.outcome} · ${summary.changed?.length ?? 0} changed file(s)${brief.summary_stale ? ` · Stale: ${brief.summary_stale_reason || 'summary no longer matches the node'}` : ''}`
    grid.append(Panel('Completion summary', [el('p', brief.summary_stale ? 'gw-action-error' : 'gw-copy', summaryCopy)]))
  } else if (brief.summary_missing) {
    grid.append(Panel('Completion summary', [el('p', 'gw-action-error', 'No completion summary is recorded for this review/done node.')]))
  }
  section.append(head, grid)
  return section
}

function briefNodes(nodes: BriefNode[], tickets: Ticket[], fallback: string) {
  if (!nodes.length) return el('p', 'gw-quiet', fallback)
  const list = el('div', 'gw-brief-list')
  for (const node of nodes) {
    const full = tickets.find((ticket) => ticket.id === node.id)
    const row = button('gw-brief-node', '', () => full ? navigate(full) : undefined)
    row.replaceChildren(el('span', 'gw-mono', node.id), el('span', undefined, node.title), Badge(pretty(node.status), statusTone(node.status)))
    list.append(row)
  }
  return list
}

function stringList(items: string[], fallback: string) {
  if (!items.length) return el('p', 'gw-quiet', fallback)
  const list = el('ul', 'gw-string-list')
  for (const item of items) list.append(el('li', undefined, item))
  return list
}

function renderReadyList(ready: Ticket[]) {
  const panel = el('section', 'gw-ready-panel gw-work-list')
  const head = el('div', 'gw-work-list-head')
  head.append(el('div', undefined, undefined), Badge(`${ready.length} eligible`, ready.length ? 'warn' : 'idle', false))
  head.firstElementChild!.append(el('p', 'gw-eyebrow', 'Value ordered'), el('h2', undefined, 'Ready nodes'))
  panel.append(head)
  if (!ready.length) {
    panel.append(EmptyState('No ready nodes', 'Todo nodes appear here once every dependency is satisfied.'))
    return panel
  }
  for (const [index, ticket] of ready.entries()) {
    const row = el('div', 'gw-work-row')
    const rank = el('span', 'gw-work-rank', String(index + 1))
    const copy = el('button', 'gw-work-copy')
    copy.type = 'button'
    copy.addEventListener('click', () => navigate(ticket))
    copy.append(el('span', 'gw-mono', ticket.id), el('strong', undefined, ticket.title), el('span', 'gw-row-detail', `Priority ${ticket.priority ?? 0} · ${pretty(ticket.work_type || ticket.kind)}`))
    const claim = Button('Claim', { size: 'small', onClick: () => void claimTicket(ticket, row, claim) })
    row.append(rank, copy, Badge(pretty(ticket.status), statusTone(ticket.status)), claim)
    panel.append(row)
  }
  return panel
}

function renderBlockedList(blocked: BlockedTicket[], tickets: Ticket[]) {
  const panel = el('section', 'gw-blocked-panel gw-work-list')
  const head = el('div', 'gw-work-list-head')
  head.append(el('div', undefined, undefined), Badge(`${blocked.length} waiting`, blocked.length ? 'bad' : 'idle', false))
  head.firstElementChild!.append(el('p', 'gw-eyebrow', 'Unmet dependencies'), el('h2', undefined, 'Blocked nodes'))
  panel.append(head)
  if (!blocked.length) {
    panel.append(EmptyState('No dependency-blocked nodes', 'Todo nodes with unmet dependencies will appear here.'))
    return panel
  }
  for (const ticket of blocked) {
    const row = el('div', 'gw-blocked-row')
    const copy = el('button', 'gw-work-copy')
    copy.type = 'button'
    copy.addEventListener('click', () => navigate(ticket))
    copy.append(el('span', 'gw-mono', ticket.id), el('strong', undefined, ticket.title))
    const blockers = el('div', 'gw-blockers')
    blockers.append(Icon(LockKeyhole), el('span', 'gw-blocked-label', 'Blocked by'))
    for (const blocker of ticket.blocked_by) {
      const target = tickets.find((item) => item.id === blocker.id)
      blockers.append(button('gw-blocker', `${blocker.id} · ${pretty(blocker.status)}`, () => target ? navigate(target) : undefined))
    }
    row.append(copy, blockers)
    panel.append(row)
  }
  return panel
}

async function claimTicket(ticket: Ticket, container: HTMLElement, control: HTMLButtonElement) {
  control.disabled = true
  try {
    await requestJSON(`/api/v1/tickets/${encodeURIComponent(ticket.id)}/claim`, 'POST', { actor: 'human.owner' })
    await refresh()
  } catch (error) {
    control.disabled = false
    showActionError(container, error)
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
      getJSON<LandPreview>(`/api/v1/tickets/${encodeURIComponent(id)}/land/preview`).catch(() => ({ id, staged: false, diff: '' })),
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
      renderActionRail(ticket, dependencies, appData.tickets),
      renderFocusedDetail(ticket, brief, validations, runs, runEvents, preview),
    )
  } catch (error) {
    loading.replaceWith(ErrorState('Unable to load node', error instanceof Error ? error.message : 'Unknown coordinator error'))
  }
}

function renderActionRail(ticket: Ticket, dependencies: Dependencies, tickets: Ticket[]) {
  const rail = el('section', 'gw-action-rail')
  const copy = el('div', 'gw-action-copy')
  copy.append(el('p', 'gw-eyebrow', 'Act on focus'), el('h2', undefined, 'Node actions'))
  const actions = el('div', 'gw-action-buttons')
  const unmet = dependencies.depends_on
    .map((id) => tickets.find((item) => item.id === id))
    .filter((item) => item && item.status !== 'done')
  const claim = Button('Claim', {
    variant: 'primary',
    disabled: ticket.status !== 'todo' || unmet.length > 0,
    onClick: () => void requestJSON(`/api/v1/tickets/${encodeURIComponent(ticket.id)}/claim`, 'POST', { actor: 'human.owner' }).then(refresh).catch((error) => showActionError(rail, error)),
  })
  actions.append(
    claim,
    Button('Edit', { icon: Pencil, onClick: () => openEditForm(ticket) }),
    Button('Transition', { onClick: () => openTransitionForm(ticket) }),
    Button('Dependencies', { icon: Link, onClick: () => openDependencyForm(ticket, dependencies, tickets) }),
    Button('Reparent', { onClick: () => openReparentForm(ticket, tickets) }),
    Button('Triage', { onClick: () => openTriageForm(ticket) }),
    Button('Add child', { icon: Plus, onClick: () => openCreateForm(ticket) }),
    Button('Request breakdown', { icon: Split, onClick: () => openDecomposeForm(ticket) }),
    Button('Request replan', { variant: 'ghost', onClick: () => openEscalateForm(ticket) }),
  )
  const hint = el('p', 'gw-action-hint')
  if (ticket.status !== 'todo') hint.textContent = `Claim is unavailable while this node is ${pretty(ticket.status)}.`
  else if (unmet.length) hint.textContent = `Claim is waiting on ${unmet.map((item) => item!.id).join(', ')}.`
  rail.append(copy, actions, hint)
  return rail
}

function showActionError(container: HTMLElement, cause: unknown) {
  let error = container.querySelector<HTMLElement>('.gw-action-error')
  if (!error) {
    error = el('p', 'gw-action-error')
    error.setAttribute('role', 'alert')
    container.append(error)
  }
  error.textContent = cause instanceof Error ? cause.message : 'The coordinator rejected the action.'
}

function openEditForm(ticket: Ticket) {
  openForm(`Edit ${ticket.id}`, 'Update mutable node metadata. Status and parentage have dedicated guarded actions.', (form) => {
    form.append(
      field('Title', namedInput('title', { label: 'Title', value: ticket.title })),
      field('Description', namedInput('description', { label: 'Description', value: ticket.description, multiline: true })),
    )
    const row = el('div', 'gw-form-row')
    row.append(
      field('Kind', namedInput('kind', { label: 'Kind', value: ticket.kind })),
      field('Work type', namedInput('work_type', { label: 'Work type', value: ticket.work_type })),
    )
    const owner = el('div', 'gw-form-row')
    owner.append(
      field('Assignee', namedInput('assignee', { label: 'Assignee', value: ticket.assignee })),
      field('Requested actor', namedInput('requested_actor', { label: 'Requested actor', value: ticket.requested_actor })),
    )
    form.append(row, owner,
      field('Priority', namedInput('priority', { label: 'Priority', value: ticket.priority?.toString() ?? '', placeholder: '0.0 – 1.0' })),
      field('Labels', namedInput('labels', { label: 'Labels', value: ticket.labels.join('\n'), multiline: true })),
      field('Acceptance', namedInput('acceptance', { label: 'Acceptance', value: ticket.acceptance.join('\n'), multiline: true })))
  }, async (form) => {
    const priorityText = values(form.get('priority'))
    await requestJSON<Ticket>(`/api/v1/tickets/${encodeURIComponent(ticket.id)}`, 'PATCH', {
      ...ticket,
      title: values(form.get('title')),
      description: values(form.get('description')),
      kind: values(form.get('kind')),
      work_type: values(form.get('work_type')),
      assignee: values(form.get('assignee')),
      requested_actor: values(form.get('requested_actor')),
      priority: priorityText ? Number(priorityText) : null,
      labels: lines(form.get('labels')),
      acceptance: lines(form.get('acceptance')),
    })
  })
}

function openTransitionForm(ticket: Ticket) {
  const allowed = transitions[ticket.status] ?? []
  openForm(`Transition ${ticket.id}`, `Choose a lifecycle edge valid from ${pretty(ticket.status)}.`, (form) => {
    const status = selectInput('New status', allowed)
    status.name = 'status'
    status.required = true
    form.append(field('New status', status, allowed.length ? undefined : 'This node has no outgoing lifecycle transitions.'))
  }, async (form) => {
    await requestJSON(`/api/v1/tickets/${encodeURIComponent(ticket.id)}/transition`, 'POST', { status: values(form.get('status')) })
  })
}

function openTriageForm(ticket: Ticket) {
  openForm(`Triage ${ticket.id}`, 'Classify structure before execution: a leaf is one verifiable change; a composite needs children.', (form) => {
    const type = selectInput('Node type', ['leaf', 'composite'], ticket.node_type)
    type.name = 'node_type'
    form.append(field('Node type', type))
  }, async (form) => {
    await requestJSON(`/api/v1/tickets/${encodeURIComponent(ticket.id)}/triage`, 'POST', { node_type: values(form.get('node_type')) })
  })
}

function openDependencyForm(ticket: Ticket, dependencies: Dependencies, tickets: Ticket[]) {
  const candidates = tickets.filter((item) => item.id !== ticket.id).map((item) => item.id)
  openForm(`Link dependency for ${ticket.id}`, 'The focused node will wait for the selected node. Cycle checks run on the server.', (form) => {
    if (dependencies.depends_on.length) {
      const current = el('div', 'gw-linked-list')
      current.append(el('span', 'gw-field-label', 'Current dependencies'))
      for (const id of dependencies.depends_on) {
        const row = el('div', 'gw-linked-row')
        row.append(el('span', 'gw-mono', id), Button('Remove', {
          variant: 'danger',
          size: 'small',
          onClick: () => void requestJSON(`/api/v1/tickets/${encodeURIComponent(ticket.id)}/dependencies/${encodeURIComponent(id)}`, 'DELETE')
            .then(() => { form.closest('dialog')?.close(); return refresh() })
            .catch((error) => showActionError(form, error)),
        }))
        current.append(row)
      }
      form.append(current)
    }
    const target = selectInput('Depends on', candidates, '', 'Choose a node')
    target.name = 'depends_on'
    target.required = true
    form.append(field('Depends on', target))
  }, async (form) => {
    await requestJSON(`/api/v1/tickets/${encodeURIComponent(ticket.id)}/dependencies`, 'POST', { depends_on: values(form.get('depends_on')) })
  })
}

function openReparentForm(ticket: Ticket, tickets: Ticket[]) {
  const candidates = tickets.filter((item) => item.id !== ticket.id).map((item) => item.id)
  openForm(`Reparent ${ticket.id}`, 'Move this node beneath another node. Parent-cycle checks run on the server.', (form) => {
    const parent = selectInput('New parent', candidates, ticket.parent_id, 'No parent · make root')
    parent.name = 'parent'
    form.append(field('New parent', parent))
  }, async (form) => {
    await requestJSON(`/api/v1/tickets/${encodeURIComponent(ticket.id)}/reparent`, 'POST', { parent: values(form.get('parent')) })
  })
}

function openDecomposeForm(ticket: Ticket) {
  openForm(`Request breakdown for ${ticket.id}`, 'Propose child nodes through the human-gated decomposition flow.', (form) => {
    form.append(
      field('Child titles', namedInput('children', { label: 'Child titles', multiline: true, placeholder: 'One child title per line' })),
      field('Parent contract JSON', namedInput('contract', { label: 'Parent contract JSON', multiline: true, value: '{}' })),
    )
  }, async (form) => {
    const children = lines(form.get('children')).map((title) => ({ title }))
    if (!children.length) throw new Error('Add at least one child title.')
    let contract: unknown
    try {
      contract = JSON.parse(values(form.get('contract')) || '{}')
    } catch {
      throw new Error('Parent contract must be valid JSON.')
    }
    await requestJSON(`/api/v1/tickets/${encodeURIComponent(ticket.id)}/decompose`, 'POST', { contract, children })
  })
}

function openEscalateForm(ticket: Ticket) {
  openForm(`Request replan for ${ticket.id}`, 'Escalation opens a human-gated re-plan decision and propagates revision upward.', (form) => {
    form.append(field('Reason', namedInput('reason', { label: 'Reason', multiline: true, placeholder: 'What changed or needs reconsideration?' })))
  }, async (form) => {
    await requestJSON(`/api/v1/tickets/${encodeURIComponent(ticket.id)}/escalate`, 'POST', { reason: values(form.get('reason')) })
  })
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
    const row = button('gw-evidence-row gw-run-link', '', () => {
      window.location.hash = `#/run/${encodeURIComponent(run.id)}`
    })
    row.replaceChildren()
    const copy = el('div')
    copy.append(el('strong', 'gw-mono', run.id), el('span', 'gw-row-detail', `${run.actor_id} · ${run.mode} · ${timeLabel(run.updated_at)}`))
    row.append(copy, Badge(pretty(run.status), statusTone(run.status)))
    list.append(row)
  }
  return list
}

async function renderRun(id: string) {
  const app = el('main', 'gw gw-app gw-run-page')
  app.append(appHeader(id, 'Run transcript, evidence, metrics, and controls.', true))
  const loading = el('div', 'gw-loading', 'Loading run evidence…')
  app.append(loading)
  mount.replaceChildren(app)

  try {
    const [run, events] = await Promise.all([
      getJSON<RunDetail>(`/api/v1/runs/${encodeURIComponent(id)}`),
      getJSON<RunEvent[]>(`/api/v1/runs/${encodeURIComponent(id)}/events`),
    ])
    loading.remove()
    app.append(renderRunControls(run), renderRunEvidence(run, events))
  } catch (error) {
    loading.replaceWith(ErrorState('Unable to load run', error instanceof Error ? error.message : 'Unknown coordinator error'))
  }
}

function renderRunControls(run: RunDetail) {
  const rail = el('section', 'gw-action-rail gw-run-actions')
  const copy = el('div', 'gw-action-copy')
  copy.append(el('p', 'gw-eyebrow', 'Live control'), el('h2', undefined, `${run.id} · ${pretty(run.status)}`))
  const actions = el('div', 'gw-action-buttons')
  const error = el('p', 'gw-action-error')
  error.setAttribute('role', 'alert')
  const control = (op: 'pause' | 'resume' | 'cancel') => async () => {
    const buttons = [...actions.querySelectorAll('button')]
    const disabled = buttons.map((item) => item.disabled)
    for (const item of buttons) item.disabled = true
    error.textContent = ''
    try {
      await requestJSON<Run>(`/api/v1/runs/${encodeURIComponent(run.id)}/${op}`, 'POST')
      await refresh()
    } catch (cause) {
      error.textContent = cause instanceof Error ? cause.message : 'The coordinator rejected the run control.'
      buttons.forEach((item, index) => { item.disabled = disabled[index]! })
    }
  }
  actions.append(
    Button('Pause', { disabled: run.status !== 'running', onClick: () => { void control('pause')() } }),
    Button('Resume', { variant: 'primary', disabled: run.status !== 'paused' && run.status !== 'interrupted', onClick: () => { void control('resume')() } }),
    Button('Cancel', { variant: 'danger', disabled: run.status === 'completed' || run.status === 'cancelled', onClick: () => { void control('cancel')() } }),
  )
  const hint = el('p', 'gw-action-hint', 'Controls use the same coordinator lifecycle gates as gw run pause, resume, and cancel.')
  rail.append(copy, actions, hint, error)
  return rail
}

function renderRunEvidence(run: RunDetail, events: RunEvent[]) {
  const detail = el('section', 'gw-detail')
  const heading = el('div', 'gw-detail-head')
  const title = el('div')
  title.append(el('p', 'gw-eyebrow', 'Durable run evidence'), el('h2', undefined, `${run.ticket_id} · ${run.mode}`))
  heading.append(title, Badge(pretty(run.status), statusTone(run.status)))
  detail.append(heading)

  const metadata = el('div', 'gw-run-metadata')
  metadata.append(
    ActorBadge(`${run.actor_id}${run.model ? ` · ${run.model}` : ''}`, 'ai'),
    el('span', 'gw-mono gw-quiet', run.runtime),
    el('span', 'gw-mono gw-quiet', run.workspace_path || 'No workspace recorded'),
  )

  const metrics = el('div', 'gw-metric-list')
  for (const [label, value] of [
    ['Input tokens', run.input_tokens.toLocaleString()],
    ['Output tokens', run.output_tokens.toLocaleString()],
    ['Total tokens', run.total_tokens.toLocaleString()],
    ['Cost', run.cost === undefined ? 'Not reported' : `$${run.cost.toFixed(4)}`],
  ]) {
    const item = el('div', 'gw-metric-item')
    item.append(el('span', 'gw-field-label', label), el('strong', 'gw-mono', value))
    metrics.append(item)
  }

  const grid = el('div', 'gw-detail-grid gw-run-grid')
  grid.append(
    Panel('Run', [metadata, runFactList(run)]),
    Panel('Token & cost metrics', [metrics]),
    Panel('Plan', [runEventList(run.plan, 'No plan events recorded in the run transcript.')]),
    Panel('Changed files', [changedFileList(run.changed_files)]),
    Panel('Validations', [validationList(run.validations)]),
    Panel('Linked approval', [approvalView(run.approval)]),
    Panel('Transcript', [runEventList(events, 'No transcript events have been recorded.')]),
  )
  detail.append(grid)
  return detail
}

function runFactList(run: Run) {
  const facts = el('div', 'gw-run-facts')
  for (const [label, value] of [
    ['Started', timeLabel(run.started_at)],
    ['Updated', timeLabel(run.updated_at)],
    ['Completed', run.completed_at ? timeLabel(run.completed_at) : 'Not completed'],
    ['Base commit', run.base_commit || 'Not recorded'],
  ]) {
    const row = el('div', 'gw-scope-row')
    row.append(el('span', 'sk', label), el('span', 'sv', value))
    facts.append(row)
  }
  return facts
}

function runEventList(events: RunEvent[], empty: string) {
  if (!events.length) return el('p', 'gw-quiet', empty)
  const list = el('div', 'gw-tl gw-run-transcript')
  for (const event of events) {
    const row = el('div', 'gw-tl-item')
    const rail = el('div', 'gw-tl-rail')
    const tone = event.event_type.includes('fail') || event.event_type.includes('error') ? 'bad' : event.event_type.includes('checkpoint') ? 'ok' : 'run'
    rail.append(el('span', `gw-tl-node ${tone}`), el('span', 'gw-tl-line'))
    const body = el('div', 'gw-tl-body')
    const message = event.message || (event.payload && event.payload !== '{}' ? event.payload : 'No message')
    body.append(el('p', 'gw-tl-text', message), el('p', 'gw-tl-time', `${pretty(event.event_type)} · ${timeLabel(event.created_at)}`))
    row.append(rail, body)
    list.append(row)
  }
  return list
}

function changedFileList(files: string[]) {
  if (!files.length) return el('p', 'gw-quiet', 'No changed files recorded for this run.')
  const list = el('div', 'gw-file-list')
  for (const file of files) list.append(el('div', 'gw-diff-file', file))
  return list
}

function approvalView(approval: Approval | undefined) {
  if (!approval) return el('p', 'gw-quiet', 'No approval is linked to this run.')
  const wrap = el('div', 'gw-evidence-list')
  const row = el('div', 'gw-evidence-row')
  const copy = el('div')
  copy.append(el('strong', 'gw-mono', approval.id), el('span', 'gw-row-detail', `${pretty(approval.type)} · ${approval.summary}`))
  row.append(copy, Badge(pretty(approval.status), statusTone(approval.status)))
  wrap.append(row, el('p', 'gw-quiet', `${pretty(approval.risk_class)} risk · requested ${timeLabel(approval.created_at)}`))
  return wrap
}

function renderApprovalsInbox(appData: AppData) {
  const app = el('main', 'gw gw-app gw-approvals-page')
  app.append(appHeader('Approvals', 'One inbox for pending envelopes, landings, decisions, and policy exceptions.', true))

  const approvals = [...appData.approvals].sort((a, b) => {
    const exceptionRank = Number(b.type === 'exception') - Number(a.type === 'exception')
    if (exceptionRank) return exceptionRank
    return Date.parse(a.created_at) - Date.parse(b.created_at)
  })
  const highRisk = approvals.filter((approval) => approval.risk_class === 'high' || approval.risk_class === 'critical').length
  const strip = el('section', 'gw-strip')
  strip.setAttribute('aria-label', 'Approval inbox summary')
  strip.append(
    metric('Pending', String(approvals.length), approvals.length ? 'warn' : undefined),
    metric('High risk', String(highRisk), highRisk ? 'bad' : undefined),
    metric('Landing', String(approvals.filter((approval) => approval.type === 'land_to_main').length)),
    metric('Envelope', String(approvals.filter((approval) => approval.type === 'approve_envelope').length)),
  )
  app.append(strip)

  const list = el('section', 'gw-approval-list')
  list.setAttribute('aria-label', 'Pending approvals')
  if (!approvals.length) {
    list.append(EmptyState('No pending approvals', 'New gate requests will appear here.'))
  } else {
    for (const approval of approvals) list.append(approvalInboxRow(approval, appData))
  }
  app.append(list)
  mount.replaceChildren(app)
}

async function renderPolicies() {
  const app = el('main', 'gw gw-app gw-policies-page')
  app.append(appHeader('Policies', 'Ordered trust rules, validation gates, and human-reviewed policy learning.', true))
  const loading = el('div', 'gw-loading', 'Loading policy canon…')
  app.append(loading)
  mount.replaceChildren(app)
  try {
    const [policies, suggestions] = await Promise.all([
      getJSON<Policies>('/api/v1/policies'),
      getJSON<PolicySuggestion[]>('/api/v1/policies/suggestions'),
    ])
    loading.remove()
    if (!selectedPolicyRuleID || !policies.rules.some((item) => item.rule.id === selectedPolicyRuleID)) {
      selectedPolicyRuleID = policies.rules[0]?.rule.id ?? ''
    }
    const layout = el('div', 'gw-policy-layout')
    const primary = el('div', 'gw-policy-primary')
    primary.append(renderTrustRules(policies), renderValidationTemplates(policies.validation_templates))
    const rail = el('aside', 'gw-policy-rail')
    rail.append(renderRuleEditor(policies), renderSuggestionQueue(suggestions))
    layout.append(primary, rail)
    app.append(layout)
  } catch (error) {
    loading.replaceWith(ErrorState('Unable to load policies', error instanceof Error ? error.message : 'Unknown coordinator error'))
  }
}

function renderTrustRules(policies: Policies) {
  const list = el('div', 'gw-policy-list')
  for (const item of policies.rules) {
    const row = button(`gw-policy-row${item.rule.id === selectedPolicyRuleID ? ' selected' : ''}`, '', () => {
      selectedPolicyRuleID = item.rule.id
      void renderPolicies()
    })
    const copy = el('div', 'gw-policy-copy')
    copy.append(el('span', 'gw-id', item.rule.id), el('strong', undefined, item.rule.description || summarizeMatch(item.rule.when)))
    copy.append(el('span', 'gw-row-detail gw-mono', summarizeMatch(item.rule.when)))
    row.append(copy, Badge(pretty(item.group), item.group === 'require_human' ? 'warn' : item.group === 'auto_approve' ? 'ok' : 'run'))
    list.append(row)
  }
  if (!policies.rules.length) list.append(EmptyState('No trust rules', 'Add rules to .groundwork/policies/trust.yaml to establish explicit policy.'))
  return Panel('Trust rules · evaluated top-down', [list], [el('span', 'gw-id', `${policies.rules.length} rules · stable ids`)])
}

function summarizeMatch(match: TrustMatch) {
  const entries = Object.entries(match)
  if (!entries.length) return 'matches every action'
  return entries.map(([key, value]) => `${pretty(key)}: ${Array.isArray(value) ? value.join(', ') : String(value)}`).join(' · ')
}

function renderValidationTemplates(templates: ValidationTemplate[]) {
  const section = el('section', 'gw-validation-section')
  section.append(el('p', 'gw-sect-label', 'Validation templates · by file type'))
  const cards = el('div', 'gw-validation-grid')
  for (const item of templates) {
    const body = el('div', 'gw-validation-card-body')
    const globs = item.template.match.files ?? []
    body.append(el('div', 'gw-mono gw-row-detail', globs.length ? globs.join(', ') : 'All files'))
    if (item.template.required.length) {
      for (const check of item.template.required) {
        const line = el('div', 'gw-validation-check')
        line.append(Icon(ShieldCheck), el('span', undefined, check.command || check.name))
        body.append(line)
      }
    } else body.append(el('p', 'gw-quiet', 'No command checks required.'))
    if (item.template.landing_risk_floor) body.append(Badge(`risk floor · ${item.template.landing_risk_floor}`, 'idle'))
    cards.append(Panel(pretty(item.name), [body]))
  }
  if (!templates.length) cards.append(EmptyState('No validation templates', 'No file-type validation policy is configured.'))
  section.append(cards)
  return section
}

function renderRuleEditor(policies: Policies) {
  const selected = policies.rules.find((item) => item.rule.id === selectedPolicyRuleID)
  if (!selected || !policies.trust) return Panel('Rule editor', [el('p', 'gw-quiet', 'Select a trust rule to inspect it.')])
  const form = el('form', 'gw-rule-form')
  const id = namedInput('id', { label: 'Stable rule id', value: selected.rule.id, readOnly: true })
  const ruleJSON = namedInput('rule', { label: 'Rule JSON', value: JSON.stringify(selected.rule, null, 2), multiline: true })
  const ticketID = namedInput('ticket_id', { label: 'Change ticket', placeholder: 'T-0000' })
  const notice = el('p', 'gw-form-notice')
  notice.setAttribute('role', 'status')
  form.append(
    field('Stable rule id', id, 'IDs and top-down order cannot change in this editor.'),
    field('Rule · JSON', ruleJSON, 'Edit description, match conditions, actions, or reviewer requirements. The stable id must remain unchanged.'),
    field('Change ticket', ticketID, 'Policy amendments are attached to work and require human approval.'),
  )
  const controls = el('div', 'gw-rule-controls')
  const save = Button('Request amendment', { variant: 'primary', type: 'submit' })
  controls.append(save, Button('Reset', { onClick: () => void renderPolicies() }))
  form.append(controls, notice)
  form.addEventListener('submit', (event) => {
    event.preventDefault()
    void (async () => {
      save.disabled = true
      notice.className = 'gw-form-notice'
      notice.textContent = ''
      try {
        const edited = JSON.parse(ruleJSON.value) as TrustRule
        if (edited.id !== selected.rule.id) throw new Error(`Rule id must remain ${selected.rule.id}.`)
        const next = structuredClone(policies.trust!)
        const rule = next[selected.group][selected.order - 1]
        if (!rule || rule.id !== selected.rule.id) throw new Error('The selected rule changed; reload the policy surface.')
        next[selected.group][selected.order - 1] = edited
        const result = await requestJSON<{ approval: Approval }>('/api/v1/policies', 'PUT', { ticket_id: ticketID.value.trim(), trust: next })
        notice.className = 'gw-form-notice ok'
        notice.textContent = `Amendment queued as ${result.approval.id}. Approve it in the inbox to apply trust.yaml.`
      } catch (error) {
        notice.className = 'gw-form-notice bad'
        notice.textContent = error instanceof Error ? error.message : 'The coordinator rejected the amendment.'
      } finally {
        save.disabled = false
      }
    })()
  })
  return Panel('Rule editor', [form], [el('span', 'gw-id', selected.rule.id)])
}

function renderSuggestionQueue(suggestions: PolicySuggestion[]) {
  const list = el('div', 'gw-suggestion-list')
  for (const item of suggestions) {
    const card = el('article', 'gw-suggestion')
    card.append(el('strong', undefined, `${pretty(item.action_type)} · ${pretty(item.work_type)} → ${item.level}`))
    card.append(el('p', 'gw-quiet', item.rationale), el('span', 'gw-id', `${item.id} · ${timeLabel(item.created_at)}`))
    const actions = el('div', 'gw-suggestion-actions')
    const decide = (operation: 'promote' | 'dismiss') => async () => {
      await requestJSON(`/api/v1/policies/suggestions/${encodeURIComponent(item.id)}/${operation}`, 'POST')
      await renderPolicies()
    }
    actions.append(
      Button('Promote', { variant: 'primary', size: 'small', onClick: () => void decide('promote')().catch((error) => window.alert(error instanceof Error ? error.message : 'Promotion failed')) }),
      Button('Dismiss', { size: 'small', onClick: () => void decide('dismiss')().catch((error) => window.alert(error instanceof Error ? error.message : 'Dismissal failed')) }),
    )
    card.append(actions)
    list.append(card)
  }
  if (!suggestions.length) list.append(EmptyState('Queue clear', 'Groundwork has no pending policy-learning suggestions.'))
  return Panel('Suggestion queue', [list], [Badge(String(suggestions.length), suggestions.length ? 'warn' : 'idle')])
}

function settingsFacts(items: Array<[string, string]>) {
  const list = el('dl', 'gw-settings-facts')
  for (const [label, value] of items) {
    const row = el('div', 'gw-settings-fact')
    row.append(el('dt', undefined, label), el('dd', 'gw-mono', value || 'not configured'))
    list.append(row)
  }
  return list
}

function renderDoctorChecks(report: DoctorReport) {
  const list = el('div', 'gw-doctor-list')
  for (const check of report.checks) {
    const row = el('div', 'gw-doctor-check')
    const copy = el('div')
    copy.append(el('strong', undefined, pretty(check.name)), el('span', 'gw-row-detail', check.detail))
    row.append(copy, Badge(check.status, check.status === 'error' ? 'bad' : check.status === 'warn' ? 'warn' : 'ok'))
    list.append(row)
  }
  return list
}

async function renderSettings() {
  const app = el('main', 'gw gw-app gw-settings-page')
  app.append(appHeader('Settings', 'Resolved coordinator configuration, agent guidance, and project health.', true))
  const loading = el('div', 'gw-loading', 'Reading local configuration and running doctor…')
  app.append(loading)
  mount.replaceChildren(app)

  try {
    let [settings, report] = await Promise.all([
      getJSON<Settings>('/api/v1/settings'),
      requestJSON<DoctorReport>('/api/v1/doctor', 'POST'),
    ])
    loading.remove()

    const grid = el('div', 'gw-settings-grid')
    grid.append(
      Panel('Project paths', [settingsFacts([
        ['Repository', settings.repository_path],
        ['SQLite', settings.sqlite_path],
        ['Config', settings.config_path],
      ])], [Icon(FolderGit2, 'Project paths')]),
      Panel('Coordinator server', [settingsFacts([
        ['Address', settings.server.address],
        ['Bind', settings.server.bind],
        ['Port', settings.server.port],
      ])]),
      Panel('Agent runtime', [settingsFacts([
        ['Engine', settings.agent.engine],
        ['Model', settings.agent.model || 'runtime default'],
        ['Sandbox', settings.agent.sandbox],
      ])]),
      Panel('Scheduling', [settingsFacts([
        ['Concurrency', String(settings.concurrency.max)],
        ['Lease TTL', settings.concurrency.lease_ttl],
        ['Heartbeat', settings.concurrency.lease_heartbeat],
      ])]),
    )

    const agentsBody = el('div', 'gw-settings-action')
    const agentsCopy = el('div')
    agentsCopy.append(
      el('p', 'gw-settings-path gw-mono', settings.agents_md.path),
      el('p', 'gw-quiet', settings.agents_md.detail),
    )
    const syncNotice = el('p', 'gw-form-notice')
    syncNotice.setAttribute('role', 'status')
    const syncButton = Button(settings.agents_md.state === 'synced' ? 'Sync again' : 'Sync AGENTS.md', {
      variant: settings.agents_md.state === 'synced' ? 'default' : 'primary',
    })
    syncButton.addEventListener('click', () => {
      void (async () => {
        syncButton.disabled = true
        syncNotice.textContent = ''
        try {
          const status = await requestJSON<AgentsMDStatus>('/api/v1/settings/agents-md/sync', 'POST')
          settings = { ...settings, agents_md: status }
          syncNotice.className = 'gw-form-notice ok'
          syncNotice.textContent = status.detail
          syncButton.textContent = 'Sync again'
          agentsCopy.replaceChildren(
            el('p', 'gw-settings-path gw-mono', status.path),
            el('p', 'gw-quiet', status.detail),
          )
          statusBadge.replaceWith(statusBadge = Badge('synced', 'ok'))
        } catch (error) {
          syncNotice.className = 'gw-form-notice bad'
          syncNotice.textContent = error instanceof Error ? error.message : 'AGENTS.md sync failed.'
        } finally {
          syncButton.disabled = false
        }
      })()
    })
    agentsBody.append(agentsCopy, syncButton, syncNotice)
    let statusBadge = Badge(pretty(settings.agents_md.state), settings.agents_md.state === 'synced' ? 'ok' : 'warn')

    const doctorBody = el('div', 'gw-doctor-body')
    const doctorSummary = el('p', 'gw-quiet', report.healthy ? 'All required health checks passed.' : 'One or more health checks require attention.')
    let doctorList = renderDoctorChecks(report)
    let doctorBadge = Badge(report.healthy ? 'healthy' : 'attention', report.healthy ? 'ok' : 'bad')
    const runDoctor = Button('Run doctor', { icon: Activity })
    runDoctor.addEventListener('click', () => {
      void (async () => {
        runDoctor.disabled = true
        doctorSummary.textContent = 'Running gw doctor checks…'
        try {
          report = await requestJSON<DoctorReport>('/api/v1/doctor', 'POST')
          doctorSummary.textContent = report.healthy ? 'All required health checks passed.' : 'One or more health checks require attention.'
          const nextList = renderDoctorChecks(report)
          doctorList.replaceWith(nextList)
          doctorList = nextList
          doctorBadge.replaceWith(doctorBadge = Badge(report.healthy ? 'healthy' : 'attention', report.healthy ? 'ok' : 'bad'))
        } catch (error) {
          doctorSummary.textContent = error instanceof Error ? error.message : 'Doctor failed to run.'
        } finally {
          runDoctor.disabled = false
        }
      })()
    })
    doctorBody.append(doctorSummary, doctorList)

    app.append(
      grid,
      Panel('AGENTS.md sync', [agentsBody], [statusBadge]),
      Panel('gw doctor', [doctorBody], [doctorBadge, runDoctor]),
    )
  } catch (error) {
    loading.replaceWith(ErrorState('Unable to load settings', error instanceof Error ? error.message : 'Unknown coordinator error'))
  }
}

function approvalInboxRow(approval: Approval, appData: AppData) {
  const ticket = appData.tickets.find((candidate) => candidate.id === approval.ticket_id)
  const row = el('button', `gw-approval-row${approval.type === 'exception' ? ' exception' : ''}`)
  row.type = 'button'
  row.addEventListener('click', () => { window.location.hash = `#/approval/${encodeURIComponent(approval.id)}` })
  const copy = el('span', 'gw-approval-copy')
  copy.append(
    el('strong', undefined, approval.summary),
    el('span', 'gw-row-detail', `${approval.id} · ${approval.ticket_id}${ticket ? ` · ${ticket.title}` : ''}`),
    el('span', 'gw-row-detail', `Requested by ${approval.requested_by_actor || 'unknown actor'} · ${timeLabel(approval.created_at)}`),
  )
  const badges = el('span', 'gw-approval-badges')
  badges.append(Badge(pretty(approval.type), approval.type === 'exception' ? 'bad' : 'idle'), Badge(`${pretty(approval.risk_class)} risk`, riskTone(approval.risk_class)))
  row.append(copy, badges, Icon(ChevronRight))
  return row
}

function riskTone(risk: string): SemanticTone {
  if (risk === 'critical' || risk === 'high') return 'bad'
  if (risk === 'medium') return 'warn'
  return 'idle'
}

async function renderApproval(id: string) {
  const app = el('main', 'gw gw-app gw-approval-page')
  app.append(appHeader(id, 'Inspect the gate evidence before recording a decision.', true))
  const loading = el('div', 'gw-loading', 'Loading approval…')
  app.append(loading)
  mount.replaceChildren(app)
  try {
    const approval = await getJSON<Approval>(`/api/v1/approvals/${encodeURIComponent(id)}`)
    const preview = approval.type === 'land_to_main'
      ? await getJSON<LandPreview>(`/api/v1/tickets/${encodeURIComponent(approval.ticket_id)}/land/preview`).then((value) => ({ value })).catch((error: unknown) => ({ error }))
      : undefined
    loading.remove()
    app.append(approvalDecisionRail(approval), approvalDetail(approval, preview))
  } catch (error) {
    loading.replaceWith(ErrorState('Unable to load approval', error instanceof Error ? error.message : 'Unknown coordinator error'))
  }
}

function approvalDecisionRail(approval: Approval) {
  const rail = el('section', 'gw-action-rail gw-approval-actions')
  const copy = el('div', 'gw-action-copy')
  copy.append(el('p', 'gw-eyebrow', 'Human gate'), el('h2', undefined, `${approval.id} · ${pretty(approval.status)}`))
  const controls = el('div', 'gw-approval-controls')
  const reason = TextInput({ label: 'Decision reason', multiline: true, placeholder: 'Optional for approve; explain rejection or what needs clarification' })
  reason.setAttribute('aria-label', 'Decision reason')
  const buttons = el('div', 'gw-action-buttons')
  const error = el('p', 'gw-action-error')
  error.setAttribute('role', 'alert')
  const decide = (operation: 'approve' | 'reject' | 'clarify') => async () => {
    const choices = [...buttons.querySelectorAll('button')]
    for (const choice of choices) choice.disabled = true
    error.textContent = ''
    try {
      await requestJSON<Approval>(`/api/v1/approvals/${encodeURIComponent(approval.id)}/${operation}`, 'POST', { reason: reason.value.trim() })
      await refresh()
      window.location.hash = '#/approvals'
    } catch (cause) {
      error.textContent = cause instanceof Error ? cause.message : 'The coordinator rejected the decision.'
      for (const choice of choices) choice.disabled = false
    }
  }
  buttons.append(
    Button('Approve', { variant: 'primary', disabled: approval.status !== 'pending', onClick: () => { void decide('approve')() } }),
    Button('Reject', { variant: 'danger', disabled: approval.status !== 'pending', onClick: () => { void decide('reject')() } }),
    Button('Request clarification', { disabled: approval.status !== 'pending', onClick: () => { void decide('clarify')() } }),
  )
  controls.append(reason, buttons)
  rail.append(copy, controls, el('p', 'gw-action-hint', 'Decisions use the same coordinator approval service as gw approval approve, reject, and clarify.'), error)
  return rail
}

function approvalDetail(approval: Approval, preview?: { value: LandPreview } | { error: unknown }) {
  const detail = el('section', 'gw-detail')
  const heading = el('div', 'gw-detail-head')
  const title = el('div')
  title.append(el('p', 'gw-eyebrow', pretty(approval.type)), el('h2', undefined, approval.summary))
  heading.append(title, Badge(pretty(approval.risk_class), riskTone(approval.risk_class)))
  detail.append(heading)

  const facts = el('div', 'gw-run-facts')
  const factValues = [
    ['Ticket', approval.ticket_id],
    ['Requested by', approval.requested_by_actor || 'Not recorded'],
    ['Requested', timeLabel(approval.created_at)],
    ['Required actors', approval.required_actors?.join(', ') || 'Any authorized owner'],
    ['Required roles', approval.required_roles?.join(', ') || 'No additional role'],
    ['Reversible', approval.reversible === undefined ? 'Not assessed' : approval.reversible ? 'Yes' : 'No'],
  ]
  for (const [label, value] of factValues) {
    const row = el('div', 'gw-scope-row')
    row.append(el('span', 'sk', label), el('span', 'sv', value))
    facts.append(row)
  }
  const panels = [Panel('Approval details', [facts])]
  if (approval.action_json && approval.action_json !== '{}') panels.push(Panel('Requested action', [actionJSONView(approval.action_json)]))
  if (approval.type === 'land_to_main') panels.push(Panel('Landing diff preview', [landingPreviewView(preview)]))
  const grid = el('div', 'gw-detail-grid gw-approval-grid')
  grid.append(...panels)
  detail.append(grid)
  return detail
}

function actionJSONView(value: string) {
  let shown = value
  try { shown = JSON.stringify(JSON.parse(value), null, 2) } catch { /* Show the coordinator payload as recorded. */ }
  return el('pre', 'gw-diff-code', shown)
}

function landingPreviewView(preview?: { value: LandPreview } | { error: unknown }) {
  if (!preview) return el('p', 'gw-quiet', 'Landing preview was not requested.')
  if ('error' in preview) return ErrorState('Landing preview unavailable', preview.error instanceof Error ? preview.error.message : 'Unknown coordinator error')
  return diffView(preview.value, [])
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
  const [state, tickets, runs, readiness, approvals] = await Promise.all([
    getJSON<State>('/api/v1/state'),
    getJSON<Ticket[]>('/api/v1/tickets'),
    getJSON<Run[]>('/api/v1/runs'),
    getJSON<Readiness>('/api/v1/readiness'),
    getJSON<Approval[]>('/api/v1/approvals?status=pending'),
  ])
  return { state, tickets, runs, readiness, approvals }
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
  if (route.view === 'run') await renderRun(route.id)
  else if (route.view === 'approval') await renderApproval(route.id)
  else if (route.view === 'node') await renderNode(data, route.id)
  else if (route.view === 'readiness') renderReadiness(data)
  else if (route.view === 'approvals') renderApprovalsInbox(data)
  else if (route.view === 'policies') await renderPolicies()
  else if (route.view === 'settings') await renderSettings()
  else renderRootsBoard(data)
}

window.addEventListener('hashchange', () => { void render() })
void render()
