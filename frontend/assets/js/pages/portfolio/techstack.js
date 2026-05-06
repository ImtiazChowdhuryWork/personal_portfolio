/* ============================================================
 * FILE: pages/portfolio/techstack.js
 * WHAT IT IS:     Tech Stack section loader
 * WHERE USED:     main.js → called on DOMContentLoaded
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const TechStackSection = (() => {

  // Maps category keys from the API to display labels
  const categoryLabels = {
    core:             'Core',
    state_management: 'State Management',
    backend:          'Backend & APIs',
    payments:         'Payments',
    tools:            'Tools & DevOps',
  };

  // Display order for categories
  const categoryOrder = ['core', 'state_management', 'backend', 'payments', 'tools'];

  /**
   * FUNCTION: init
   * WHAT IT DOES: Fetches skills from the API, groups them by category,
   *               and renders each group with animated skill bars.
   * WHERE CALLED: main.js on DOMContentLoaded
   */
  async function init() {
    const container = document.getElementById('tech-stack-content');
    if (!container) return;

    // Show loading state
    container.innerHTML = `<div style="color:var(--color-text-muted);padding:2rem">Loading tech stack...</div>`;

    try {
      // Fetch skills from the backend API
      const res = await API.get('/skills');
      const skills = res.data || [];

      if (!skills.length) {
        container.innerHTML = renderFallback();
        initBarsAfterRender(container);
        return;
      }

      // Group skills by category
      const groups = {};
      skills.forEach(skill => {
        if (!groups[skill.category]) groups[skill.category] = [];
        groups[skill.category].push(skill);
      });

      // Render groups in the defined order
      const groupsHTML = categoryOrder
        .filter(cat => groups[cat])
        .map(cat => `
          <div class="tech-group">
            <h4 class="tech-group-title">${categoryLabels[cat] || cat}</h4>
            <div class="skill-group-items" data-category="${cat}"></div>
          </div>
        `).join('');

      container.innerHTML = `<div class="techstack-layout"><div class="techstack-col-1"></div><div class="techstack-col-2"></div></div>`;
      const col1 = container.querySelector('.techstack-col-1');
      const col2 = container.querySelector('.techstack-col-2');

      // Split categories between two columns for two-column layout
      const orderedCats = categoryOrder.filter(cat => groups[cat]);
      const half = Math.ceil(orderedCats.length / 2);

      orderedCats.slice(0, half).forEach(cat => {
        const groupEl = createGroupEl(cat, groups[cat]);
        col1.appendChild(groupEl);
      });

      orderedCats.slice(half).forEach(cat => {
        const groupEl = createGroupEl(cat, groups[cat]);
        col2.appendChild(groupEl);
      });

      // Now render animated bars into each group
      orderedCats.forEach(cat => {
        const groupContainer = container.querySelector(`[data-skill-cat="${cat}"]`);
        if (groupContainer) {
          SkillBar.render({ container: groupContainer, skills: groups[cat] });
        }
      });

    } catch (err) {
      // API failed — show hardcoded fallback data so the section isn't empty
      container.innerHTML = renderFallback();
      initBarsAfterRender(container);
    }
  }

  function createGroupEl(cat, skills) {
    const div = document.createElement('div');
    div.className = 'tech-group';
    div.innerHTML = `
      <h4 class="tech-group-title">${categoryLabels[cat] || cat}</h4>
      <div data-skill-cat="${cat}"></div>
    `;
    // Render skill bars immediately
    const barContainer = div.querySelector(`[data-skill-cat="${cat}"]`);
    SkillBar.render({ container: barContainer, skills });
    return div;
  }

  // Fallback static data when API isn't running
  function renderFallback() {
    return `<p style="color:var(--color-text-muted)">Tech stack data will appear here once the backend is connected.</p>`;
  }

  function initBarsAfterRender(container) {
    // No-op for static fallback
  }

  return { init };
})();
