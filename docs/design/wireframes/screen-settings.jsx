// Groundwork — Settings screen. Configure local Groundwork instance.

function SettingsGroup({ title, icon, children }) {
  return (
    <div className="gw-panel">
      <div className="gw-panel-head"><Icon name={icon} size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>{title}</h3></div>
      <div style={{ padding: '4px 16px' }}>{children}</div>
    </div>
  );
}
function Setting({ label, hint, children }) {
  return (
    <div className="gw-field">
      <div className="gw-field-label"><div className="fl">{label}</div>{hint && <div className="fh">{hint}</div>}</div>
      <div className="gw-field-control">{children}</div>
    </div>
  );
}

function ScreenSettings() {
  return (
    <Shell active="settings"
      crumbs={['orchard-platform', 'Settings']}
      topActions={<><Btn sm>Discard</Btn><Btn variant="primary">Save changes</Btn></>}>
      <div className="gw-page-head">
        <div><h2>Settings</h2><div className="gw-page-sub">Local instance · written to <span className="gw-mono">.groundwork/config.toml</span></div></div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 320px', gap: 16, alignItems: 'start' }}>
        <div className="gw-stack">
          <SettingsGroup title="Repository" icon="folder">
            <Setting label="Repo root" hint="Absolute path to the git repository"><input className="gw-input ro" defaultValue="/Users/judah/code/orchard-platform" readOnly /></Setting>
            <Setting label=".groundwork/ path" hint="Tickets, policies, run logs"><input className="gw-input ro" defaultValue="./.groundwork" readOnly /></Setting>
            <Setting label="SQLite path" hint="Operational state store"><input className="gw-input ro" defaultValue="./.groundwork/state.sqlite" readOnly /></Setting>
            <Setting label="Main branch" hint="Protected landing target"><input className="gw-input" defaultValue="main" /></Setting>
          </SettingsGroup>

          <SettingsGroup title="Agent engine" icon="terminal">
            <Setting label="Codex command" hint="Binary invoked per run"><input className="gw-input" defaultValue="codex exec --json --cd {worktree}" /></Setting>
            <Setting label="Sandbox / approval mode" hint="Default permission posture for new runs">
              <div className="gw-seg"><button>read-only</button><button className="on">workspace-write</button><button>danger-full</button></div>
            </Setting>
          </SettingsGroup>

          <SettingsGroup title="Concurrency & leases" icon="cpu">
            <Setting label="Max concurrent agents" hint="Parallel runs across worktrees"><input className="gw-input" defaultValue="4" style={{ maxWidth: 120 }} /></Setting>
            <Setting label="Lease TTL" hint="Run claim expiry before reaped"><input className="gw-input" defaultValue="90s" style={{ maxWidth: 120 }} /></Setting>
            <Setting label="Renewal interval" hint="Heartbeat cadence"><input className="gw-input" defaultValue="30s" style={{ maxWidth: 120 }} /></Setting>
          </SettingsGroup>

          <SettingsGroup title="Server" icon="globe">
            <Setting label="Bind address" hint="Local-only by default"><input className="gw-input" defaultValue="127.0.0.1" style={{ maxWidth: 200 }} /></Setting>
            <Setting label="Port"><input className="gw-input" defaultValue="4500" style={{ maxWidth: 120 }} /></Setting>
            <Setting label="Export settings" hint="Bundle config + policy for sharing">
              <div style={{ display: 'flex', gap: 8 }}><Btn icon="download" sm>Export bundle</Btn><Btn icon="copy" sm>Copy config</Btn></div>
            </Setting>
          </SettingsGroup>
        </div>

        {/* right rail */}
        <div className="gw-stack">
          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="doc" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>AGENTS.md integration</h3></div>
            <div className="gw-panel-body" style={{ display: 'flex', flexDirection: 'column', gap: 11 }}>
              <div className="gw-between"><span className="gw-muted">Status</span><span className="gw-badge ok"><Icon name="check" size={11} sw={2.4} />Synced</span></div>
              <div className="gw-between"><span className="gw-muted">File</span><span className="gw-mono" style={{ fontSize: 11 }}>./AGENTS.md</span></div>
              <div className="gw-between"><span className="gw-muted">Last sync</span><span className="gw-mono" style={{ fontSize: 11 }}>2m ago</span></div>
              <p className="gw-subtle" style={{ fontSize: 11.5, lineHeight: 1.5 }}>Policy summary and ticket conventions are injected into agent context on each run.</p>
              <Btn icon="refresh" sm style={{ alignSelf: 'flex-start' }}>Re-sync now</Btn>
            </div>
          </div>

          <div className="gw-panel">
            <div className="gw-panel-head"><Icon name="gauge" size={15} style={{ color: 'var(--gw-fg-muted)' }} /><h3>Doctor</h3><div className="gw-spacer" /><span className="gw-badge ok"><i className="bdot" />healthy</span></div>
            <div className="gw-panel-body" style={{ display: 'flex', flexDirection: 'column', gap: 9 }}>
              <Health ok label="git repository" detail="clean · main @ a1b2c3d" />
              <Health ok label="state.sqlite" detail="readable · 2.4 MB · WAL on" />
              <Health ok label="codex binary" detail="v0.41 · on PATH" />
              <Health ok label="server bind" detail="127.0.0.1:4500 listening" />
              <Health warn label="worktree disk" detail="78% used · prune recommended" />
              <Health ok label="lease reaper" detail="running · 0 stale" />
              <div className="gw-divline" style={{ margin: '3px 0' }} />
              <Btn icon="refresh" sm style={{ alignSelf: 'flex-start' }}>Run gw doctor</Btn>
            </div>
          </div>
        </div>
      </div>
    </Shell>
  );
}

function Health({ ok, warn, label, detail }) {
  const tone = warn ? 'warn' : ok ? 'ok' : 'bad';
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 9 }}>
      <span style={{ display: 'flex', color: 'var(--gw-' + (warn ? 'warn' : ok ? 'ok' : 'bad') + ')' }}>
        <Icon name={warn ? 'alert' : ok ? 'check' : 'x'} size={14} sw={2.2} />
      </span>
      <span style={{ fontSize: 12.5, fontWeight: 500 }}>{label}</span>
      <span className="gw-mono gw-subtle" style={{ fontSize: 10.5, marginLeft: 'auto' }}>{detail}</span>
    </div>
  );
}

Object.assign(window, { ScreenSettings });
