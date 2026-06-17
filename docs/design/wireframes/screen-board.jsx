// Groundwork — Board screen. Ticket flow by state (kanban).

const GW_COLS = [
  { id: 'backlog', name: 'Backlog', dot: 'var(--gw-idle)' },
  { id: 'todo', name: 'Todo', dot: 'var(--gw-idle)' },
  { id: 'in_progress', name: 'In progress', dot: 'var(--gw-run)' },
  { id: 'blocked', name: 'Blocked', dot: 'var(--gw-bad)' },
  { id: 'review', name: 'Review', dot: 'var(--gw-warn)' },
  { id: 'rework', name: 'Rework', dot: 'var(--gw-warn)' },
  { id: 'approved', name: 'Approved', dot: 'var(--gw-ok)' },
  { id: 'landing', name: 'Landing', dot: 'var(--gw-run)' },
  { id: 'done', name: 'Done', dot: 'var(--gw-idle)' },
];

function TicketCard({ t }) {
  return (
    <div className={'gw-ticket' + (t.approval ? ' l-accent' : '')}>
      <div className="gw-ticket-top">
        <span className="gw-ticket-id">{t.id}</span>
        {t.blocked && <span title="Blocked" style={{ color: 'var(--gw-bad)', display: 'flex' }}><Icon name="alert" size={13} /></span>}
        {t.approval && <span title="Awaiting approval" style={{ color: 'var(--gw-accent)', display: 'flex' }}><Icon name="shield" size={13} /></span>}
        <span style={{ marginLeft: 'auto' }}><RiskBadge score={t.risk} showBar={false} /></span>
      </div>
      <div className="gw-ticket-title">{t.title}</div>
      <div className="gw-ticket-meta">
        {t.labels.map((l) => <Chip key={l}>{l}</Chip>)}
      </div>
      <div className="gw-ticket-foot">
        {t.agent ? <Agent id={t.agent} /> : <span className="gw-subtle" style={{ fontSize: 11.5, display: 'flex', alignItems: 'center', gap: 5 }}><Icon name="user" size={13} />Unassigned</span>}
        {t.val !== 'none' && <span style={{ display: 'flex' }} title={'Validations ' + t.val}>
          {t.val === 'pass' ? <Icon name="check" size={13} style={{ color: 'var(--gw-ok)' }} sw={2.4} />
            : t.val === 'fail' ? <Icon name="x" size={13} style={{ color: 'var(--gw-bad)' }} sw={2.4} />
              : <Icon name="clock" size={13} style={{ color: 'var(--gw-warn)' }} />}
        </span>}
        <span className="gw-ticket-time">{t.updated}</span>
      </div>
    </div>
  );
}

function ScreenBoard() {
  const D = window.GW_DATA;
  const byCol = (c) => D.tickets.filter((t) => t.col === c);
  return (
    <Shell active="board"
      crumbs={['orchard-platform', 'Board']}
      topActions={<>
        <div className="gw-seg"><button>All agents</button><button className="on">Active</button></div>
        <Btn icon="plus" variant="primary">New ticket</Btn>
      </>}>
      <div className="gw-page-head">
        <div>
          <h2>Board</h2>
          <div className="gw-page-sub">13 tickets across 9 states · drag to transition · double-click to open</div>
        </div>
        <div className="gw-head-actions">
          <Btn icon="board" sm>Group: status</Btn>
          <Btn icon="settings" sm>Columns</Btn>
        </div>
      </div>

      <div className="gw-board">
        {GW_COLS.map((c) => {
          const items = byCol(c.id);
          return (
            <div className="gw-col" key={c.id}>
              <div className="gw-col-head">
                <span className="gw-col-dot" style={{ background: c.dot }} />
                <span className="gw-col-name">{c.name}</span>
                <span className="gw-col-count" style={{ marginLeft: 'auto' }}>{items.length}</span>
              </div>
              <div className="gw-col-cards">
                {items.map((t) => <TicketCard key={t.id} t={t} />)}
                {items.length === 0 && <div style={{ padding: '14px 8px', textAlign: 'center', color: 'var(--gw-fg-subtle)', fontSize: 11.5, border: '1px dashed var(--gw-border-2)', borderRadius: 6 }}>Empty</div>}
              </div>
            </div>
          );
        })}
      </div>
    </Shell>
  );
}

Object.assign(window, { ScreenBoard, TicketCard, GW_COLS });
