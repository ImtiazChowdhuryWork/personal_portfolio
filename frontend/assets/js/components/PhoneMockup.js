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
   * @param {HTMLElement} opts.container - where to render the phone
   * @param {string[]} opts.screenshots  - array of image URLs
   * @param {number} opts.interval       - ms between screenshot changes
   */
  function render({ container, screenshots = [], interval = 3000 }) {
    if (!container) return;

    const hasScreenshots = screenshots.length > 0;
    const screenContent = hasScreenshots
      ? screenshots.map((url, i) => `
          <img class="phone-screenshot ${i === 0 ? 'active' : ''}"
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

    const dotsHTML = hasScreenshots && screenshots.length > 1
      ? `<div class="phone-dots">
           ${screenshots.map((_, i) => `
             <div class="phone-dot ${i === 0 ? 'active' : ''}" data-idx="${i}"></div>
           `).join('')}
         </div>`
      : '';

    container.innerHTML = `
      <div class="phone-mockup">
        <div class="phone-frame">
          <div class="phone-notch"></div>
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
      startCycling(container, screenshots.length, interval);
    }
  }

  function startCycling(container, total, interval) {
    let current = 0;
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
