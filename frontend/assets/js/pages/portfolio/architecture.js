/* ============================================================
 * FILE: pages/portfolio/architecture.js
 * WHAT IT IS:     Architecture Expertise section loader
 * WHERE USED:     main.js → called on DOMContentLoaded
 * LAST UPDATED:   2026-05-19 — dynamic API fetch
 * ============================================================ */

const ArchitectureSection = (() => {

  async function init() {
    const container = document.getElementById('architecture-grid');
    if (!container) return;

    try {
      const res = await API.get('/architecture');
      const items = res.data || [];
      if (!items.length) { container.innerHTML = ''; return; }
      container.innerHTML = `<div class="arch-grid">${items.map(buildCard).join('')}</div>`;
    } catch (err) {
      container.innerHTML = '';
    }
  }

  function buildCard(arch) {
    const projects = (arch.projects || [])
      .map(p => `<span class="arch-project-tag">${p}</span>`).join('');
    return `
      <div class="arch-card scroll-animation" data-animation="fade_from_bottom">
        <div class="arch-card-header">
          <div class="arch-card-num">${arch.number || ''}</div>
          <h3 class="arch-card-name">${arch.name}</h3>
        </div>
        <pre class="arch-diagram">${arch.diagram || ''}</pre>
        <p class="arch-desc">${arch.description || ''}</p>
        <div class="arch-projects">${projects}</div>
      </div>
    `;
  }

  return { init };
})();
