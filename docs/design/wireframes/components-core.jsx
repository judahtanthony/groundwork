// Groundwork — core components: icon set, brand mark, badges, primitives, app shell.
// Exports to window for use by screen files. Plain inline-SVG icons (no CDN).

const GW_ICONS = {
  dashboard: 'M3 3h7v7H3zM14 3h7v5h-7zM14 12h7v9h-7zM3 14h7v7H3z',
  board: 'M4 4h4v16H4zM10 4h4v11h-4zM16 4h4v8h-4z',
  ticket: 'M4 7a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v2a2 2 0 0 0 0 4v2a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2v-2a2 2 0 0 0 0-4z M13 5v14',
  runs: 'M5 4l5 4-5 4M5 12l5 4-5 4M13 20h6',
  approvals: 'M12 3l7 3v5c0 4.5-3 7.5-7 9-4-1.5-7-4.5-7-9V6z M9 12l2 2 4-4',
  policies: 'M5 4v16M5 7h10l-2 3 2 3H5M16 4v7',
  settings: 'M12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6z M19.4 13a1.6 1.6 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.6 1.6 0 0 0-1.8-.3 1.6 1.6 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.6 1.6 0 0 0-1-1.5 1.6 1.6 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.6 1.6 0 0 0 .3-1.8 1.6 1.6 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1a1.6 1.6 0 0 0 1.5-1 1.6 1.6 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.6 1.6 0 0 0 1.8.3H9a1.6 1.6 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.6 1.6 0 0 0 1 1.5 1.6 1.6 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.6 1.6 0 0 0-.3 1.8V9a1.6 1.6 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.6 1.6 0 0 0-1.5 1z',
  search: 'M11 4a7 7 0 1 0 0 14 7 7 0 0 0 0-14zM20 20l-3.5-3.5',
  refresh: 'M21 12a9 9 0 1 1-3-6.7L21 7M21 3v4h-4',
  play: 'M6 4l14 8-14 8z',
  pause: 'M7 4h4v16H7zM15 4h4v16h-4z',
  x: 'M6 6l12 12M18 6L6 18',
  check: 'M5 12l5 5L20 6',
  chevronR: 'M9 5l7 7-7 7',
  chevronD: 'M5 9l7 7 7-7',
  arrowR: 'M4 12h15M13 5l7 7-7 7',
  more: 'M5 12h.01M12 12h.01M19 12h.01',
  copy: 'M9 9h11v11H9zM5 15H4V4h11v1',
  external: 'M14 4h6v6M20 4l-9 9M19 14v5a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1h5',
  folder: 'M3 6a1 1 0 0 1 1-1h5l2 2h8a1 1 0 0 1 1 1v10a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1z',
  branch: 'M6 4a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM6 8v8M6 20a2 2 0 1 0 0-4 2 2 0 0 0 0 4zM18 6a2 2 0 1 0 0-4 2 2 0 0 0 0 4zM18 6c0 5-6 4-6 9',
  clock: 'M12 4a8 8 0 1 0 0 16 8 8 0 0 0 0-16zM12 8v4l3 2',
  alert: 'M12 4l9 16H3zM12 10v4M12 17h.01',
  shield: 'M12 3l7 3v5c0 4.5-3 7.5-7 9-4-1.5-7-4.5-7-9V6z',
  file: 'M7 3h7l4 4v14H7zM14 3v4h4',
  plus: 'M12 5v14M5 12h14',
  terminal: 'M5 6l5 5-5 5M12 17h7',
  cpu: 'M7 7h10v10H7zM4 9h3M4 15h3M17 9h3M17 15h3M9 4v3M15 4v3M9 17v3M15 17v3',
  gauge: 'M12 14l4-4M5.5 18a8 8 0 1 1 13 0z',
  lock: 'M6 10h12v10H6zM8 10V7a4 4 0 0 1 8 0v3',
  zap: 'M13 3L5 13h6l-1 8 8-10h-6z',
  undo: 'M9 7L4 12l5 5M4 12h11a5 5 0 0 1 0 10h-1',
  download: 'M12 4v11M7 11l5 5 5-5M5 20h14',
  message: 'M4 5h16v11H9l-5 4z',
  edit: 'M4 20h4L19 9l-4-4L4 16zM14 6l4 4',
  bell: 'M6 9a6 6 0 0 1 12 0c0 5 2 6 2 6H4s2-1 2-6M10 20a2 2 0 0 0 4 0',
  pin: 'M12 3l2 6h6l-5 4 2 7-5-4-5 4 2-7-5-4h6z',
  flow: 'M5 12h14M12 5l7 7-7 7',
  user: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8zM4 21a8 8 0 0 1 16 0',
  pkg: 'M12 3l8 4v10l-8 4-8-4V7zM4 7l8 4 8-4M12 11v10',
  doc: 'M7 3h7l4 4v14H7zM14 3v4h4M9 13h6M9 16h6',
  globe: 'M12 4a8 8 0 1 0 0 16 8 8 0 0 0 0-16zM4 12h16M12 4c2.5 2 2.5 14 0 16M12 4c-2.5 2-2.5 14 0 16',
  trash: 'M5 7h14M9 7V5h6v2M7 7l1 13h8l1-13',
  eye: 'M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7zM12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6z',
  history: 'M4 12a8 8 0 1 1 3 6.2M4 12H1m3 0V9M12 8v4l3 2',
  stop: 'M7 7h10v10H7z',
  layers: 'M12 3l9 5-9 5-9-5zM3 13l9 5 9-5M3 17l9 5 9-5',
};

function Icon({ name, size = 16, sw = 1.75, fill = false, style }) {
  const d = GW_ICONS[name] || '';
  return (
    <svg width={size} height={size} viewBox="0 0 24 24"
      fill={fill ? 'currentColor' : 'none'} stroke={fill ? 'none' : 'currentColor'}
      strokeWidth={sw} strokeLinecap="round" strokeLinejoin="round" style={style}>
      {d.split('M').filter(Boolean).map((seg, i) => <path key={i} d={'M' + seg} />)}
    </svg>
  );
}

// Groundwork emblem — stratified ground layers (foundation / groundwork).
function GwMark({ size = 26, light = true }) {
  const stroke = light ? '#E9E5DC' : '#1E1F21';
  const accent = '#9c4b3b';
  return (
    <svg width={size} height={size} viewBox="0 0 28 28" fill="none">
      <rect x="1" y="1" width="26" height="26" rx="6" fill={light ? '#292A2C' : '#fff'} stroke={light ? '#3a3c40' : '#D6D2C8'} />
      <path d="M5 18.5h18" stroke={stroke} strokeWidth="1.6" strokeLinecap="round" opacity=".55" />
      <path d="M5 14.5h18" stroke={stroke} strokeWidth="1.6" strokeLinecap="round" opacity=".8" />
      <path d="M5 10.5h18" stroke={accent} strokeWidth="1.8" strokeLinecap="round" />
      <circle cx="14" cy="10.5" r="2.1" fill={accent} stroke={light ? '#292A2C' : '#fff'} strokeWidth="1.4" />
    </svg>
  );
}

// ---- status badge ----
const GW_STATUS = {
  running:   { cls: 'run',  label: 'Running' },
  blocked:   { cls: 'bad',  label: 'Blocked' },
  review:    { cls: 'warn', label: 'In review' },
  rework:    { cls: 'warn', label: 'Rework' },
  approved:  { cls: 'ok',   label: 'Approved' },
  landing:   { cls: 'run',  label: 'Landing' },
  done:      { cls: 'idle', label: 'Done' },
  queued:    { cls: 'idle', label: 'Queued' },
  paused:    { cls: 'idle', label: 'Paused' },
  failed:    { cls: 'bad',  label: 'Failed' },
  cancelled: { cls: 'idle', label: 'Cancelled' },
  todo:      { cls: 'idle', label: 'Todo' },
  in_progress:{cls: 'run',  label: 'In progress' },
  backlog:   { cls: 'idle', label: 'Backlog' },
};
function StatusBadge({ status, label }) {
  const s = GW_STATUS[status] || { cls: 'idle', label: label || status };
  return <span className={'gw-badge ' + s.cls}><i className="bdot" />{label || s.label}</span>;
}

// ---- validation badge ----
function ValBadge({ state, count, compact }) {
  // state: pass | fail | pending | none
  if (state === 'pass') return <span className="gw-badge ok"><Icon name="check" size={12} sw={2.4} />{compact ? (count || 'Pass') : (count ? count + ' passing' : 'Validations pass')}</span>;
  if (state === 'fail') return <span className="gw-badge bad"><Icon name="x" size={12} sw={2.4} />{compact ? 'Fail' : (count ? count + ' failing' : 'Validation failed')}</span>;
  if (state === 'pending') return <span className="gw-badge warn"><Icon name="clock" size={12} sw={2.2} />{compact ? 'Running' : 'Validating'}</span>;
  return <span className="gw-badge idle"><i className="bdot" />{compact ? '—' : 'No checks'}</span>;
}

// ---- risk badge ----
function RiskBadge({ score, showBar = true }) {
  const tier = score >= 67 ? 'high' : score >= 34 ? 'med' : 'low';
  return (
    <span className={'gw-risk ' + tier} title={'Risk score ' + score + ' / 100'}>
      {showBar && <span className="rbar"><i style={{ width: Math.max(8, score) + '%' }} /></span>}
      {score}
    </span>
  );
}

// ---- agent token ----
const GW_AGENT_COLORS = { codex: '#3A6EA5', 'codex-2': '#5F6B57', 'codex-3': '#9c4b3b', reviewer: '#946400', human: '#6B6F76' };
function Agent({ id, name }) {
  const key = id || 'codex';
  const color = GW_AGENT_COLORS[key] || '#3A6EA5';
  const initials = (name || key).replace(/[^a-z0-9]/gi, '').slice(0, 2).toUpperCase();
  return <span className="gw-agent"><span className="av" style={{ background: color }}>{initials}</span>{name || key}</span>;
}

function Chip({ children, color }) {
  return <span className={'gw-chip' + (color ? ' dotted' : '')} style={color ? { '--c': color } : null}>{children}</span>;
}

function Toggle({ on }) { return <span className={'gw-toggle' + (on ? ' on' : '')} />; }

function Btn({ children, icon, variant, sm, style }) {
  return (
    <button className={'gw-btn' + (variant ? ' ' + variant : '') + (sm ? ' sm' : '')} style={style}>
      {icon && <Icon name={icon} size={sm ? 13 : 14} />}{children}
    </button>
  );
}
function IconBtn({ icon, title }) { return <button className="gw-icon-btn" title={title}><Icon name={icon} /></button>; }

Object.assign(window, { GW_ICONS, Icon, GwMark, StatusBadge, ValBadge, RiskBadge, Agent, Chip, Toggle, Btn, IconBtn, GW_STATUS });
