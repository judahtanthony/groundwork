// Groundwork — Run Detail screen (desktop + narrow). Inspect one agent run.

function ScreenRun() {
  return (
    <Shell active="runs"
      crumbs={['orchard-platform', 'Runs', 'run_8a90']}
      topActions={<><IconBtn icon="download" title="Export transcript" /><IconBtn icon="more" title="More" /></>}>

      {/* run header */}
      <div className="gw-panel" style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16, padding: '14px 16px', flexWrap: 'wrap' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 11 }}>
            <span className="gw-id" style={{ fontSize: 13 }}>run_8a90</span>
            <StatusBadge status="blocked" label="Awaiting approval" />
            <RiskBadge score={72} />
          </div>
          <div className="gw-divline" style={{ width: 1, height: 24, alignSelf: 'center' }} />
          <RunMeta k="Ticket" mono="GW-298" link />
          <RunMeta k="Agent" node={<Agent id="codex-2" />} />
          <RunMeta k="Workspace" mono=".wt/GW-298" />
          <RunMeta k="Base" mono="a1b2c3d" />
          <RunMeta k="Elapsed" mono="11m 48s" />
          <div className="gw-spacer" />
          <div style={{ display: 'flex', gap: 8 }}>
            <Btn icon="pause" sm>Pause</Btn>
            <Btn icon="play" sm>Resume</Btn>
            <Btn icon="folder" sm>Open workspace</Btn>
            <Btn icon="copy" sm>Copy resume cmd</Btn>
            <Btn icon="stop" variant="danger" sm>Cancel</Btn>
          </div>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1.45fr 1fr', gap: 16, alignItems: 'start' }}>
        {/* left: transcript + events */}
        <div className="gw-stack">
          <div className="gw-panel" style={{ overflow: 'hidden' }}>
            <div className="gw-panel-head">
              <Icon name="terminal" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Live transcript</h3>
              <span className="gw-badge run" style={{ marginLeft: 2 }}><i className="bdot" />streaming</span>
              <div className="gw-spacer" /><span className="gw-id">tail · follow</span>
            </div>
            <div className="gw-transcript" style={{ border: 'none', borderRadius: 0 }}>
              <div className="gw-transcript-body">
                <TrLine t="14:20:02" who="agent" cls="agent" text="Planning idempotent reconciliation worker. 5 steps." />
                <TrLine t="14:21:14" who="tool" cls="tool" text="read billing/reconcile/legacy_job.go (412 lines)" />
                <TrLine t="14:23:40" who="agent" cls="agent" text="Refactoring into worker + dedup ledger keyed by (invoice_id, period)." />
                <TrLine t="14:26:11" who="tool" cls="tool" text="write billing/reconcile/worker.go (+186 −0)" />
                <TrLine t="14:29:03" who="agent" cls="agent" text="Migration 0042 needed to add reconcile_ledger table + unique index." />
                <TrLine t="14:31:52" who="sys" cls="sys" text="policy: write to billing/* + DB migration → requires human approval" />
                <TrLine t="14:31:55" who="agent" cls="agent" text="Pausing. Requested approval apr_91 with migration plan attached." />
                <div className="gw-tr-line" style={{ marginTop: 4 }}><span className="t" /><span style={{ color: '#c98f7e' }}>▍ blocked — awaiting operator decision</span></div>
              </div>
            </div>
          </div>

          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="history" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Event timeline</h3></div>
            <div className="gw-panel-body">
              <div className="gw-tl">
                <TlEvent tone="bad" text={<>Approval requested · <span className="gw-mono">billing/reconcile/*</span> + migration 0042</>} time="14:31:55" />
                <TlEvent tone="warn" text={<>Policy gate triggered · money-path + non-reversible DB change</>} time="14:31:52" />
                <TlEvent tone="run" text={<>Wrote <span className="gw-mono">worker.go</span> · +186 lines</>} time="14:26:11" />
                <TlEvent tone="run" text={<>Read <span className="gw-mono">legacy_job.go</span></>} time="14:21:14" />
                <TlEvent tone="accent" text={<>Plan accepted · 5 steps</>} time="14:20:02" />
                <TlEvent tone="" text={<>Run started from GW-298</>} time="14:19:40" last />
              </div>
            </div>
          </div>
        </div>

        {/* right: plan, diff, validation, metrics, approval */}
        <div className="gw-stack">
          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="check" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Plan</h3><div className="gw-spacer" /><span className="gw-id">3 / 5</span></div>
            <div className="gw-panel-body" style={{ paddingTop: 4, paddingBottom: 4 }}>
              <Check done>Read legacy reconciliation job</Check>
              <Check done>Design dedup ledger schema</Check>
              <Check done>Implement worker.go</Check>
              <Check pending>Apply migration 0042 (blocked on approval)</Check>
              <Check>Backfill + verify against prod sample</Check>
            </div>
          </div>

          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="layers" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Changed files</h3></div>
            <div className="gw-panel-body">
              <DiffSummary files={3} add={210} del={12} list={[
                { path: 'billing/reconcile/worker.go', add: 186, del: 0 },
                { path: 'billing/migrations/0042_ledger.sql', add: 24, del: 0 },
                { path: 'billing/reconcile/router.go', add: 0, del: 12 },
              ]} />
            </div>
          </div>

          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="shield" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Validations</h3><div className="gw-spacer" /><ValBadge state="pending" /></div>
            <div className="gw-panel-body" style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <div className="gw-between"><span className="gw-mono" style={{ fontSize: 12 }}>go test ./billing/...</span><span className="gw-badge ok"><Icon name="check" size={11} sw={2.4} />pass</span></div>
              <div className="gw-between"><span className="gw-mono" style={{ fontSize: 12 }}>migration dry-run</span><span className="gw-badge warn"><Icon name="clock" size={11} />blocked</span></div>
            </div>
          </div>

          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="gauge" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Metrics</h3></div>
            <div className="gw-metrics" style={{ border: 'none', borderRadius: 0 }}>
              <Metric label="Input tokens" value="412" unit="K" />
              <Metric label="Output tokens" value="88" unit="K" />
              <Metric label="Wall time" value="11m 48s" />
              <Metric label="Tool calls" value="23" />
            </div>
          </div>

          <div className="gw-appr flag-bad">
            <div className="gw-appr-top">
              <div className="gw-attn-ic bad" style={{ flex: '0 0 28px' }}><Icon name="shield" size={15} /></div>
              <div>
                <div className="ttl" style={{ fontSize: 13 }}>Linked approval · apr_91</div>
                <div className="why">Write to billing + migration 0042 · high risk</div>
              </div>
            </div>
            <div className="gw-appr-actions" style={{ borderTop: 'none' }}>
              <Btn icon="approvals" variant="primary" sm>Review approval</Btn>
            </div>
          </div>
        </div>
      </div>
    </Shell>
  );
}

function RunMeta({ k, mono, node, link }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
      <span className="gw-mono" style={{ fontSize: 9.5, letterSpacing: '.1em', textTransform: 'uppercase', color: 'var(--gw-fg-subtle)' }}>{k}</span>
      {node || <span className="gw-mono" style={{ fontSize: 12, color: link ? 'var(--gw-accent)' : 'var(--gw-fg)' }}>{mono}</span>}
    </div>
  );
}
function TrLine({ t, who, cls, text }) {
  return <div className="gw-tr-line"><span className="t">{t}</span><span><span className={'who ' + cls}>{who}</span> {text}</span></div>;
}

function ScreenRunM() {
  return (
    <MobileShell active="runs" title="run_8a90">
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div className="gw-panel" style={{ padding: '12px 14px' }}>
          <div style={{ display: 'flex', gap: 8, marginBottom: 9, flexWrap: 'wrap' }}><StatusBadge status="blocked" label="Awaiting approval" /><RiskBadge score={72} /></div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 14 }}>
            <MiniField label="Ticket"><span className="gw-mono" style={{ color: 'var(--gw-accent)', fontSize: 12 }}>GW-298</span></MiniField>
            <MiniField label="Agent"><Agent id="codex-2" /></MiniField>
            <MiniField label="Elapsed"><span className="gw-mono" style={{ fontSize: 12 }}>11m 48s</span></MiniField>
          </div>
        </div>

        <div className="gw-panel" style={{ overflow: 'hidden' }}>
          <div className="gw-panel-head"><Icon name="terminal" size={14} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Transcript</h3><span className="gw-badge run" style={{ marginLeft: 'auto' }}><i className="bdot" />live</span></div>
          <div className="gw-transcript" style={{ border: 'none', borderRadius: 0 }}>
            <div className="gw-transcript-body" style={{ fontSize: 11 }}>
              <TrLine t="14:26" who="tool" cls="tool" text="write worker.go (+186)" />
              <TrLine t="14:31" who="sys" cls="sys" text="policy: write billing/* → human approval" />
              <TrLine t="14:31" who="agent" cls="agent" text="Requested approval apr_91." />
              <div className="gw-tr-line" style={{ marginTop: 3 }}><span className="t" /><span style={{ color: '#c98f7e' }}>▍ blocked</span></div>
            </div>
          </div>
        </div>

        <div className="gw-panel">
          <div className="gw-panel-head"><h3>Plan</h3><div className="gw-spacer" /><span className="gw-id">3 / 5</span></div>
          <div className="gw-panel-body" style={{ paddingTop: 2, paddingBottom: 2 }}>
            <Check done>Implement worker.go</Check>
            <Check pending>Apply migration 0042 (blocked)</Check>
            <Check>Backfill + verify</Check>
          </div>
        </div>

        <div style={{ display: 'flex', gap: 8 }}>
          <Btn icon="approvals" variant="primary" style={{ flex: 1, justifyContent: 'center' }}>Review approval</Btn>
          <Btn icon="stop" variant="danger" style={{ justifyContent: 'center' }}>Cancel</Btn>
        </div>
      </div>
    </MobileShell>
  );
}

Object.assign(window, { ScreenRun, ScreenRunM, RunMeta, TrLine });
