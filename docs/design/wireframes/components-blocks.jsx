// Groundwork — app shell (sidebar + topbar), mobile shell, and composite blocks.
// Depends on components-core.jsx (Icon, badges, etc.). Exports to window.

const GW_NAV = [
  { id: 'dashboard', label: 'Dashboard', icon: 'dashboard' },
  { id: 'board', label: 'Board', icon: 'board' },
  { id: 'tickets', label: 'Tickets', icon: 'ticket' },
  { id: 'runs', label: 'Runs', icon: 'runs', count: 4 },
  { id: 'approvals', label: 'Approvals', icon: 'approvals', count: 3, alert: true },
  { id: 'policies', label: 'Policies', icon: 'policies' },
  { id: 'settings', label: 'Settings', icon: 'settings' },
];

function Sidebar({ active }) {
  return (
    <aside className="gw-side">
      <div className="gw-side-brand">
        <GwMark size={28} />
        <div style={{ lineHeight: 1.2 }}>
          <div className="gw-brand-name">Groundwork</div>
          <div className="gw-brand-sub">agent coordinator</div>
        </div>
      </div>

      <div className="gw-repo">
        <div className="gw-repo-row">
          <Icon name="folder" size={14} style={{ color: '#9a9ca1', flex: '0 0 14px' }} />
          <span className="gw-repo-name">orchard-platform</span>
        </div>
        <div className="gw-repo-branch">
          <Icon name="branch" size={12} /> main · 3 worktrees
        </div>
      </div>

      <nav className="gw-nav">
        <div className="gw-nav-label">Operate</div>
        {GW_NAV.slice(0, 5).map((n) => (
          <div key={n.id} className={'gw-nav-item' + (active === n.id ? ' active' : '')}>
            <Icon name={n.icon} size={16} />{n.label}
            {n.count != null && <span className={'gw-nav-count' + (n.alert ? ' alert' : '')}>{n.count}</span>}
          </div>
        ))}
        <div className="gw-nav-label">Configure</div>
        {GW_NAV.slice(5).map((n) => (
          <div key={n.id} className={'gw-nav-item' + (active === n.id ? ' active' : '')}>
            <Icon name={n.icon} size={16} />{n.label}
          </div>
        ))}
      </nav>

      <div className="gw-side-foot">
        <div className="gw-server"><span className="gw-dot live" /> server · 127.0.0.1:4500</div>
        <div className="gw-server" style={{ marginTop: 6 }}><Icon name="layers" size={12} /> state.sqlite · 2.4 MB</div>
      </div>
    </aside>
  );
}

function Topbar({ crumbs, title, actions, refreshed = '2s ago' }) {
  return (
    <header className="gw-topbar">
      {crumbs ? (
        <div className="gw-crumbs">
          {crumbs.map((c, i) => (
            <React.Fragment key={i}>
              {i > 0 && <span className="sep"><Icon name="chevronR" size={12} /></span>}
              <span className={i === crumbs.length - 1 ? 'cur' : ''}>{c}</span>
            </React.Fragment>
          ))}
        </div>
      ) : <h1>{title}</h1>}
      <div className="gw-spacer" />
      <div className="gw-search"><Icon name="search" size={14} />Search tickets, runs…<span className="gw-kbd">⌘K</span></div>
      <div className="gw-refresh"><span className="gw-dot live" /> live · {refreshed}</div>
      {actions}
    </header>
  );
}

// Desktop shell wrapper
function Shell({ active, children, crumbs, title, topActions }) {
  return (
    <div className="gw gw-app">
      <Sidebar active={active} />
      <div className="gw-main">
        <Topbar crumbs={crumbs} title={title} actions={topActions} />
        <div className="gw-content">{children}</div>
      </div>
    </div>
  );
}

// Mobile shell
function MobileShell({ active, title, children }) {
  const items = [
    { id: 'dashboard', label: 'Home', icon: 'dashboard' },
    { id: 'board', label: 'Board', icon: 'board' },
    { id: 'runs', label: 'Runs', icon: 'runs' },
    { id: 'approvals', label: 'Approve', icon: 'approvals' },
  ];
  return (
    <div className="gw gw-m">
      <div className="gw-m-top">
        <GwMark size={22} />
        <div className="gw-m-title">{title}</div>
        <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: 6 }}>
          <span className="gw-dot live" /><span style={{ fontFamily: 'var(--gw-mono)', fontSize: 10, color: '#9a9ca1' }}>LIVE</span>
        </div>
      </div>
      <div className="gw-m-content">{children}</div>
      <div className="gw-m-tabs">
        {items.map((it) => (
          <div key={it.id} className={'gw-m-tab' + (active === it.id ? ' active' : '')}>
            <Icon name={it.icon} size={19} />{it.label}
          </div>
        ))}
      </div>
    </div>
  );
}

// ---- KPI tile ----
function Kpi({ label, value, foot, tone }) {
  return (
    <div className={'gw-kpi' + (tone ? ' ' + tone : '')}>
      {tone && <span className="k-accent" />}
      <div className="k-label">{label}</div>
      <div className="k-val">{value}</div>
      {foot && <div className="k-foot">{foot}</div>}
    </div>
  );
}

// ---- attention queue item ----
function AttnItem({ tone, icon, title, meta }) {
  return (
    <div className="gw-attn-item">
      <div className={'gw-attn-ic ' + tone}><Icon name={icon} size={15} /></div>
      <div className="gw-attn-body">
        <div className="gw-attn-title">{title}</div>
        <div className="gw-attn-meta">{meta.map((m, i) => <span key={i}>{i > 0 && '· '}{m}</span>)}</div>
      </div>
      <div className="gw-attn-go"><Icon name="chevronR" size={15} /></div>
    </div>
  );
}

// ---- timeline event ----
function TlEvent({ tone, text, time, last }) {
  return (
    <div className="gw-tl-item">
      <div className="gw-tl-rail">
        <div className={'gw-tl-node' + (tone ? ' ' + tone : '')} />
        {!last && <div className="gw-tl-line" />}
      </div>
      <div className="gw-tl-body">
        <div className="gw-tl-text">{text}</div>
        <div className="gw-tl-time">{time}</div>
      </div>
    </div>
  );
}

// ---- diff summary ----
function DiffSummary({ files, add, del, list }) {
  const total = add + del;
  const addBars = Math.max(0, Math.round((add / total) * 5));
  return (
    <div>
      <div className="gw-diff-head">
        <span><b style={{ fontWeight: 600 }}>{files}</b> files</span>
        <span className="gw-diff-stat gw-add">+{add}</span>
        <span className="gw-diff-stat gw-del">−{del}</span>
        <span className="gw-diff-bar">
          {[0, 1, 2, 3, 4].map((i) => <i key={i} style={{ background: i < addBars ? 'var(--gw-ok)' : 'var(--gw-bad)' }} />)}
        </span>
      </div>
      {list && (
        <div style={{ marginTop: 10 }}>
          {list.map((f, i) => (
            <div className="gw-diff-file" key={i}>
              <Icon name="file" size={13} style={{ color: 'var(--gw-fg-subtle)', flex: '0 0 13px' }} />
              <span className="fpath">{f.path}</span>
              <span className="fnums"><span className="gw-add">+{f.add}</span><span className="gw-del">−{f.del}</span></span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ---- command block ----
function CommandBlock({ label = 'Requested command', children }) {
  return (
    <div className="gw-cmd">
      <div className="gw-cmd-head">
        <Icon name="terminal" size={13} style={{ color: '#8d9096' }} />
        <span className="lbl">{label}</span>
        <span className="copy"><Icon name="copy" size={12} /> Copy</span>
      </div>
      <div className="gw-cmd-body">{children}</div>
    </div>
  );
}

// ---- scope rows ----
function ScopeRow({ k, children }) {
  return <div className="gw-scope-row"><span className="sk">{k}</span><span className="sv">{children}</span></div>;
}

// ---- policy row ----
function PolicyRow({ cond, action, scope, on, mode }) {
  const modeBadge = mode === 'auto' ? <span className="gw-badge ok"><Icon name="zap" size={11} sw={2} />Auto-approve</span>
    : mode === 'human' ? <span className="gw-badge warn"><Icon name="user" size={11} sw={2} />Require human</span>
    : <span className="gw-badge run"><Icon name="eye" size={11} sw={2} />Reviewer agent</span>;
  return (
    <div className="gw-policy">
      <div style={{ minWidth: 0 }}>
        <div className="gw-policy-cond">{cond}</div>
        {scope && <div className="gw-mono gw-subtle" style={{ fontSize: 11, marginTop: 3 }}>{scope}</div>}
      </div>
      <div className="gw-policy-act">
        {modeBadge}
        <Toggle on={on} />
      </div>
    </div>
  );
}

// ---- checklist item ----
function Check({ done, pending, children }) {
  return (
    <div className={'gw-check' + (done ? ' done' : '')}>
      <div className={'box' + (pending ? ' pend' : '')}>{done && <Icon name="check" size={11} sw={3} />}</div>
      <span className="ctext">{children}</span>
    </div>
  );
}

// ---- metric tile ----
function Metric({ label, value, unit }) {
  return <div className="gw-metric"><div className="m-l">{label}</div><div className="m-v">{value}{unit && <small> {unit}</small>}</div></div>;
}

Object.assign(window, { GW_NAV, Sidebar, Topbar, Shell, MobileShell, Kpi, AttnItem, TlEvent, DiffSummary, CommandBlock, ScopeRow, PolicyRow, Check, Metric });
