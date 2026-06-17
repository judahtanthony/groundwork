// Groundwork — Reusable component inventory specimen sheet.

function Spec({ name, note, children, wide }) {
  return (
    <div style={{ gridColumn: wide ? '1 / -1' : 'auto', border: '1px solid var(--gw-border)', borderRadius: 8, background: 'var(--gw-surface)', overflow: 'hidden' }}>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: 10, padding: '10px 14px', borderBottom: '1px solid var(--gw-border)', background: 'var(--gw-surface-2)' }}>
        <span className="gw-mono" style={{ fontSize: 11, fontWeight: 600, letterSpacing: '.04em' }}>{name}</span>
        <span className="gw-anno" style={{ fontSize: 11.5 }}>{note}</span>
      </div>
      <div style={{ padding: 16 }}>{children}</div>
    </div>
  );
}

function ScreenComponents() {
  const t = window.GW_DATA.tickets[1];
  return (
    <div className="gw" style={{ background: 'var(--gw-bg)', minHeight: '100%', padding: 24 }}>
      <div style={{ marginBottom: 18 }}>
        <div className="gw-eyebrow" style={{ marginBottom: 6 }}>Groundwork · design system</div>
        <h2 style={{ fontSize: 19, fontWeight: 600 }}>Reusable component inventory</h2>
        <div className="gw-page-sub">Every operational object below is shared across the seven screens. States are color-coded once: green pass/approved · amber waiting/risk · red blocked/failing · blue running · gray idle.</div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14, alignItems: 'start' }}>
        <Spec name="StatusBadge" note="run state, never severity">
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            <StatusBadge status="running" /><StatusBadge status="blocked" /><StatusBadge status="review" />
            <StatusBadge status="approved" /><StatusBadge status="landing" /><StatusBadge status="queued" /><StatusBadge status="done" />
          </div>
        </Spec>

        <Spec name="RiskBadge" note="0–100 · low / med / high band">
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, alignItems: 'center' }}>
            <RiskBadge score={8} /><RiskBadge score={41} /><RiskBadge score={63} /><RiskBadge score={72} />
            <span style={{ marginLeft: 8 }} /><RiskBadge score={18} showBar={false} /><RiskBadge score={51} showBar={false} /><RiskBadge score={88} showBar={false} />
          </div>
        </Spec>

        <Spec name="ValidationBadge" note="check / clock / x glyph">
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            <ValBadge state="pass" count="32" /><ValBadge state="pending" /><ValBadge state="fail" /><ValBadge state="none" />
          </div>
        </Spec>

        <Spec name="Chip · Agent" note="labels + assignee token">
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, alignItems: 'center' }}>
            <Chip>api</Chip><Chip>go</Chip><Chip>deps</Chip><Chip>docs</Chip>
            <span style={{ width: 10 }} /><Agent id="codex" /><Agent id="codex-2" /><Agent id="reviewer" name="reviewer" />
          </div>
        </Spec>

        <Spec name="TicketCard" note="board + list item · risk, validation, blocked/approval marks">
          <div style={{ maxWidth: 230 }}><TicketCard t={t} /></div>
        </Spec>

        <Spec name="TimelineEvent" note="node tone + text + mono time">
          <div className="gw-tl">
            <TlEvent tone="ok" text={<><b>codex</b> auto-approved docs edit</>} time="14:29:14 · GW-322" />
            <TlEvent tone="bad" text={<>Run blocked · type errors</>} time="14:30:51 · GW-305" />
            <TlEvent tone="accent" text={<>Status moved <b>todo → in_progress</b></>} time="14:08:44" last />
          </div>
        </Spec>

        <Spec name="DiffSummary" note="files · adds/dels · per-file list" wide>
          <div style={{ maxWidth: 460 }}>
            <DiffSummary files={6} add={214} del={47} list={[
              { path: 'api/ingest/middleware/ratelimit.go', add: 132, del: 0 },
              { path: 'api/ingest/router.go', add: 18, del: 11 },
            ]} />
          </div>
        </Spec>

        <Spec name="RunRow" note="dense table row · the Dashboard / Runs primitive" wide>
          <table className="gw-table" style={{ border: '1px solid var(--gw-border)', borderRadius: 8, overflow: 'hidden' }}>
            <thead><tr><th>Ticket</th><th>Agent</th><th>Status</th><th>Step</th><th>Elapsed</th><th>Risk</th><th>Validation</th></tr></thead>
            <tbody>
              <tr>
                <td><div className="gw-cell-title">Rate-limit middleware</div><div className="gw-id">GW-311 · run_7c19</div></td>
                <td><Agent id="codex-2" /></td><td><StatusBadge status="review" /></td>
                <td className="gw-muted" style={{ fontSize: 12 }}>Validations complete</td><td className="num">18m 03s</td>
                <td><RiskBadge score={41} /></td><td><ValBadge state="pass" /></td>
              </tr>
            </tbody>
          </table>
        </Spec>

        <Spec name="CommandApprovalBlock" note="terminal command + scope rows" wide>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
            <CommandBlock>
              <div><span className="prompt">codex-2 $</span> pnpm up axios <span className="flag">--latest</span></div>
            </CommandBlock>
            <div className="gw-scope">
              <ScopeRow k="files">web/package.json</ScopeRow>
              <ScopeRow k="network"><span style={{ color: 'var(--gw-warn)' }}>registry.npmjs.org</span></ScopeRow>
              <ScopeRow k="sandbox">workspace-write</ScopeRow>
            </div>
          </div>
        </Spec>

        <Spec name="ApprovalCard" note="action · reason · matched rule · decision">
          <div className="gw-appr flag-warn">
            <div className="gw-appr-top">
              <div className="gw-attn-ic warn" style={{ flex: '0 0 28px' }}><Icon name="approvals" size={15} /></div>
              <div><div className="ttl" style={{ fontSize: 13 }}>Land branch → main</div><div className="why">Protected branch · human-required in v1</div></div>
            </div>
            <div className="gw-appr-actions"><Btn icon="check" variant="primary" sm>Approve</Btn><Btn icon="x" variant="danger" sm>Reject</Btn></div>
          </div>
        </Spec>

        <Spec name="PolicyRuleRow" note="condition → mode + toggle">
          <div style={{ border: '1px solid var(--gw-border)', borderRadius: 8, overflow: 'hidden' }}>
            <PolicyRow cond={<><span className="op">when</span> edit <span className="gw-mono">**/*.md</span></>} mode="auto" on />
            <PolicyRow cond={<><span className="op">when</span> write <span className="gw-mono">billing/**</span></>} mode="human" on />
          </div>
        </Spec>
      </div>
    </div>
  );
}

Object.assign(window, { ScreenComponents });
