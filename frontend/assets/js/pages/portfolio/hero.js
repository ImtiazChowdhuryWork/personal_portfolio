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

  // Runs the count-up animation for a single stat element. Reads its data-*
  // attrs so the caller can tweak them (e.g. when profile data arrives) and
  // re-trigger to land on the new target without flicker.
  function runStatAnimation(el) {
    if (!el) return;
    const target = parseFloat(el.dataset.count);
    if (Number.isNaN(target)) return;
    const suffix = el.dataset.suffix || '';
    const decimals = el.dataset.decimal === 'true' ? 1 : 0;
    Utils.animateCounter(el, target, 1800, suffix, decimals);
    el.dataset.animated = 'true';
  }

  function initCounters() {
    const counters = document.querySelectorAll('[data-count]');
    if (!counters.length) return;

    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (!entry.isIntersecting) return;
        runStatAnimation(entry.target);
        observer.unobserve(entry.target);
      });
    }, { threshold: 0.5 });

    counters.forEach(el => observer.observe(el));
  }

  return { init, runStatAnimation };
})();
