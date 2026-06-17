// Groundwork — Ticket Detail screen (desktop + narrow). Manage one work item.

function ScreenTicket() {
  return (
    <Shell active="tickets"
      crumbs={['orchard-platform', 'Tickets', 'GW-311']}
      topActions={<><IconBtn icon="history" title="History" /><IconBtn icon="more" title="More" /></>}>

      <div className="gw-page-head">
        <div style={{ minWidth: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 6 }}>
            <span className="gw-id" style={{ fontSize: 12.5 }}>GW-311</span>
            <StatusBadge status="review" />
            <span className="gw-badge warn"><Icon name="approvals" size={11} sw={2} />Approval requested</span>
          </div>
          <h2 style={{ fontSize: 21 }}>Add rate-limit middleware to public ingest API</h2>
        </div>
        <div className="gw-head-actions">
          <Btn icon="undo">Request rework</Btn>
          <Btn icon="approvals" variant="primary">Approve for landing</Btn>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 320px', gap: 16, alignItems: 'start' }}>
        {/* main column */}
        <div className="gw-stack">
          <div className="gw-panel">
            <div className="gw-panel-body">
              <div className="gw-sect-label">Problem</div>
              <p className="gw-muted" style={{ fontSize: 13.5, lineHeight: 1.6 }}>
                The public ingest API has no per-client rate limiting. A single misbehaving integration
                can saturate the worker pool and starve billing reconciliation. Add token-bucket
                middleware keyed by API client, with limits sourced from the existing plan tiers.
              </p>
            </div>
          </div>

          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="check" size={15} style={{ color: 'var(--gw-ok)' }} /><h3>Acceptance criteria</h3><div className="gw-spacer" /><span className="gw-id">3 / 4</span></div>
            <div className="gw-panel-body" style={{ paddingTop: 4, paddingBottom: 4 }}>
              <Check done>Token-bucket limiter applied to all <span className="gw-mono">/ingest</span> routes</Check>
              <Check done>Limits read from plan tier, not hard-coded</Check>
              <Check done>429 response includes <span className="gw-mono">Retry-After</span> header</Check>
              <Check pending>Load test confirms no regression at p95 under 2× burst</Check>
            </div>
          </div>

          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="shield" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Validation requirements</h3><div className="gw-spacer" /><ValBadge state="pass" count="32" /></div>
            <div className="gw-panel-body" style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <div className="gw-between"><span className="gw-mono" style={{ fontSize: 12 }}>go vet ./...</span><span className="gw-badge ok"><Icon name="check" size={11} sw={2.4} />pass</span></div>
              <div className="gw-between"><span className="gw-mono" style={{ fontSize: 12 }}>go test ./api/ingest/...</span><span className="gw-badge ok"><Icon name="check" size={11} sw={2.4} />28 pass</span></div>
              <div className="gw-between"><span className="gw-mono" style={{ fontSize: 12 }}>golangci-lint run</span><span className="gw-badge ok"><Icon name="check" size={11} sw={2.4} />pass</span></div>
              <div className="gw-between"><span className="gw-mono" style={{ fontSize: 12 }}>load-test:ingest (gate)</span><span className="gw-badge warn"><Icon name="clock" size={11} />queued</span></div>
            </div>
          </div>

          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="layers" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Current diff</h3><div className="gw-spacer" /><span className="gw-id">run_7c19</span></div>
            <div className="gw-panel-body">
              <DiffSummary files={6} add={214} del={47} list={[
                { path: 'api/ingest/middleware/ratelimit.go', add: 132, del: 0 },
                { path: 'api/ingest/router.go', add: 18, del: 11 },
                { path: 'api/plan/tiers.go', add: 24, del: 6 },
                { path: 'api/ingest/ratelimit_test.go', add: 40, del: 0 },
              ]} />
            </div>
          </div>

          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="history" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Timeline</h3></div>
            <div className="gw-panel-body">
              <div className="gw-tl">
                <TlEvent tone="warn" text={<><b>codex-2</b> requested approval to land → main</>} time="14:31:55" />
                <TlEvent tone="ok" text={<>Validations passed · 32 / 32 checks</>} time="14:27:40" />
                <TlEvent tone="run" text={<><b>codex-2</b> opened run <span className="gw-mono">run_7c19</span></>} time="14:09:02" />
                <TlEvent tone="accent" text={<>Status moved <b>todo → in_progress</b></>} time="14:08:44" />
                <TlEvent tone="" text={<>Ticket created from backlog by operator</>} time="13:52:10" last />
              </div>
            </div>
          </div>
        </div>

        {/* right rail */}
        <div className="gw-stack">
          <div className="gw-panel">
            <div className="gw-panel-body" style={{ display: 'flex', flexDirection: 'column', gap: 13 }}>
              <Field label="Status"><StatusBadge status="review" /></Field>
              <Field label="Priority"><span className="gw-badge warn"><Icon name="chevronD" size={11} style={{ transform: 'rotate(180deg)' }} />High</span></Field>
              <Field label="Assignee"><Agent id="codex-2" /></Field>
              <Field label="Risk"><RiskBadge score={41} /></Field>
              <Field label="Validation"><ValBadge state="pass" count="32" /></Field>
              <Field label="Labels"><span style={{ display: 'flex', gap: 5 }}><Chip>api</Chip><Chip>go</Chip></span></Field>
              <Field label="Worktree"><span className="gw-mono" style={{ fontSize: 11 }}>.wt/GW-311</span></Field>
            </div>
          </div>

          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="zap" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Actions</h3></div>
            <div className="gw-panel-body" style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <Btn icon="play" style={{ justifyContent: 'flex-start', width: '100%' }}>Start new run</Btn>
              <Btn icon="pause" style={{ justifyContent: 'flex-start', width: '100%' }}>Pause run</Btn>
              <Btn icon="user" style={{ justifyContent: 'flex-start', width: '100%' }}>Reassign agent</Btn>
              <Btn icon="flow" style={{ justifyContent: 'flex-start', width: '100%' }}>Transition status…</Btn>
              <div className="gw-divline" style={{ margin: '3px 0' }} />
              <Btn icon="approvals" variant="primary" style={{ justifyContent: 'flex-start', width: '100%' }}>Approve for landing</Btn>
              <Btn icon="x" variant="danger" style={{ justifyContent: 'flex-start', width: '100%' }}>Cancel ticket</Btn>
            </div>
          </div>

          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="terminal" size={14} style={{ color: 'var(--gw-fg-muted)' }} /><h3>CLI</h3></div>
            <div className="gw-panel-body">
              <div className="gw-mono gw-subtle" style={{ fontSize: 11, lineHeight: 1.7 }}>
                <div><span style={{ color: 'var(--gw-fg-subtle)' }}>$</span> gw ticket show GW-311</div>
                <div><span style={{ color: 'var(--gw-fg-subtle)' }}>$</span> gw run start GW-311</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Shell>
  );
}

function Field({ label, children }) {
  return (
    <div className="gw-between" style={{ alignItems: 'center' }}>
      <span className="gw-mono" style={{ fontSize: 10.5, letterSpacing: '.1em', textTransform: 'uppercase', color: 'var(--gw-fg-subtle)' }}>{label}</span>
      {children}
    </div>
  );
}

function ScreenTicketM() {
  return (
    <MobileShell active="board" title="GW-311">
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div>
          <div style={{ display: 'flex', gap: 8, marginBottom: 8 }}><StatusBadge status="review" /><span className="gw-badge warn"><Icon name="approvals" size={11} sw={2} />Approval</span></div>
          <h2 style={{ fontSize: 17, lineHeight: 1.3 }}>Add rate-limit middleware to public ingest API</h2>
        </div>

        <div className="gw-panel" style={{ padding: '12px 14px', display: 'flex', flexWrap: 'wrap', gap: 14 }}>
          <MiniField label="Assignee"><Agent id="codex-2" /></MiniField>
          <MiniField label="Risk"><RiskBadge score={41} showBar={false} /></MiniField>
          <MiniField label="Validation"><ValBadge state="pass" /></MiniField>
        </div>

        <div className="gw-panel">
          <div className="gw-panel-head"><h3>Acceptance criteria</h3><div className="gw-spacer" /><span className="gw-id">3 / 4</span></div>
          <div className="gw-panel-body" style={{ paddingTop: 2, paddingBottom: 2 }}>
            <Check done>Token-bucket limiter on all routes</Check>
            <Check done>Limits read from plan tier</Check>
            <Check pending>Load test confirms no p95 regression</Check>
          </div>
        </div>

        <div className="gw-panel">
          <div className="gw-panel-head"><h3>Current diff</h3><div className="gw-spacer" /><span className="gw-id">run_7c19</span></div>
          <div className="gw-panel-body"><DiffSummary files={6} add={214} del={47} /></div>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <Btn icon="approvals" variant="primary" style={{ width: '100%', justifyContent: 'center' }}>Approve for landing</Btn>
          <div style={{ display: 'flex', gap: 8 }}>
            <Btn icon="undo" style={{ flex: 1, justifyContent: 'center' }}>Rework</Btn>
            <Btn icon="pause" style={{ flex: 1, justifyContent: 'center' }}>Pause</Btn>
          </div>
        </div>
      </div>
    </MobileShell>
  );
}

function MiniField({ label, children }) {
  return <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}><span className="gw-mono" style={{ fontSize: 9.5, letterSpacing: '.1em', textTransform: 'uppercase', color: 'var(--gw-fg-subtle)' }}>{label}</span>{children}</div>;
}

Object.assign(window, { ScreenTicket, ScreenTicketM, Field });
