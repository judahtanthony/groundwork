(function(){const o=document.createElement("link").relList;if(o&&o.supports&&o.supports("modulepreload"))return;for(const e of document.querySelectorAll('link[rel="modulepreload"]'))i(e);new MutationObserver(e=>{for(const t of e)if(t.type==="childList")for(const n of t.addedNodes)n.tagName==="LINK"&&n.rel==="modulepreload"&&i(n)}).observe(document,{childList:!0,subtree:!0});function l(e){const t={};return e.integrity&&(t.integrity=e.integrity),e.referrerPolicy&&(t.referrerPolicy=e.referrerPolicy),e.crossOrigin==="use-credentials"?t.credentials="include":e.crossOrigin==="anonymous"?t.credentials="omit":t.credentials="same-origin",t}function i(e){if(e.ep)return;e.ep=!0;const t=l(e);fetch(e.href,t)}})();const d=document.querySelector("#app");if(!d)throw new Error("missing #app mount point");d.innerHTML=`
  <main class="shell">
    <header>
      <p class="eyebrow">Local operator UI</p>
      <h1>Groundwork</h1>
      <p class="lede">The embedded SPA is ready for the root-centric operator surfaces.</p>
    </header>
    <section aria-labelledby="coordinator-heading">
      <div>
        <p class="eyebrow">Coordinator</p>
        <h2 id="coordinator-heading">Connecting…</h2>
        <p id="state-detail" class="detail" aria-live="polite">Reading same-origin coordinator state.</p>
      </div>
      <span id="state-dot" class="status" aria-hidden="true"></span>
    </section>
    <nav aria-label="Current operator surfaces">
      <a href="/">Dashboard</a>
      <a href="/tickets">Tickets</a>
      <a href="/approvals">Approvals</a>
    </nav>
  </main>
`;const a=document.querySelector("#coordinator-heading"),s=document.querySelector("#state-detail"),c=document.querySelector("#state-dot");async function u(){try{const r=await fetch("/api/v1/state",{headers:{Accept:"application/json"}});if(!r.ok)throw new Error(`HTTP ${r.status}`);const o=await r.json();a&&(a.textContent=o.ok?"Connected":"Unavailable"),s&&(s.textContent=`${o.total} nodes · ${o.eligible} ready · gw ${o.version}`),c?.classList.add(o.ok?"status--ok":"status--bad")}catch(r){a&&(a.textContent="Unavailable"),s&&(s.textContent=r instanceof Error?r.message:"Could not read coordinator state."),c?.classList.add("status--bad")}}u();
