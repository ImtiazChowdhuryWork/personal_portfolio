/* ============================================================
 * FILE: components/PhoneMockup.js
 * WHAT IT IS:     Phone frame component with cycling screenshots
 * WHY IT EXISTS:  The Apps Showcase shows real app screenshots inside
 *                 a phone frame. This component renders the frame and
 *                 cycles through screenshots automatically.
 * WHERE USED:     AppShowcase.js calls PhoneMockup.render() per slide
 * HOW TO USE:
 *   PhoneMockup.render({
 *     container: document.querySelector('.phone-slot'),
 *     screenshots: ['url1.jpg', 'url2.jpg'],
 *     interval: 3000
 *   });
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const PhoneMockup = (() => {

  /**
   * FUNCTION: render
   * WHAT IT DOES: Renders a phone frame with screenshots cycling inside.
   *               If no screenshots exist, shows a placeholder gradient.
   * @param {object} opts
   * @param {HTMLElement} opts.container  - where to render the phone
   * @param {string[]} opts.screenshots   - array of image URLs
   * @param {number}   opts.interval      - ms between screenshot changes
   * @param {string}   opts.variant       - 'ios' (default, notch + dynamic
   *                                         island) or 'android' (centered
   *                                         punch-hole camera, smaller radius)
   * @param {number}   opts.startIndex    - which screenshot to show first.
   *                                         Used by AppShowcase to cycle the
   *                                         iOS + Android phones out of sync.
   * @param {boolean}  opts.showDots      - render the dot indicator strip
   *                                         underneath. Default true; the
   *                                         showcase pair turns this off on
   *                                         the second phone so we don't
   *                                         show two duplicate strips.
   */
  function render({
    container,
    screenshots = [],
    interval = 3000,
    variant = 'ios',
    startIndex = 0,
    showDots = true,
  }) {
    if (!container) return;

    const hasScreenshots = screenshots.length > 0;
    // Clamp startIndex into valid range so callers can't crash the render
    // by passing a number larger than the array length.
    const startIdx = hasScreenshots
      ? ((startIndex % screenshots.length) + screenshots.length) % screenshots.length
      : 0;

    const screenContent = hasScreenshots
      ? screenshots.map((url, i) => `
          <img class="phone-screenshot ${i === startIdx ? 'active' : ''}"
               src="${url}"
               alt="App screenshot ${i + 1}"
               onerror="this.style.display='none';">
        `).join('')
      : `<div class="phone-screen-placeholder">
           <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
             <rect x="5" y="2" width="14" height="20" rx="2"/>
             <circle cx="12" cy="17" r="1"/>
           </svg>
           <span>Screenshots coming soon</span>
         </div>`;

    const dotsHTML = (showDots && hasScreenshots && screenshots.length > 1)
      ? `<div class="phone-dots">
           ${screenshots.map((_, i) => `
             <div class="phone-dot ${i === startIdx ? 'active' : ''}" data-idx="${i}"></div>
           `).join('')}
         </div>`
      : '';

    // Variant-specific top chrome: iOS gets the pill notch + dynamic island
    // dot, Android gets a small centered punch-hole camera.
    const topChrome = variant === 'android'
      ? `<div class="phone-punch-hole"></div>`
      : `<div class="phone-notch"></div>`;

    container.innerHTML = `
      <div class="phone-mockup phone-mockup--${variant}">
        <div class="phone-frame">
          ${topChrome}
          <div class="phone-screen">${screenContent}</div>
          <div class="phone-btn-right"></div>
          <div class="phone-btn-left-1"></div>
          <div class="phone-btn-left-2"></div>
        </div>
      </div>
      ${dotsHTML}
    `;

    // Start the screenshot cycling if multiple screenshots exist
    if (hasScreenshots && screenshots.length > 1) {
      startCycling(container, screenshots.length, interval, startIdx);
    }
  }

  function startCycling(container, total, interval, startIdx = 0) {
    let current = startIdx;
    setInterval(() => {
      current = (current + 1) % total;
      updateActive(container, current);
    }, interval);

    // Also handle dot clicks
    container.querySelectorAll('.phone-dot').forEach(dot => {
      dot.addEventListener('click', () => {
        current = parseInt(dot.dataset.idx);
        updateActive(container, current);
      });
    });
  }

  function updateActive(container, idx) {
    container.querySelectorAll('.phone-screenshot').forEach((img, i) => {
      img.classList.toggle('active', i === idx);
    });
    container.querySelectorAll('.phone-dot').forEach((dot, i) => {
      dot.classList.toggle('active', i === idx);
    });
  }

  return { render };
})();
