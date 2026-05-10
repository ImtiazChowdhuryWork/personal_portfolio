/* ============================================================
 * FILE: pages/portfolio/experience.js
 * WHAT IT IS:     Experience Timeline section loader
 * WHERE USED:     main.js → called on DOMContentLoaded
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const ExperienceSection = (() => {

  async function init() {
    const container = document.getElementById('experience-timeline');
    if (!container) return;

    try {
      const res = await API.get('/experience');
      const experiences = res.data || [];

      if (!experiences.length) {
        renderFallback(container);
        return;
      }

      container.innerHTML = experiences.map(exp => buildTimelineItem(exp)).join('');
    } catch {
      renderFallback(container);
    }
  }

  // Per-section visibility helper. Backend stores show_* as *bool, so the
  // value on the wire can be true / false / null. null (and undefined for
  // legacy rows saved before these flags existed) is treated as visible —
  // matching the DB column default. Only an explicit `false` hides.
  const isVisible = (val) => val !== false;

  function buildTimelineItem(exp) {
    const period = exp.end_date ? `${exp.start_date} — ${exp.end_date}` : exp.start_date;

    const logo = (isVisible(exp.show_logo) && exp.company_logo)
      ? `<img class="timeline-logo" src="${exp.company_logo}" alt="${exp.company} logo"
           onerror="this.style.display='none'">`
      : '';

    const locationSuffix = (isVisible(exp.show_location) && exp.location)
      ? ` · ${exp.location}` : '';

    const typeBadge = isVisible(exp.show_type)
      ? `<span class="timeline-type ${exp.type || 'full-time'}">${exp.type || 'Full-time'}</span>`
      : '';

    const description = (isVisible(exp.show_description) && exp.description)
      ? `<p style="font-size:var(--fs-sm);color:var(--color-text-muted);line-height:1.7;margin-bottom:1rem">${exp.description}</p>`
      : '';

    const achievements = isVisible(exp.show_achievements)
      ? Utils.parseArray(exp.achievements).map(a => `<li>${a}</li>`).join('')
      : '';

    const techTags = isVisible(exp.show_tech_used)
      ? Utils.parseArray(exp.tech_used).map(t => `<span class="timeline-tech-tag">${t}</span>`).join('')
      : '';

    return `
      <div class="timeline-item scroll-animation" data-animation="fade_from_bottom">
        <div class="timeline-header">
          ${logo}
          <div class="timeline-company-info">
            <h3 class="timeline-role">${exp.role}</h3>
            <div class="timeline-company">${exp.company}${locationSuffix}</div>
          </div>
          <div class="timeline-meta">
            <span class="timeline-period">${period}</span>
            ${exp.is_current ? `<span class="timeline-current">Present</span>` : ''}
            ${typeBadge}
          </div>
        </div>
        ${description}
        ${achievements ? `<ul class="timeline-achievements">${achievements}</ul>` : ''}
        ${techTags ? `<div class="timeline-tech">${techTags}</div>` : ''}
      </div>
    `;
  }

  function renderFallback(container) {
    container.innerHTML = `<p style="color:var(--color-text-muted)">Experience data will appear here once the backend is connected.</p>`;
  }

  return { init };
})();
