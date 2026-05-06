/* ============================================================
 * FILE: pages/portfolio/apps.js
 * WHAT IT IS:     Apps Showcase section loader
 * WHERE USED:     main.js → called on DOMContentLoaded
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const AppsSection = (() => {

  /**
   * FUNCTION: init
   * WHAT IT DOES: Fetches projects from the API and renders the
   *               AppShowcase slider with phone mockups.
   */
  async function init() {
    const container = document.getElementById('apps-showcase');
    if (!container) return;

    try {
      const res = await API.get('/projects');
      const apps = res.data || [];

      // Sort: featured first, then by sort_order
      apps.sort((a, b) => {
        if (a.featured && !b.featured) return -1;
        if (!a.featured && b.featured) return 1;
        return a.sort_order - b.sort_order;
      });

      Store.set('projects', apps);

      AppShowcase.render({
        container,
        apps,
        autoPlay: true,
        interval: 6000,
      });
    } catch (err) {
      // Show friendly fallback — portfolio still looks good without API
      container.innerHTML = `
        <div style="text-align:center;padding:6rem 2rem;color:var(--color-text-muted)">
          <div style="font-size:48px;margin-bottom:1rem">📱</div>
          <h3 style="color:var(--color-heading);margin-bottom:0.5rem">Apps Showcase</h3>
          <p>Connect the backend to display live app data here.</p>
        </div>
      `;
    }
  }

  return { init };
})();
