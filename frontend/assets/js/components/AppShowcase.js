/* ============================================================
 * FILE: components/AppShowcase.js
 * WHAT IT IS:     Full-width app showcase slider component
 * WHY IT EXISTS:  The centerpiece of the portfolio — shows each app
 *                 in a full-screen slide with phone mockup + content.
 *                 Uses Swiper.js for touch/swipe support.
 * WHERE USED:     apps.js calls AppShowcase.render() with API data
 * HOW TO USE:
 *   AppShowcase.render({
 *     container: document.querySelector('#apps-showcase'),
 *     apps: projectsArray,
 *     autoPlay: true,
 *     interval: 5000
 *   });
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const AppShowcase = (() => {

  let swiperInstance = null;

  /**
   * FUNCTION: render
   * WHAT IT DOES: Renders the full apps showcase including:
   *               - Swiper slider with one slide per app
   *               - Phone mockup with cycling screenshots on each slide
   *               - Thumbnail strip at the bottom for quick navigation
   *               - App Store / Play Store badges
   * @param {object} opts
   * @param {HTMLElement} opts.container - the section element
   * @param {Array} opts.apps - project objects from the API
   * @param {boolean} opts.autoPlay - auto-advance slides
   * @param {number} opts.interval - ms between auto-advance
   */
  function render({ container, apps, autoPlay = true, interval = 6000 }) {
    if (!container || !apps?.length) {
      container.innerHTML = `<div style="text-align:center;padding:4rem;color:var(--color-text-muted)">No apps to display yet.</div>`;
      return;
    }

    const slidesHTML = apps.map(app => buildSlide(app)).join('');
    const thumbnailsHTML = apps.map((app, i) => `
      <div class="app-thumbnail ${i === 0 ? 'active' : ''}" data-slide="${i}">
        <div class="app-thumbnail-img" style="background:var(--color-bg-4);display:flex;align-items:center;justify-content:center;font-size:20px;">
          ${app.thumbnail
            ? `<img src="${app.thumbnail}" alt="${app.name}" style="width:100%;height:100%;object-fit:cover;border-radius:8px;">`
            : '📱'
          }
        </div>
        <span class="app-thumbnail-name">${app.name}</span>
      </div>
    `).join('');

    container.innerHTML = `
      <div class="swiper apps-swiper">
        <div class="swiper-wrapper">
          ${slidesHTML}
        </div>
        <div class="swiper-button-next"></div>
        <div class="swiper-button-prev"></div>
      </div>
      <div class="app-thumbnails">${thumbnailsHTML}</div>
    `;

    // Initialize Swiper after DOM is rendered
    initSwiper(container, apps.length, autoPlay, interval);

    // Wire up thumbnail clicks
    container.querySelectorAll('.app-thumbnail').forEach(thumb => {
      thumb.addEventListener('click', () => {
        const idx = parseInt(thumb.dataset.slide);
        if (swiperInstance) swiperInstance.slideTo(idx);
      });
    });

    // Render phone mockups inside each slide
    container.querySelectorAll('.app-slide').forEach((slide, i) => {
      const phoneSlot = slide.querySelector('.phone-slot');
      if (phoneSlot) {
        PhoneMockup.render({
          container: phoneSlot,
          screenshots: Utils.parseArray(apps[i]?.screenshots),
          interval: 3000,
        });
      }
    });
  }

  function buildSlide(app) {
    const techTags = Utils.parseArray(app.tech_stack)
      .map(t => `<span class="app-tech-tag">${t}</span>`).join('');
    const features = Utils.parseArray(app.features)
      .map(f => `<li>${f}</li>`).join('');

    const platformBadges = [
      app.app_store_url ? `<span class="platform-badge ios">📱 iOS</span>` : '',
      app.play_store_url ? `<span class="platform-badge android">🤖 Android</span>` : '',
    ].filter(Boolean).join('');

    const storeLinks = [
      app.app_store_url ? `
        <a href="${app.app_store_url}" target="_blank" rel="noopener" class="store-badge">
          <span class="store-badge-icon">🍎</span>
          <span class="store-badge-text">
            <span class="store-badge-sub">Download on the</span>
            <span class="store-badge-name">App Store</span>
          </span>
        </a>` : '',
      app.play_store_url ? `
        <a href="${app.play_store_url}" target="_blank" rel="noopener" class="store-badge">
          <span class="store-badge-icon">▶</span>
          <span class="store-badge-text">
            <span class="store-badge-sub">Get it on</span>
            <span class="store-badge-name">Google Play</span>
          </span>
        </a>` : '',
    ].filter(Boolean).join('');

    return `
      <div class="swiper-slide">
        <div class="app-slide">
          <div class="app-slide-content">
            <div class="app-platform-badges">${platformBadges}</div>
            <h2 class="app-slide-title">${app.name}</h2>
            <p class="app-slide-desc">${app.description || ''}</p>
            <div class="app-tech-tags">${techTags}</div>
            <ul class="app-features">${features}</ul>
            <div class="app-store-links">${storeLinks}</div>
          </div>
          <div class="app-slide-phone">
            <div class="phone-slot"></div>
          </div>
        </div>
      </div>
    `;
  }

  function initSwiper(container, totalSlides, autoPlay, interval) {
    if (swiperInstance) swiperInstance.destroy(true, true);

    if (typeof Swiper === 'undefined') return;

    swiperInstance = new Swiper(container.querySelector('.apps-swiper'), {
      loop: totalSlides > 1,
      effect: 'slide',
      speed: 600,
      autoplay: autoPlay ? { delay: interval, disableOnInteraction: false } : false,
      navigation: {
        nextEl: container.querySelector('.swiper-button-next'),
        prevEl: container.querySelector('.swiper-button-prev'),
      },
      on: {
        slideChange() {
          const idx = this.realIndex;
          container.querySelectorAll('.app-thumbnail').forEach((t, i) => {
            t.classList.toggle('active', i === idx);
          });
        },
      },
    });
  }

  return { render };
})();
