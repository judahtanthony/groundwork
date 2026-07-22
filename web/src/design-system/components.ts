import {
  Bot,
  Boxes,
  ArrowRight,
  CircleAlert,
  CircleCheck,
  CircleDashed,
  CircleUser,
  GitBranch,
  GripVertical,
  LockKeyhole,
  Scale,
  type IconNode,
  createElement as createLucideElement,
} from 'lucide'

export type ComponentChild = Node | string | number | null | undefined | false
export type SemanticTone = 'ok' | 'warn' | 'bad' | 'run' | 'idle'

function append(parent: HTMLElement, children: ComponentChild[]) {
  for (const child of children.flat()) {
    if (child === null || child === undefined || child === false) continue
    parent.append(child instanceof Node ? child : document.createTextNode(String(child)))
  }
}

function classes(...values: Array<string | false | null | undefined>) {
  return values.filter(Boolean).join(' ')
}

export function Icon(icon: IconNode, label?: string) {
  const element = createLucideElement(icon, {
    'aria-hidden': label ? undefined : 'true',
    'aria-label': label,
    focusable: 'false',
  })
  if (label) element.setAttribute('role', 'img')
  return element
}

export type ButtonOptions = {
  variant?: 'default' | 'primary' | 'danger' | 'ghost'
  size?: 'default' | 'small'
  icon?: IconNode
  disabled?: boolean
  type?: 'button' | 'submit' | 'reset'
  onClick?: (event: MouseEvent) => void
}

export function Button(label: string, options: ButtonOptions = {}) {
  const button = document.createElement('button')
  button.className = classes(
    'gw-btn',
    options.variant && options.variant !== 'default' && options.variant,
    options.size === 'small' && 'sm',
  )
  button.type = options.type ?? 'button'
  button.disabled = options.disabled ?? false
  if (options.icon) button.append(Icon(options.icon))
  button.append(label)
  if (options.onClick) button.addEventListener('click', options.onClick)
  return button
}

export function IconButton(label: string, icon: IconNode, onClick?: (event: MouseEvent) => void) {
  const button = document.createElement('button')
  button.className = 'gw-icon-btn'
  button.type = 'button'
  button.title = label
  button.setAttribute('aria-label', label)
  button.append(Icon(icon))
  if (onClick) button.addEventListener('click', onClick)
  return button
}

export function Panel(title: string, children: ComponentChild[], actions: ComponentChild[] = []) {
  const panel = document.createElement('section')
  panel.className = 'gw-panel'

  const head = document.createElement('header')
  head.className = 'gw-panel-head'
  const heading = document.createElement('h3')
  heading.textContent = title
  const spacer = document.createElement('span')
  spacer.className = 'gw-spacer'
  head.append(heading, spacer)
  append(head, actions)

  const body = document.createElement('div')
  body.className = 'gw-panel-body'
  append(body, children)
  panel.append(head, body)
  return panel
}

export function Badge(label: string, tone: SemanticTone = 'idle', withDot = true) {
  const badge = document.createElement('span')
  badge.className = `gw-badge ${tone}`
  if (withDot) {
    const dot = document.createElement('span')
    dot.className = 'bdot'
    dot.setAttribute('aria-hidden', 'true')
    badge.append(dot)
  }
  badge.append(label)
  return badge
}

export type RiskLevel = 'low' | 'medium' | 'high' | 'critical'

export function RiskBadge(score: number, level: RiskLevel) {
  const normalizedScore = Math.max(0, Math.min(100, Math.round(score)))
  const risk = document.createElement('span')
  risk.className = `gw-risk ${level === 'medium' ? 'med' : level === 'critical' ? 'high' : level}`
  risk.setAttribute('aria-label', `${level} risk, score ${normalizedScore} of 100`)

  const bar = document.createElement('span')
  bar.className = 'rbar'
  bar.setAttribute('aria-hidden', 'true')
  const fill = document.createElement('i')
  fill.style.width = `${normalizedScore}%`
  bar.append(fill)
  risk.append(bar, `${level} · ${normalizedScore}`)
  return risk
}

export function Chip(label: string, color?: string) {
  const chip = document.createElement('span')
  chip.className = classes('gw-chip', Boolean(color) && 'dotted')
  chip.textContent = label
  if (color) chip.style.setProperty('--c', color)
  return chip
}

export type ActorKind = 'human' | 'ai' | 'judge'

const actorIcons: Record<ActorKind, IconNode> = {
  human: CircleUser,
  ai: Bot,
  judge: Scale,
}

export function ActorBadge(label: string, kind: ActorKind) {
  const badge = document.createElement('span')
  badge.className = `gw-actor ${kind}`
  const dot = document.createElement('span')
  dot.className = 'adot'
  dot.setAttribute('aria-hidden', 'true')
  dot.append(Icon(actorIcons[kind]))
  badge.append(dot, label)
  return badge
}

export type NodeType = 'leaf' | 'composite'

export function NodeTypeBadge(type: NodeType) {
  const badge = document.createElement('span')
  badge.className = classes('gw-ntype', type === 'composite' && 'composite')
  badge.append(Icon(type === 'composite' ? Boxes : CircleDashed), type)
  return badge
}

export function KindLabel(kind: string) {
  const label = document.createElement('span')
  label.className = 'gw-kind'
  label.textContent = kind
  return label
}

export function DependencyBadge(ticketID: string, direction: 'blocks' | 'blocked-by') {
  const badge = document.createElement('span')
  badge.className = `gw-dep ${direction}`
  badge.append(Icon(direction === 'blocked-by' ? LockKeyhole : GitBranch), `${direction} ${ticketID}`)
  return badge
}

export function WaitBadge(label = 'Waiting on dependencies') {
  const badge = document.createElement('span')
  badge.className = 'gw-wait'
  badge.append(Icon(LockKeyhole), label)
  return badge
}

export type AutonomyLevel = 'human' | 'reviewer' | 'auto'

export function AutonomyBadge(level: AutonomyLevel) {
  const badge = document.createElement('span')
  badge.className = `gw-auto ${level}`
  badge.textContent = level
  return badge
}

export function Toggle(label: string, checked = false, onChange?: (checked: boolean) => void) {
  const toggle = document.createElement('button')
  toggle.className = classes('gw-toggle', checked && 'on')
  toggle.type = 'button'
  toggle.setAttribute('role', 'switch')
  toggle.setAttribute('aria-label', label)
  toggle.setAttribute('aria-checked', String(checked))
  toggle.addEventListener('click', () => {
    checked = !checked
    toggle.classList.toggle('on', checked)
    toggle.setAttribute('aria-checked', String(checked))
    onChange?.(checked)
  })
  return toggle
}

export type TextInputOptions = {
  label: string
  value?: string
  placeholder?: string
  readOnly?: boolean
  multiline?: boolean
  onInput?: (value: string) => void
}

export function TextInput(options: TextInputOptions) {
  const input = options.multiline ? document.createElement('textarea') : document.createElement('input')
  input.className = classes('gw-input', options.readOnly && 'ro')
  input.setAttribute('aria-label', options.label)
  input.value = options.value ?? ''
  input.placeholder = options.placeholder ?? ''
  input.readOnly = options.readOnly ?? false
  if (options.onInput) input.addEventListener('input', () => options.onInput?.(input.value))
  return input
}

export function SegmentedControl<T extends string>(
  label: string,
  options: readonly T[],
  selected: T,
  onChange?: (value: T) => void,
) {
  const control = document.createElement('div')
  control.className = 'gw-seg'
  control.setAttribute('role', 'group')
  control.setAttribute('aria-label', label)

  for (const option of options) {
    const button = document.createElement('button')
    button.type = 'button'
    button.className = option === selected ? 'on' : ''
    button.setAttribute('aria-pressed', String(option === selected))
    button.textContent = option
    button.addEventListener('click', () => {
      selected = option
      for (const item of control.querySelectorAll('button')) {
        const active = item === button
        item.classList.toggle('on', active)
        item.setAttribute('aria-pressed', String(active))
      }
      onChange?.(option)
    })
    control.append(button)
  }
  return control
}

export function DragHandle(label = 'Drag to reorder') {
  const handle = document.createElement('span')
  handle.className = 'gw-drag'
  handle.title = label
  handle.setAttribute('aria-label', label)
  handle.append(Icon(GripVertical))
  return handle
}

export type DependencyPeekItem = {
  ticketID: string
  detail: string
  crossTree?: boolean
  onSelect?: () => void
}

export function DependencyPeek(title: string, items: readonly DependencyPeekItem[]) {
  const peek = document.createElement('div')
  peek.className = 'gw-peek'
  const heading = document.createElement('div')
  heading.className = 'gw-peek-head'
  heading.textContent = title
  peek.append(heading)

  for (const item of items) {
    const row = document.createElement('button')
    row.className = classes('gw-peek-item', item.crossTree && 'xtree')
    row.type = 'button'
    const id = document.createElement('span')
    id.className = 'pi-id'
    id.textContent = item.ticketID
    const detail = document.createElement('span')
    detail.className = 'pi-title'
    detail.textContent = item.detail
    const jump = document.createElement('span')
    jump.className = 'pi-jump'
    jump.setAttribute('aria-hidden', 'true')
    jump.append(Icon(ArrowRight))
    row.append(id, detail, jump)
    if (item.onSelect) row.addEventListener('click', item.onSelect)
    peek.append(row)
  }
  return peek
}

export function EmptyState(title: string, detail: string, icon: IconNode = CircleDashed) {
  const state = document.createElement('div')
  state.className = 'gw-empty'
  const symbol = document.createElement('span')
  symbol.className = 'ec-ic'
  symbol.append(Icon(icon))
  const heading = document.createElement('p')
  heading.className = 'ec-t'
  heading.textContent = title
  const description = document.createElement('p')
  description.className = 'ec-s'
  description.textContent = detail
  state.append(symbol, heading, description)
  return state
}

export function ErrorState(code: string, message: string) {
  const state = document.createElement('div')
  state.className = 'gw-error'
  state.setAttribute('role', 'alert')
  state.append(Icon(CircleAlert))
  const copy = document.createElement('div')
  const label = document.createElement('p')
  label.className = 'er-code'
  label.textContent = code
  const detail = document.createElement('p')
  detail.className = 'er-msg'
  detail.textContent = message
  copy.append(label, detail)
  state.append(copy)
  return state
}

export function ChecklistItem(label: string, done = false) {
  const item = document.createElement('div')
  item.className = classes('gw-check', done && 'done')
  const box = document.createElement('span')
  box.className = 'box'
  box.setAttribute('aria-hidden', 'true')
  if (done) box.append(Icon(CircleCheck))
  const text = document.createElement('span')
  text.className = 'ctext'
  text.textContent = label
  item.append(box, text)
  return item
}

export function Skeleton(width = '100%', height = '1rem') {
  const skeleton = document.createElement('span')
  skeleton.className = 'gw-skel'
  skeleton.style.display = 'block'
  skeleton.style.width = width
  skeleton.style.height = height
  skeleton.setAttribute('aria-hidden', 'true')
  return skeleton
}
