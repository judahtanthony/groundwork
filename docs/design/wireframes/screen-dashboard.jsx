// Groundwork — Dashboard screen (desktop + narrow). Live operations overview.

function ScreenDashboard() {
  const D = window.GW_DATA;
  return (
    <Shell active="dashboard"
      crumbs={['orchard-platform', 'Dashboard']}
      topActions={<><IconBtn icon="bell" title="Activity" /><Btn icon="plus" variant="primary">New ticket</Btn></>}>

      {/* KPI row */}
      <div className="gw-kpis" style={{ marginBottom: 16 }}>
        <Kpi label="Active runs" value="4" foot="2 awaiting input" tone="run" />
        <Kpi label="Blocked" value="2" foot="needs intervention" tone="bad" />
        <Kpi label="Pending approvals" value="3" foot="1 high-risk" tone="warn" />
        <Kpi label="In review" value="2" foot="ready to land" tone="warn" />
        <Kpi label="Validations failing" value="2" foot="across 2 tickets" tone="bad" />
        <Kpi label="Landed today" value="7" foot="all checks green" tone="ok" />
      </div>

      {/* Active runs */}
      <div className="gw-panel" style={{ marginBottom: 16 }}>
        <div className="gw-panel-head">
          <Icon name="runs" size={15} style={{ color: 'var(--gw-fg-muted)' }} />
          <h3>Active runs</h3>
          <span className="gw-badge run" style={{ marginLeft: 2 }}><i className="bdot" />4 live</span>
          <div className="gw-spacer" />
          <span className="gw-id">auto-refresh · 2s</span>
        </div>
        <table className="gw-table">
          <thead>
            <tr>
              <th style={{ width: '30%' }}>Ticket</th><th>Agent</th><th>Status</th>
              <th>Current step</th><th>Elapsed</th><th>Last event</th><th>Risk</th><th>Validation</th>
            </tr>
          </thead>
          <tbody>
            {D.runs.slice(0, 4).map((r) => (
              <tr key={r.id}>
                <td>
                  <div className="gw-cell-title" style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', maxWidth: 320 }}>{r.title}</div>
                  <div className="gw-id">{r.ticket} · {r.id}</div>
                </td>
                <td><Agent id={r.agent} /></td>
                <td><StatusBadge status={r.status} /></td>
                <td className="gw-muted" style={{ fontSize: 12 }}>{r.step}</td>
                <td className="num">{r.elapsed}</td>
                <td className="gw-subtle" style={{ fontSize: 11.5 }}>{r.last}</td>
                <td><RiskBadge score={r.risk} /></td>
                <td><ValBadge state={r.val} compact /></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* three columns */}
      <div style={{ display: 'grid', gridTemplateColumns: '1.1fr 1.1fr 0.85fr', gap: 16, alignItems: 'start' }}>
        {/* attention queue */}
        <div className="gw-panel">
          <div className="gw-panel-head">
            <Icon name="alert" size={15} style={{ color: 'var(--gw-warn)' }} />
            <h3>Attention queue</h3>
            <div className="gw-spacer" />
            <span className="gw-id">7 items</span>
          </div>
          <div className="gw-attn">
            <AttnItem tone="bad" icon="shield" title="Approve billing write + migration 0042" meta={['GW-298', 'high risk · 72', 'codex-2']} />
            <AttnItem tone="warn" icon="approvals" title="Land rate-limit branch → main" meta={['GW-280', 'human-required', '3m']} />
            <AttnItem tone="bad" icon="x" title="Run blocked · 14 type errors" meta={['GW-305', 'run_7b04', '6m']} />
            <AttnItem tone="bad" icon="x" title="Validation failing · lease heartbeat" meta={['GW-294', 'rework', '22m']} />
            <AttnItem tone="warn" icon="clock" title="Stale lease renewed automatically" meta={['GW-294', 'run_79c2', '38s stale']} />
            <AttnItem tone="warn" icon="approvals" title="Approve dependency bump (axios)" meta={['GW-305', 'medium · 63', 'codex-3']} />
          </div>
        </div>

        {/* recent timeline */}
        <div className="gw-panel">
          <div className="gw-panel-head">
            <Icon name="history" size={15} style={{ color: 'var(--gw-fg-muted)' }} />
            <h3>Recent events</h3>
            <div className="gw-spacer" />
            <span className="gw-id">last 15m</span>
          </div>
          <div className="gw-panel-body">
            <div className="gw-tl">
              {D.events.map((e, i) => <TlEvent key={i} {...e} last={i === D.events.length - 1} />)}
            </div>
          </div>
        </div>

        {/* resource summary */}
        <div className="gw-stack">
          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="gauge" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Resources</h3></div>
            <div className="gw-metrics" style={{ borderRadius: 0, border: 'none' }}>
              <Metric label="Tokens today" value="3.1" unit="M" />
              <Metric label="Est. cost" value="$18.42" />
              <Metric label="Agent runtime" value="2h 41m" />
              <Metric label="Concurrency" value="3 / 4" />
            </div>
          </div>
          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="cpu" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Local runtime</h3></div>
            <div className="gw-panel-body" style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              <div className="gw-between"><span className="gw-muted">Worktrees</span><span className="gw-mono">3 active · 1 idle</span></div>
              <div className="gw-between"><span className="gw-muted">Sandbox</span><span className="gw-badge ok"><i className="bdot" />workspace-write</span></div>
              <div className="gw-between"><span className="gw-muted">DB</span><span className="gw-mono">state.sqlite · 2.4 MB</span></div>
              <div className="gw-between"><span className="gw-muted">Uptime</span><span className="gw-mono">4h 12m</span></div>
            </div>
          </div>
        </div>
      </div>
    </Shell>
  );
}

function ScreenDashboardM() {
  const D = window.GW_DATA;
  return (
    <MobileShell active="dashboard" title="Dashboard">
      <div className="gw-panel" style={{ padding: '11px 13px' }}>
        <div className="gw-between">
          <div>
            <div className="gw-id">orchard-platform</div>
            <div style={{ fontWeight: 600, fontSize: 13, marginTop: 2 }}><Icon name="branch" size={12} style={{ verticalAlign: -1 }} /> main</div>
          </div>
          <span className="gw-badge ok"><span className="gw-dot live" />running</span>
        </div>
      </div>

      <div className="gw-m-kpis">
        <Kpi label="Active runs" value="4" tone="run" />
        <Kpi label="Blocked" value="2" tone="bad" />
        <Kpi label="Approvals" value="3" tone="warn" />
        <Kpi label="Failing" value="2" tone="bad" />
      </div>

      <div className="gw-panel">
        <div className="gw-panel-head"><Icon name="alert" size={14} style={{ color: 'var(--gw-warn)' }} /><h3>Attention queue</h3><div className="gw-spacer" /><span className="gw-id">7</span></div>
        <div className="gw-attn">
          <AttnItem tone="bad" icon="shield" title="Approve billing write" meta={['GW-298', 'high · 72']} />
          <AttnItem tone="warn" icon="approvals" title="Land branch → main" meta={['GW-280', '3m']} />
          <AttnItem tone="bad" icon="x" title="Run blocked · type errors" meta={['GW-305', '6m']} />
        </div>
      </div>

      <div className="gw-panel">
        <div className="gw-panel-head"><Icon name="runs" size={14} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Active runs</h3></div>
        <div style={{ padding: '4px 0' }}>
          {D.runs.slice(0, 3).map((r) => (
            <div key={r.id} style={{ padding: '10px 14px', borderBottom: '1px solid var(--gw-border)' }}>
              <div className="gw-between" style={{ marginBottom: 5 }}>
                <span className="gw-id">{r.ticket}</span><StatusBadge status={r.status} />
              </div>
              <div style={{ fontSize: 12.5, fontWeight: 600, marginBottom: 7, lineHeight: 1.35 }}>{r.title}</div>
              <div className="gw-between"><Agent id={r.agent} /><div style={{ display: 'flex', gap: 6 }}><RiskBadge score={r.risk} showBar={false} /><span className="num gw-subtle" style={{ fontSize: 11 }}>{r.elapsed}</span></div></div>
            </div>
          ))}
        </div>
      </div>
    </MobileShell>
  );
}

Object.assign(window, { ScreenDashboard, ScreenDashboardM });
