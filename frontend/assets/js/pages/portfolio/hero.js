/* ============================================================
 * FILE: pages/portfolio/hero.js
 * WHAT IT IS:     Hero section initialization
 * WHY IT EXISTS:  Handles the hero: animated counters, Vanta.js
 *                 3D background, and rotating text badge
 * WHERE USED:     Loaded by index.html, called from main.js
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const HeroSection = (() => {

  /**
   * FUNCTION: init
   * WHAT IT DOES: 1. Initializes the Vanta.js animated background
   *               2. Starts the stat counter animations
   *               3. Renders the rotating badge
   * WHERE CALLED: main.js → on DOMContentLoaded
   */
  function init() {
    initBackground();
    initCounters();
  }

  function initBackground() {
    const heroBg = document.getElementById('hero-bg');
    if (!heroBg) return;

    // Try Vanta.js NET effect — the animated lines/dots background
    // Falls back gracefully if Vanta or THREE.js aren't loaded
    if (typeof VANTA !== 'undefined' && typeof THREE !== 'undefined') {
      try {
        VANTA.NET({
          el: '#hero-bg',
          THREE,
          mouseControls: true,
          touchControls: true,
          gyroControls: false,
          minHeight: 200.0,
          minWidth: 200.0,
          scale: 1.0,
          scaleMobile: 1.0,
          color: 0x54c5f8,       // Flutter blue for the lines
          backgroundColor: 0x0a0a0f, // Match --color-bg
          points: 12,
          maxDistance: 20,
          spacing: 18,
        });
      } catch (e) {
        // Vanta failed — show a CSS fallback gradient instead
        applyGradientFallback(heroBg);
      }
    } else {
      applyGradientFallback(heroBg);
    }
  }

  function applyGradientFallback(el) {
    // CSS-only animated background when Vanta.js isn't available
    el.style.cssText = `
      background: radial-gradient(ellipse at 20% 50%, rgba(84,197,248,0.08) 0%, transparent 60%),
                  radial-gradient(ellipse at 80% 20%, rgba(224,64,251,0.06) 0%, transparent 50%),
                  var(--color-bg);
    `;
  }

  function initCounters() {
    // Find all counter elements — they have data-count and data-suffix attributes
    // Set in the HTML as: <span data-count="5" data-suffix="+">0+</span>
    const counters = document.querySelectorAll('[data-count]');
    if (!counters.length) return;

    // Use IntersectionObserver to only start counting when visible
    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          const el = entry.target;
          const target = parseFloat(el.dataset.count);
          const suffix = el.dataset.suffix || '';
          const isDecimal = el.dataset.decimal === 'true';

          if (isDecimal) {
            // For "2.5+" — count to integer then snap to decimal
            Utils.animateCounter(el, Math.floor(target), 1500, '');
            setTimeout(() => { el.textContent = target + suffix; }, 1500);
          } else {
            Utils.animateCounter(el, target, 1500, suffix);
          }
          observer.unobserve(el); // Count only once
        }
      });
    }, { threshold: 0.5 });

    counters.forEach(el => observer.observe(el));
  }

  return { init };
})();
