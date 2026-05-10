/* ============================================================
 * FILE: components/ProjectDetail.js
 * WHAT IT IS:     Project details modal — opens when a visitor
 *                 clicks "View Details" on a slide in the Apps
 *                 Showcase. Shows the full project content
 *                 (long description, all screenshots, every
 *                 feature, store + repo links) without leaving
 *                 the single-page portfolio.
 * WHERE USED:     AppShowcase.js calls ProjectDetail.show(app)
 *                 from the per-slide "View Details" button.
 * DEPENDS ON:     core/utils.js (parseArray)
 *                 components/modal.css + project-detail.css
 *                 #project-detail-modal element in index.html
 * IF REMOVED:     "View Details" button on each slide does
 *                 nothing — long_description, full screenshot
 *                 gallery, features list, etc. become unreachable
 *                 from the public portfolio.
 * LAST UPDATED:   2026-05-10 — initial creation
 * ============================================================ */

const ProjectDetail = (() => {

  let _escHandler = null;

  /**
   * FUNCTION: show
   * WHAT IT DOES:   Renders the modal body for a given project
   *                 and opens the overlay. Locks body scroll while
   *                 the modal is open. Esc + backdrop click close.
   * @param {object} project — the API project object
   */
  function show(project) {
    if (!project) return;
    const overlay = document.getElementById('project-detail-modal');
    if (!overlay) return;
    overlay.querySelector('.pd-body').innerHTML = render(project);
    overlay.classList.add('open');
    document.body.style.overflow = 'hidden';
    // Reset scroll on every open so visitors land at the top
    overlay.querySelector('.pd-scroll').scrollTop = 0;
    // Esc to close — bind once per open, removed on close
    _escHandler = (e) => { if (e.key === 'Escape') close(); };
    document.addEventListener('keydown', _escHandler);
  }

  function close() {
    const overlay = document.getElementById('project-detail-modal');
    if (!overlay) return;
    overlay.classList.remove('open');
    document.body.style.overflow = '';
    if (_escHandler) {
      document.removeEventListener('keydown', _escHandler);
      _escHandler = null;
    }
  }

  function render(p) {
    const tech     = Utils.parseArray(p.tech_stack);
    const features = Utils.parseArray(p.features);
    const shots    = Utils.parseArray(p.screenshots);

    const statusClass = p.status === 'live'        ? 'is-live'
                      : p.status === 'development' ? 'is-dev'
                      : p.status === 'archived'    ? 'is-archived' : '';
    const statusLabel = p.status === 'live'        ? '🟢 Live'
                      : p.status === 'development' ? '🟡 In Development'
                      : p.status === 'archived'    ? '⚪ Archived' : (p.status || '');

    // Header: thumbnail (or initial fallback) + name + meta badges
    const headerArt = p.thumbnail
      ? `<img class="pd-art" src="${p.thumbnail}" alt="${p.name}"
           onerror="this.outerHTML='<div class=&quot;pd-art pd-art-fallback&quot;>${(p.name || '?').slice(0,1).toUpperCase()}</div>'">`
      : `<div class="pd-art pd-art-fallback">${(p.name || '?').slice(0,1).toUpperCase()}</div>`;

    const featuredBadge = p.featured ? `<span class="pd-badge pd-badge-featured">⭐ Featured</span>` : '';
    const statusBadge   = statusLabel ? `<span class="pd-badge ${statusClass}">${statusLabel}</span>` : '';

    // Sections — each one only renders if there's content to show, so
    // sparse projects don't get empty headings staring back at visitors.
    const shortDesc = p.description
      ? `<p class="pd-tagline">${p.description}</p>`
      : '';

    const gallery = shots.length
      ? `<section class="pd-section">
           <h4 class="pd-section-title"><i class="ph ph-images-square"></i> Screenshots</h4>
           <div class="pd-gallery">
             ${shots.map((url, i) => `
               <div class="pd-shot">
                 <img src="${url}" alt="Screenshot ${i + 1}"
                   onerror="this.parentNode.outerHTML=''">
               </div>
             `).join('')}
           </div>
         </section>`
      : '';

    const about = p.long_description
      ? `<section class="pd-section">
           <h4 class="pd-section-title"><i class="ph ph-text-align-left"></i> About</h4>
           <p class="pd-prose">${escapeHtml(p.long_description).replace(/\n/g, '<br>')}</p>
         </section>`
      : '';

    const techBlock = tech.length
      ? `<section class="pd-section">
           <h4 class="pd-section-title"><i class="ph ph-stack"></i> Tech Stack</h4>
           <div class="pd-tech">
             ${tech.map(t => `<span class="pd-tech-chip">${escapeHtml(t)}</span>`).join('')}
           </div>
         </section>`
      : '';

    const featuresBlock = features.length
      ? `<section class="pd-section">
           <h4 class="pd-section-title"><i class="ph ph-list-checks"></i> Key Features</h4>
           <ul class="pd-features">
             ${features.map(f => `<li>${escapeHtml(f)}</li>`).join('')}
           </ul>
         </section>`
      : '';

    const links = [
      p.app_store_url ? `<a class="pd-link pd-link-store" href="${p.app_store_url}" target="_blank" rel="noopener"><i class="ph ph-apple-logo"></i><span><small>Download on</small><strong>App Store</strong></span></a>` : '',
      p.play_store_url ? `<a class="pd-link pd-link-store" href="${p.play_store_url}" target="_blank" rel="noopener"><i class="ph ph-google-play-logo"></i><span><small>Get it on</small><strong>Google Play</strong></span></a>` : '',
      p.github_url ? `<a class="pd-link pd-link-github" href="${p.github_url}" target="_blank" rel="noopener"><i class="ph ph-github-logo"></i><span>View source on GitHub</span></a>` : '',
    ].filter(Boolean).join('');

    const linksBlock = links
      ? `<section class="pd-section">
           <h4 class="pd-section-title"><i class="ph ph-link"></i> Get It</h4>
           <div class="pd-links">${links}</div>
         </section>`
      : '';

    return `
      <div class="pd-header">
        ${headerArt}
        <div class="pd-header-info">
          <h2 class="pd-title">${escapeHtml(p.name || 'Untitled project')}</h2>
          <div class="pd-meta">
            ${statusBadge}
            ${featuredBadge}
          </div>
          ${shortDesc}
        </div>
      </div>
      ${gallery}
      ${about}
      ${techBlock}
      ${featuresBlock}
      ${linksBlock}
    `;
  }

  // Defensive HTML escape — we render user-authored content from the DB,
  // so anything that survived the dashboard's plain-text inputs still
  // shouldn't run as markup. Only the tagline + features get this; long
  // description goes through escape + newline-to-<br>.
  function escapeHtml(s) {
    return String(s == null ? '' : s).replace(/[&<>"']/g, m => (
      { '&':'&amp;', '<':'&lt;', '>':'&gt;', '"':'&quot;', "'":'&#39;' }[m]
    ));
  }

  return { show, close };
})();
