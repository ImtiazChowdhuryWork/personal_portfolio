/* ============================================================
 * FILE: components/SkillBar.js
 * WHAT IT IS:     Animated skill bar renderer
 * WHY IT EXISTS:  The Tech Stack section needs skill bars that
 *                 animate from 0% to the actual percentage when
 *                 the user scrolls into view. This component does that.
 * WHERE USED:     techstack.js passes skill data to SkillBar.render()
 * HOW TO USE:
 *   SkillBar.render({
 *     container: document.querySelector('#skills-group-core'),
 *     skills: [{ name: 'Flutter', percentage: 95, icon: '...', description: '...' }]
 *   });
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const SkillBar = (() => {

  /**
   * FUNCTION: render
   * WHAT IT DOES: Renders skill bar HTML into the container,
   *               then observes when the container enters the viewport
   *               and triggers the fill animation.
   * @param {object} opts
   * @param {HTMLElement} opts.container - the DOM element to render into
   * @param {Array} opts.skills - array of skill objects from the API
   */
  function render({ container, skills }) {
    if (!container || !skills?.length) return;

    container.innerHTML = skills.map(skill => `
      <div class="skill-item">
        <div class="skill-icon">
          ${skill.icon
            ? `<img src="${skill.icon}" alt="${skill.name}" onerror="this.parentNode.innerHTML='${skill.name[0]}';">`
            : `<span style="font-size:11px;font-weight:700;color:var(--color-primary)">${skill.name.slice(0,2).toUpperCase()}</span>`
          }
        </div>
        <div class="skill-info">
          <div class="skill-name">${skill.name}</div>
          <div class="skill-bar-track">
            <div class="skill-bar-fill" data-target="${skill.percentage}" style="width:0%"></div>
          </div>
        </div>
        <span class="skill-percent">0%</span>
      </div>
    `).join('');

    // Observe when the container scrolls into view and animate bars
    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          animateBars(container);
          observer.unobserve(entry.target); // Only animate once
        }
      });
    }, { threshold: 0.1 });

    observer.observe(container);
  }

  function animateBars(container) {
    container.querySelectorAll('.skill-bar-fill').forEach(bar => {
      const target = parseInt(bar.dataset.target);
      const percentEl = bar.closest('.skill-item').querySelector('.skill-percent');

      // Set the CSS width — the transition in skill-bar.css does the animation
      bar.style.width = target + '%';

      // Also count up the number text
      let start = 0;
      const interval = setInterval(() => {
        start += 2;
        if (start >= target) {
          start = target;
          clearInterval(interval);
        }
        if (percentEl) percentEl.textContent = start + '%';
      }, 16);
    });
  }

  return { render };
})();
