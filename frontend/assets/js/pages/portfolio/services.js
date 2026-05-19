/* ============================================================
 * FILE: pages/portfolio/services.js
 * WHAT IT IS:     "What I Offer" section loader
 * WHERE USED:     main.js — fetches services from API on load
 * LAST UPDATED:   2026-05-19 — dynamic API fetch replacing static data
 * ============================================================ */

const ServicesSection = (() => {

  async function init() {
    const container = document.getElementById('services-grid');
    if (!container) return;

    try {
      const res = await API.get('/services');
      const services = res.data || [];

      if (!services.length) {
        container.innerHTML = '';
        return;
      }

      container.innerHTML = services.map(s => `
        <div class="card scroll-animation" data-animation="fade_from_bottom">
          <div class="card-icon">${s.icon || '🛠'}</div>
          <h3 class="card-title">${s.title}</h3>
          <p class="card-desc">${s.description || ''}</p>
        </div>
      `).join('');
    } catch (err) {
      // Silently hide the grid on error — section heading still shows
      container.innerHTML = '';
    }
  }

  return { init };
})();
