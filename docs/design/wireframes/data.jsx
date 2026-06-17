// Groundwork — shared mock data for all screens. Plain JS, assigns to window.
window.GW_DATA = (function () {
  const tickets = [
    { id: 'GW-318', title: 'Migrate deploy pipeline off legacy Capistrano scripts', status: 'in_progress', col: 'in_progress', agent: 'codex', labels: ['infra', 'deploy'], risk: 58, val: 'pending', updated: '2m', run: 'run_8f2a' },
    { id: 'GW-311', title: 'Add rate-limit middleware to public ingest API', status: 'review', col: 'review', agent: 'codex-2', labels: ['api', 'go'], risk: 41, val: 'pass', approval: true, updated: '14m', run: 'run_7c19' },
    { id: 'GW-305', title: 'Bump axios 0.27 → 1.7 across web workspace', status: 'blocked', col: 'blocked', agent: 'codex-3', labels: ['deps', 'web'], risk: 63, val: 'fail', blocked: true, updated: '6m', run: 'run_7b04' },
    { id: 'GW-322', title: 'Document the .groundwork/ ticket schema in AGENTS.md', status: 'approved', col: 'approved', agent: 'codex', labels: ['docs'], risk: 8, val: 'pass', updated: '1m', run: 'run_8f55' },
    { id: 'GW-298', title: 'Rewrite billing reconciliation job as idempotent worker', status: 'in_progress', col: 'in_progress', agent: 'codex-2', labels: ['billing', 'go'], risk: 72, val: 'pending', approval: true, updated: '4m', run: 'run_8a90' },
    { id: 'GW-330', title: 'Flaky e2e: checkout spinner never resolves on retry', status: 'todo', col: 'todo', agent: null, labels: ['bug', 'web'], risk: 35, val: 'none', updated: '40m' },
    { id: 'GW-331', title: 'Extract shared validation schema package', status: 'todo', col: 'todo', agent: null, labels: ['refactor'], risk: 22, val: 'none', updated: '1h' },
    { id: 'GW-280', title: 'Land: rate-limit middleware to main', status: 'landing', col: 'landing', agent: 'codex-2', labels: ['api'], risk: 44, val: 'pass', approval: true, updated: '3m', run: 'run_7c19' },
    { id: 'GW-294', title: 'Replace cron health-check with lease heartbeat', status: 'rework', col: 'rework', agent: 'codex-3', labels: ['infra'], risk: 51, val: 'fail', updated: '22m', run: 'run_79c2' },
    { id: 'GW-260', title: 'Add structured logging to ingest workers', status: 'done', col: 'done', agent: 'codex', labels: ['go', 'obs'], risk: 18, val: 'pass', updated: '3h' },
    { id: 'GW-256', title: 'Tighten SQL escaping in legacy report builder', status: 'done', col: 'done', agent: 'codex-2', labels: ['security'], risk: 29, val: 'pass', updated: '5h' },
    { id: 'GW-340', title: 'Investigate p95 latency regression on /search', status: 'backlog', col: 'backlog', agent: null, labels: ['perf'], risk: 0, val: 'none', updated: '2d' },
    { id: 'GW-341', title: 'Spike: agent-authored migration safety checks', status: 'backlog', col: 'backlog', agent: null, labels: ['spike'], risk: 0, val: 'none', updated: '2d' },
  ];

  const runs = [
    { id: 'run_8f2a', ticket: 'GW-318', title: 'Migrate deploy pipeline off legacy Capistrano', agent: 'codex', status: 'running', step: 'Editing deploy/release.sh (4 / 7)', elapsed: '6m 12s', last: 'wrote 38 lines', risk: 58, val: 'pending' },
    { id: 'run_8a90', ticket: 'GW-298', title: 'Idempotent billing reconciliation worker', agent: 'codex-2', status: 'running', step: 'Awaiting approval · write billing/*', elapsed: '11m 48s', last: 'requested approval', risk: 72, val: 'pending', blocked: true },
    { id: 'run_7c19', ticket: 'GW-311', title: 'Rate-limit middleware for ingest API', agent: 'codex-2', status: 'review', step: 'Validations complete', elapsed: '18m 03s', last: 'ready for review', risk: 41, val: 'pass' },
    { id: 'run_7b04', ticket: 'GW-305', title: 'Bump axios across web workspace', agent: 'codex-3', status: 'blocked', step: 'Blocked · 14 type errors', elapsed: '9m 31s', last: 'tsc failed', risk: 63, val: 'fail', blocked: true },
    { id: 'run_79c2', ticket: 'GW-294', title: 'Lease heartbeat health-check', agent: 'codex-3', status: 'paused', step: 'Paused by operator', elapsed: '24m 10s', last: 'paused', risk: 51, val: 'fail' },
  ];

  const approvals = [
    {
      id: 'apr_91', tier: 'high', ticket: 'GW-298', run: 'run_8a90', agent: 'codex-2',
      action: 'Write to billing/reconcile/*.go and run migration 0042',
      why: 'Worker rewrite touches money-moving code and adds a non-reversible DB migration.',
      risk: 72, flag: 'flag-bad', kind: 'fs+db',
      cmd: 'write', recommend: 'Hold — request a dry-run migration plan first.',
    },
    {
      id: 'apr_88', tier: 'medium', ticket: 'GW-305', run: 'run_7b04', agent: 'codex-3',
      action: 'Update axios 0.27.2 → 1.7.4 in web/package.json',
      why: 'Major-version dependency bump with known breaking changes to request config.',
      risk: 63, flag: 'flag-warn', kind: 'deps',
      cmd: 'pnpm', recommend: 'Approve with a pinned range and require web e2e to pass.',
    },
    {
      id: 'apr_84', tier: 'medium', ticket: 'GW-280', run: 'run_7c19', agent: 'codex-2',
      action: 'Land branch agent/GW-311 → main (fast-forward, 6 commits)',
      why: 'Landing to the protected main branch always requires human approval in v1.',
      risk: 44, flag: 'flag-warn', kind: 'git',
      cmd: 'git', recommend: 'Approve — validations green, diff reviewed.',
    },
  ];

  const events = [
    { tone: 'bad', text: <><b>codex-2</b> requested approval to write <span className="gw-mono">billing/reconcile/*</span></>, time: '14:32:08 · GW-298' },
    { tone: 'bad', text: <><b>run_7b04</b> blocked — <span className="gw-mono">tsc</span> reported 14 type errors</>, time: '14:30:51 · GW-305' },
    { tone: 'ok', text: <><b>codex</b> auto-approved docs edit to <span className="gw-mono">AGENTS.md</span></>, time: '14:29:14 · GW-322' },
    { tone: 'run', text: <><b>codex-2</b> finished validations · 32 / 32 checks passing</>, time: '14:27:40 · GW-311' },
    { tone: 'warn', text: <>Lease for <b>run_79c2</b> renewed (stale 38s)</>, time: '14:24:02 · GW-294' },
    { tone: 'ok', text: <><b>codex</b> opened worktree <span className="gw-mono">.wt/GW-318</span></>, time: '14:18:33 · GW-318' },
  ];

  return { tickets, runs, approvals, events, repo: 'orchard-platform', branch: 'main' };
})();
