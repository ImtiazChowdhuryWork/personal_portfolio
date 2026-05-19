/* ============================================================
 * FILE: components/ProjectDetail.js
 * WHAT IT IS:     Project details modal — opens when a visitor
 *                 clicks "View Details" on a slide in the Apps
 *                 Showcase. Shows the full project content
 *                 (long description, screenshot viewer with phone
 *                 mockup, all features, store + repo links).
 * WHERE USED:     AppShowcase.js calls ProjectDetail.show(app)
 *                 from the per-slide "View Details" button.
 * DEPENDS ON:     core/utils.js (parseArray)
 *                 components/modal.css + project-detail.css
 *                 #project-detail-modal element in index.html
 * LAST UPDATED:   2026-05-19 — mockup screenshot viewer
 * ============================================================ */

const ProjectDetail = (() => {

  let _escHandler  = null;
  let _keyHandler  = null;
  let _cycleTimer  = null;
  let _resumeTimer = null;
  let _currentIdx  = 0;
  let _totalShots  = 0;

  function show(project) {
    if (!project) return;
    const overlay = document.getElementById('project-detail-modal');
    if (!overlay) return;

    overlay.querySelector('.pd-body').innerHTML = render(project);
    overlay.classList.add('open');
    document.body.style.overflow = 'hidden';
    overlay.querySelector('.pd-scroll').scrollTop = 0;

    const shots = Utils.parseArray(project.screenshots);
    if (shots.length > 0) _initViewer(overlay, shots.length);

    _escHandler = (e) => {
      if (e.key === 'Escape' && !document.querySelector('.pd-lightbox')) close();
    };
    _keyHandler = (e) => {
      if (_totalShots < 2) return;
      if (e.key === 'ArrowLeft')  _navigate(overlay, -1);
      if (e.key === 'ArrowRight') _navigate(overlay,  1);
    };
    document.addEventListener('keydown', _escHandler);
    document.addEventListener('keydown', _keyHandler);
  }

  function close() {
    const overlay = document.getElementById('project-detail-modal');
    if (!overlay) return;
    overlay.classList.remove('open');
    document.body.style.overflow = '';
    if (_cycleTimer)  { clearInterval(_cycleTimer);  _cycleTimer  = null; }
    if (_resumeTimer) { clearTimeout(_resumeTimer);  _resumeTimer = null; }
    if (_escHandler)  { document.removeEventListener('keydown', _escHandler); _escHandler = null; }
    if (_keyHandler)  { document.removeEventListener('keydown', _keyHandler); _keyHandler = null; }
    _currentIdx = 0;
    _totalShots = 0;
  }

  // ─── Viewer init ────────────────────────────────────────────

  function _initViewer(overlay, total) {
    _currentIdx = 0;
    _totalShots = total;

    const viewer = overlay.querySelector('.pd-viewer');
    if (!viewer) return;

    // Tap phone → fullscreen (works for single screenshot too)
    viewer.querySelector('.pd-viewer-phone')?.addEventListener('click', () => {
      _openLightbox(overlay);
    });

    if (total < 2) return;

    // Thumbnail clicks
    viewer.querySelectorAll('.pd-thumb').forEach(thumb => {
      thumb.addEventListener('click', () => {
        _goTo(overlay, parseInt(thumb.dataset.idx));
        _pauseCycle(overlay, total);
      });
    });

    // Prev / Next buttons
    viewer.querySelector('.pd-nav-prev')?.addEventListener('click', () => {
      _navigate(overlay, -1);
      _pauseCycle(overlay, total);
    });
    viewer.querySelector('.pd-nav-next')?.addEventListener('click', () => {
      _navigate(overlay, 1);
      _pauseCycle(overlay, total);
    });

    // Touch swipe on the phone frame
    let touchStartX = 0;
    const frame = viewer.querySelector('.pd-viewer-frame');
    frame?.addEventListener('touchstart', e => {
      touchStartX = e.touches[0].clientX;
    }, { passive: true });
    frame?.addEventListener('touchend', e => {
      const delta = e.changedTouches[0].clientX - touchStartX;
      if (Math.abs(delta) > 50) {
        e.preventDefault();
        _navigate(overlay, delta < 0 ? 1 : -1);
        _pauseCycle(overlay, total);
      }
    });

    // Auto-cycle
    _startCycle(overlay, total);
  }

  function _goTo(overlay, idx) {
    _currentIdx = ((idx % _totalShots) + _totalShots) % _totalShots;

    overlay.querySelectorAll('.pd-viewer-img').forEach((img, i) => {
      img.classList.toggle('active', i === _currentIdx);
    });
    overlay.querySelectorAll('.pd-thumb').forEach((t, i) => {
      t.classList.toggle('active', i === _currentIdx);
    });

    const counter = overlay.querySelector('.pd-counter');
    if (counter) counter.textContent = `${_currentIdx + 1} / ${_totalShots}`;

    // Scroll active thumb into view inside the thumbs panel
    overlay.querySelectorAll('.pd-thumb')[_currentIdx]
      ?.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
  }

  function _navigate(overlay, dir) {
    _goTo(overlay, _currentIdx + dir);
  }

  function _startCycle(overlay, total) {
    if (_cycleTimer) clearInterval(_cycleTimer);
    _cycleTimer = setInterval(() => _navigate(overlay, 1), 3000);
  }

  function _pauseCycle(overlay, total) {
    if (_cycleTimer)  { clearInterval(_cycleTimer); _cycleTimer = null; }
    if (_resumeTimer) clearTimeout(_resumeTimer);
    _resumeTimer = setTimeout(() => _startCycle(overlay, total), 5000);
  }

  function _openLightbox(overlay) {
    const active = overlay.querySelector('.pd-viewer-img.active');
    if (!active) return;

    const lb = document.createElement('div');
    lb.className = 'pd-lightbox';
    lb.innerHTML = `<img src="${active.src}" alt="Screenshot fullscreen">`;

    function closeLightbox() {
      lb.classList.add('pd-lightbox--closing');
      setTimeout(() => lb.remove(), 270);
    }

    function escHandler(e) {
      if (e.key === 'Escape') {
        document.removeEventListener('keydown', escHandler);
        closeLightbox();
      }
    }

    lb.addEventListener('click', () => {
      document.removeEventListener('keydown', escHandler);
      closeLightbox();
    });
    document.addEventListener('keydown', escHandler);
    document.body.appendChild(lb);
  }

  // ─── HTML render ────────────────────────────────────────────

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

    const headerArt = p.thumbnail
      ? `<img class="pd-art" src="${p.thumbnail}" alt="${p.name}"
           onerror="this.outerHTML='<div class=&quot;pd-art pd-art-fallback&quot;>${(p.name || '?').slice(0,1).toUpperCase()}</div>'">`
      : `<div class="pd-art pd-art-fallback">${(p.name || '?').slice(0,1).toUpperCase()}</div>`;

    const featuredBadge = p.featured ? `<span class="pd-badge pd-badge-featured">⭐ Featured</span>` : '';
    const statusBadge   = statusLabel ? `<span class="pd-badge ${statusClass}">${statusLabel}</span>` : '';
    const shortDesc     = p.description ? `<p class="pd-tagline">${p.description}</p>` : '';

    // Screenshot viewer — phone mockup left, thumbnail grid right
    const screenshotsSection = shots.length
      ? `<section class="pd-section pd-viewer-section">
           <h4 class="pd-section-title"><i class="ph ph-images-square"></i> Screenshots</h4>
           <div class="pd-viewer">

             <div class="pd-viewer-left">
               <div class="pd-viewer-phone" title="Tap to fullscreen">
                 <div class="pd-viewer-frame">
                   <div class="pd-viewer-notch"></div>
                   <div class="pd-viewer-screen">
                     ${shots.map((url, i) => `
                       <img class="pd-viewer-img ${i === 0 ? 'active' : ''}"
                            src="${url}" alt="Screenshot ${i + 1}"
                            onerror="this.style.display='none'">
                     `).join('')}
                   </div>
                 </div>
               </div>
               ${shots.length > 1 ? `
               <div class="pd-viewer-controls">
                 <button class="pd-nav-btn pd-nav-prev" aria-label="Previous">
                   <i class="ph ph-caret-left"></i>
                 </button>
                 <span class="pd-counter">1 / ${shots.length}</span>
                 <button class="pd-nav-btn pd-nav-next" aria-label="Next">
                   <i class="ph ph-caret-right"></i>
                 </button>
               </div>` : ''}
             </div>

             ${shots.length > 1 ? `
             <div class="pd-viewer-thumbs">
               ${shots.map((url, i) => `
                 <div class="pd-thumb ${i === 0 ? 'active' : ''}" data-idx="${i}">
                   <img src="${url}" alt="Screenshot ${i + 1}"
                        onerror="this.parentNode.style.display='none'">
                 </div>
               `).join('')}
             </div>` : ''}

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
      p.app_store_url  ? `<a class="pd-link pd-link-store" href="${p.app_store_url}" target="_blank" rel="noopener"><i class="ph ph-apple-logo"></i><span><small>Download on</small><strong>App Store</strong></span></a>` : '',
      p.play_store_url ? `<a class="pd-link pd-link-store" href="${p.play_store_url}" target="_blank" rel="noopener"><i class="ph ph-google-play-logo"></i><span><small>Get it on</small><strong>Google Play</strong></span></a>` : '',
      p.github_url     ? `<a class="pd-link pd-link-github" href="${p.github_url}" target="_blank" rel="noopener"><i class="ph ph-github-logo"></i><span>View source on GitHub</span></a>` : '',
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
      ${screenshotsSection}
      ${about}
      ${techBlock}
      ${featuresBlock}
      ${linksBlock}
    `;
  }

  function escapeHtml(s) {
    return String(s == null ? '' : s).replace(/[&<>"']/g, m => (
      { '&':'&amp;', '<':'&lt;', '>':'&gt;', '"':'&quot;', "'":'&#39;' }[m]
    ));
  }

  return { show, close };
})();
