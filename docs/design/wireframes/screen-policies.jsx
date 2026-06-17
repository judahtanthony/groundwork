// Groundwork — Policies screen. Manage trust, risk, and validation rules.

function ValTemplate({ icon, name, glob, checks, tone }) {
  return (
    <div className="gw-panel" style={{ padding: 0 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 9, padding: '11px 13px', borderBottom: '1px solid var(--gw-border)' }}>
        <div className={'gw-attn-ic ' + tone} style={{ width: 26, height: 26, flex: '0 0 26px', borderRadius: 6 }}><Icon name={icon} size={14} /></div>
        <div>
          <div style={{ fontSize: 12.5, fontWeight: 600 }}>{name}</div>
          <div className="gw-mono gw-subtle" style={{ fontSize: 10.5 }}>{glob}</div>
        </div>
      </div>
      <div style={{ padding: '9px 13px', display: 'flex', flexDirection: 'column', gap: 6 }}>
        {checks.map((c, i) => <div key={i} className="gw-mono" style={{ fontSize: 11, color: 'var(--gw-fg-muted)', display: 'flex', alignItems: 'center', gap: 7 }}><Icon name="check" size={11} sw={2.4} style={{ color: 'var(--gw-ok)', flex: '0 0 11px' }} />{c}</div>)}
      </div>
    </div>
  );
}

function ScreenPolicies() {
  return (
    <Shell active="policies"
      crumbs={['orchard-platform', 'Policies']}
      topActions={<><Btn icon="download" sm>Export policy</Btn><Btn icon="plus" variant="primary">New rule</Btn></>}>
      <div className="gw-page-head">
        <div><h2>Policies</h2><div className="gw-page-sub">Trust, risk thresholds, and validation gates · stored in <span className="gw-mono">.groundwork/policy.toml</span></div></div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 360px', gap: 16, alignItems: 'start' }}>
        <div className="gw-stack">
          {/* trust rules */}
          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="shield" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Trust rules</h3><div className="gw-spacer" /><span className="gw-id">7 active · evaluated top-down</span></div>
            <div>
              <PolicyRow cond={<><span className="op">when</span> write <span className="gw-mono">billing/**</span></>} mode="human" scope="any change · always escalate" on />
              <PolicyRow cond={<><span className="op">when</span> migration is non-reversible</>} mode="human" scope="DB schema · destructive" on />
              <PolicyRow cond={<><span className="op">when</span> land → <span className="gw-mono">main</span></>} mode="human" scope="protected branch · v1 default" on />
              <PolicyRow cond={<><span className="op">when</span> dependency major bump</>} mode="reviewer" scope="reviewer-agent then human if risk ≥ 60" on />
              <PolicyRow cond={<><span className="op">when</span> edit <span className="gw-mono">**/*.md</span> · risk &lt; 20</>} mode="auto" scope="docs · validations must pass" on />
              <PolicyRow cond={<><span className="op">when</span> add test files only</>} mode="auto" scope="*_test.go · no source change" on />
              <PolicyRow cond={<><span className="op">when</span> run shell <span className="gw-mono">rm -rf</span></>} mode="human" scope="destructive · blocked by default" on />
            </div>
          </div>

          {/* validation templates */}
          <div>
            <div className="gw-sect-label" style={{ marginBottom: 10 }}>Validation templates · by file type</div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
              <ValTemplate icon="doc" tone="ok" name="Documentation" glob="**/*.md, docs/**" checks={['markdownlint', 'link-check', 'spell (project dict)']} />
              <ValTemplate icon="terminal" tone="run" name="Go" glob="**/*.go" checks={['go vet ./...', 'go test ./...', 'golangci-lint run']} />
              <ValTemplate icon="globe" tone="warn" name="Web / TypeScript" glob="web/**/*.{ts,tsx}" checks={['tsc --noEmit', 'eslint', 'vitest run', 'e2e (gate)']} />
              <ValTemplate icon="settings" tone="idle" name="Config" glob="**/*.{toml,yaml,json}" checks={['schema validate', 'secret scan', 'diff review required']} />
            </div>
          </div>
        </div>

        {/* right rail: rule editor + suggestion queue */}
        <div className="gw-stack">
          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="edit" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Rule editor</h3><div className="gw-spacer" /><span className="gw-id">R-01</span></div>
            <div className="gw-panel-body" style={{ display: 'flex', flexDirection: 'column', gap: 13 }}>
              <EditField label="Match · path glob"><input className="gw-input" defaultValue="billing/**" readOnly /></EditField>
              <EditField label="Action">
                <div className="gw-seg" style={{ width: '100%' }}>
                  <button style={{ flex: 1 }}>Auto</button>
                  <button className="on" style={{ flex: 1 }}>Require human</button>
                  <button style={{ flex: 1 }}>Reviewer</button>
                </div>
              </EditField>
              <EditField label="Risk threshold"><input className="gw-input" defaultValue="≥ 0  (always)" readOnly /></EditField>
              <EditField label="Validation gate"><input className="gw-input" defaultValue="go test ./billing/..." readOnly /></EditField>
              <EditField label="On reject"><input className="gw-input" defaultValue="pause run · notify operator" readOnly /></EditField>
              <div style={{ display: 'flex', gap: 8, marginTop: 2 }}>
                <Btn variant="primary" sm style={{ flex: 1, justifyContent: 'center' }}>Save rule</Btn>
                <Btn icon="undo" sm>Reset</Btn>
              </div>
            </div>
          </div>

          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="zap" size={15} style={{ color: 'var(--gw-warn)' }} /><h3>Suggestion queue</h3><div className="gw-spacer" /><span className="gw-badge warn"><i className="bdot" />2</span></div>
            <div className="gw-panel-body" style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <div style={{ border: '1px solid var(--gw-border)', borderRadius: 8, padding: 12 }}>
                <div style={{ fontSize: 12.5, fontWeight: 600, marginBottom: 4 }}>Auto-approve docs edits under risk 20</div>
                <p className="gw-muted" style={{ fontSize: 11.5, lineHeight: 1.5, marginBottom: 8 }}>
                  You approved 6 similar <span className="gw-mono">*.md</span> edits this week. Promote to an auto rule?
                </p>
                <div className="gw-mono gw-subtle" style={{ fontSize: 10.5, marginBottom: 10 }}>examples · GW-322, GW-307, GW-301 · rollback: disable R-04, 1 click</div>
                <div style={{ display: 'flex', gap: 7 }}><Btn variant="primary" sm>Promote</Btn><Btn sm>Dismiss</Btn></div>
              </div>
              <div style={{ border: '1px solid var(--gw-border)', borderRadius: 8, padding: 12 }}>
                <div style={{ fontSize: 12.5, fontWeight: 600, marginBottom: 4 }}>Reviewer-agent for test-only changes</div>
                <p className="gw-muted" style={{ fontSize: 11.5, lineHeight: 1.5, marginBottom: 8 }}>
                  Repeated approvals on <span className="gw-mono">*_test.go</span>. Route to reviewer-agent first.
                </p>
                <div className="gw-mono gw-subtle" style={{ fontSize: 10.5, marginBottom: 10 }}>examples · GW-260, GW-256 · rollback: revert to manual</div>
                <div style={{ display: 'flex', gap: 7 }}><Btn variant="primary" sm>Promote</Btn><Btn sm>Dismiss</Btn></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Shell>
  );
}

function EditField({ label, children }) {
  return <div style={{ display: 'flex', flexDirection: 'column', gap: 5 }}><span className="gw-mono" style={{ fontSize: 10, letterSpacing: '.08em', textTransform: 'uppercase', color: 'var(--gw-fg-subtle)' }}>{label}</span>{children}</div>;
}

Object.assign(window, { ScreenPolicies, ValTemplate });
