/* ============================================================
 * FILE: pages/portfolio/hero.js
 * WHAT IT IS:     Hero section — stat counter animations only.
 *                 Background is plain dark like Drake (no animation).
 * LAST UPDATED:   2026-05-07 — removed Vanta, plain background
 * ============================================================ */

const HeroSection = (() => {

  function init() {
    initCounters();
  }

  function initCounters() {
    const counters = document.querySelectorAll('[data-count]');
    if (!counters.length) return;

    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (!entry.isIntersecting) return;
        const el = entry.target;
        const target = parseFloat(el.dataset.count);
        const suffix = el.dataset.suffix || '';
        const isDecimal = el.dataset.decimal === 'true';

        if (isDecimal) {
          Utils.animateCounter(el, Math.floor(target), 1500, '');
          setTimeout(() => { el.textContent = target + suffix; }, 1500);
        } else {
          Utils.animateCounter(el, target, 1500, suffix);
        }
        observer.unobserve(el);
      });
    }, { threshold: 0.5 });

    counters.forEach(el => observer.observe(el));
  }

  return { init };
})();
