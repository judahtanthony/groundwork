// Groundwork — Approvals Inbox (desktop + narrow). Decide tactical approvals.

function ApprovalListItem({ a, active }) {
  const tierBadge = a.tier === 'high' ? <span className="gw-badge bad">High</span>
    : a.tier === 'medium' ? <span className="gw-badge warn">Medium</span>
      : <span className="gw-badge ok">Low</span>;
  return (
    <div style={{
      padding: '11px 14px', borderBottom: '1px solid var(--gw-border)', cursor: 'pointer',
      background: active ? 'var(--gw-accent-tint)' : 'transparent',
      borderLeft: active ? '3px solid var(--gw-accent)' : '3px solid transparent',
    }}>
      <div className="gw-between" style={{ marginBottom: 5 }}>
        <span className="gw-id">{a.ticket} · {a.run}</span>
        <RiskBadge score={a.risk} showBar={false} />
      </div>
      <div style={{ fontSize: 12.5, fontWeight: 600, lineHeight: 1.35, marginBottom: 7 }}>{a.action}</div>
      <div className="gw-between"><Agent id={a.agent} />{tierBadge}</div>
    </div>
  );
}

function ScreenApprovals() {
  const D = window.GW_DATA;
  const a = D.approvals[0]; // high-risk billing
  return (
    <Shell active="approvals"
      crumbs={['orchard-platform', 'Approvals']}
      topActions={<><div className="gw-seg"><button className="on">Pending</button><button>Resolved</button></div></>}>

      <div style={{ display: 'grid', gridTemplateColumns: '340px 1fr', gap: 16, alignItems: 'start' }}>
        {/* inbox list */}
        <div className="gw-panel" style={{ overflow: 'hidden' }}>
          <div className="gw-panel-head"><Icon name="approvals" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Inbox</h3><div className="gw-spacer" /><span className="gw-badge bad"><i className="bdot" />3 pending</span></div>

          <div style={{ padding: '9px 14px 6px', background: 'var(--gw-surface-2)', borderBottom: '1px solid var(--gw-border)' }}>
            <span className="gw-mono" style={{ fontSize: 9.5, letterSpacing: '.12em', textTransform: 'uppercase', color: 'var(--gw-bad)' }}>High risk · 1</span>
          </div>
          <ApprovalListItem a={D.approvals[0]} active />

          <div style={{ padding: '9px 14px 6px', background: 'var(--gw-surface-2)', borderBottom: '1px solid var(--gw-border)' }}>
            <span className="gw-mono" style={{ fontSize: 9.5, letterSpacing: '.12em', textTransform: 'uppercase', color: 'var(--gw-warn)' }}>Medium risk · 2</span>
          </div>
          <ApprovalListItem a={D.approvals[1]} />
          <ApprovalListItem a={D.approvals[2]} />

          <div style={{ padding: '9px 14px 6px', background: 'var(--gw-surface-2)', borderBottom: '1px solid var(--gw-border)' }}>
            <span className="gw-mono" style={{ fontSize: 9.5, letterSpacing: '.12em', textTransform: 'uppercase', color: 'var(--gw-ok)' }}>Auto-approved · last hour</span>
          </div>
          <div style={{ padding: '10px 14px', display: 'flex', alignItems: 'center', gap: 9 }}>
            <Icon name="zap" size={14} style={{ color: 'var(--gw-ok)' }} />
            <div style={{ minWidth: 0, flex: 1 }}>
              <div style={{ fontSize: 12, fontWeight: 600 }}>Docs edit to AGENTS.md</div>
              <div className="gw-id">GW-322 · low · 8 · matched rule R-04</div>
            </div>
            <span className="gw-badge ok"><Icon name="check" size={11} sw={2.4} /></span>
          </div>
        </div>

        {/* detail */}
        <div className="gw-appr flag-bad">
          <div className="gw-appr-top">
            <div className="gw-attn-ic bad" style={{ flex: '0 0 28px' }}><Icon name="alert" size={16} /></div>
            <div style={{ flex: 1 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 9, marginBottom: 3 }}>
                <span className="gw-id">{a.ticket} · {a.run}</span><span className="gw-badge bad">High risk</span><RiskBadge score={a.risk} />
              </div>
              <div className="ttl">{a.action}</div>
              <div className="why">{a.why}</div>
            </div>
          </div>

          <div className="gw-appr-sect">
            <div className="gw-sect-label">Requested command</div>
            <CommandBlock>
              <div><span className="prompt">codex-2 $</span> psql -f <span className="str">billing/migrations/0042_ledger.sql</span></div>
              <div><span className="prompt">codex-2 $</span> gw apply <span className="flag">--write</span> billing/reconcile/worker.go</div>
            </CommandBlock>
          </div>

          <div className="gw-appr-sect">
            <div className="gw-sect-label">Scope</div>
            <div className="gw-scope">
              <ScopeRow k="files"><span style={{ color: 'var(--gw-bad)' }}>billing/reconcile/*</span> · billing/migrations/0042_ledger.sql</ScopeRow>
              <ScopeRow k="database">migration · adds table + unique index · <span style={{ color: 'var(--gw-bad)' }}>non-reversible</span></ScopeRow>
              <ScopeRow k="network">none</ScopeRow>
              <ScopeRow k="sandbox">workspace-write · escalation requested</ScopeRow>
            </div>
          </div>

          <div className="gw-appr-sect">
            <div className="gw-sect-label">Matched policy rules</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <div className="gw-between" style={{ padding: '8px 11px', background: 'var(--gw-bad-bg)', border: '1px solid var(--gw-bad-bd)', borderRadius: 6 }}>
                <span style={{ fontSize: 12.5 }}><span className="gw-mono" style={{ color: 'var(--gw-bad)', fontSize: 11 }}>R-01</span> · Write to <span className="gw-mono">billing/**</span> always requires human</span>
                <span className="gw-badge bad"><Icon name="user" size={11} sw={2} />Require human</span>
              </div>
              <div className="gw-between" style={{ padding: '8px 11px', background: 'var(--gw-warn-bg)', border: '1px solid var(--gw-warn-bd)', borderRadius: 6 }}>
                <span style={{ fontSize: 12.5 }}><span className="gw-mono" style={{ color: 'var(--gw-warn)', fontSize: 11 }}>R-07</span> · Non-reversible DB migrations require human</span>
                <span className="gw-badge warn"><Icon name="user" size={11} sw={2} />Require human</span>
              </div>
            </div>
          </div>

          <div className="gw-appr-sect" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <div>
              <div className="gw-sect-label">Risk explanation</div>
              <p className="gw-muted" style={{ fontSize: 12.5, lineHeight: 1.55 }}>
                Score driven by money-path code (+28), non-reversible migration (+22), and absence of a
                tested rollback (+22). Validations on worker logic pass, but the migration dry-run is blocked.
              </p>
            </div>
            <div>
              <div className="gw-sect-label">Agent recommendation</div>
              <p className="gw-muted" style={{ fontSize: 12.5, lineHeight: 1.55 }}>
                <span style={{ color: 'var(--gw-fg)', fontWeight: 600 }}>Hold.</span> {a.recommend} The agent has
                attached a migration plan and can produce a reversible variant on request.
              </p>
            </div>
          </div>

          <div className="gw-appr-actions">
            <Btn icon="check" variant="primary">Approve once</Btn>
            <Btn icon="policies">Approve + suggest rule</Btn>
            <Btn icon="message">Ask agent to clarify</Btn>
            <div className="gw-spacer" />
            <Btn icon="x" variant="danger">Reject</Btn>
          </div>
        </div>
      </div>
    </Shell>
  );
}

function ScreenApprovalsM() {
  const a = window.GW_DATA.approvals[0];
  return (
    <MobileShell active="approvals" title="Approvals">
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div className="gw-between"><span className="gw-mono" style={{ fontSize: 10, letterSpacing: '.1em', textTransform: 'uppercase', color: 'var(--gw-bad)' }}>High risk · 1</span><span className="gw-badge bad"><i className="bdot" />3 pending</span></div>

        <div className="gw-appr flag-bad">
          <div className="gw-appr-top">
            <div className="gw-attn-ic bad" style={{ flex: '0 0 28px' }}><Icon name="alert" size={15} /></div>
            <div>
              <div style={{ display: 'flex', gap: 8, marginBottom: 4 }}><RiskBadge score={72} showBar={false} /><span className="gw-id">{a.ticket}</span></div>
              <div className="ttl" style={{ fontSize: 13 }}>{a.action}</div>
              <div className="why">{a.why}</div>
            </div>
          </div>
          <div className="gw-appr-sect">
            <div className="gw-sect-label">Command</div>
            <CommandBlock><div><span className="prompt">$</span> psql -f <span className="str">0042_ledger.sql</span></div></CommandBlock>
          </div>
          <div className="gw-appr-sect">
            <div className="gw-sect-label">Matched rule</div>
            <div className="gw-between" style={{ padding: '8px 10px', background: 'var(--gw-bad-bg)', border: '1px solid var(--gw-bad-bd)', borderRadius: 6 }}>
              <span style={{ fontSize: 12 }}><span className="gw-mono" style={{ fontSize: 10.5 }}>R-01</span> billing/** → human</span>
              <Icon name="user" size={13} style={{ color: 'var(--gw-bad)' }} />
            </div>
          </div>
          <div className="gw-appr-actions" style={{ flexDirection: 'column' }}>
            <Btn icon="check" variant="primary" style={{ width: '100%', justifyContent: 'center' }}>Approve once</Btn>
            <div style={{ display: 'flex', gap: 8, width: '100%' }}>
              <Btn icon="message" style={{ flex: 1, justifyContent: 'center' }}>Clarify</Btn>
              <Btn icon="x" variant="danger" style={{ flex: 1, justifyContent: 'center' }}>Reject</Btn>
            </div>
          </div>
        </div>
      </div>
    </MobileShell>
  );
}

Object.assign(window, { ScreenApprovals, ScreenApprovalsM });
