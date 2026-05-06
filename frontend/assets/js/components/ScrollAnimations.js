/* ============================================================
 * FILE: components/ScrollAnimations.js
 * WHAT IT IS:     Scroll-triggered entrance animations using
 *                 IntersectionObserver + CSS transitions.
 *                 Same visual effect as Drake's GSAP animations —
 *                 fade up from below, 1.2s, power4.out easing.
 *                 No CDN dependency — always works.
 * HOW IT WORKS:
 *   1. Every .scroll-animation element starts hidden (opacity:0,
 *      transformed) via CSS in base.css.
 *   2. IntersectionObserver watches each element.
 *   3. When the element enters the viewport, JS adds .is-visible
 *      which triggers the CSS transition to natural position.
 *   4. When the element leaves, .is-visible is removed so it
 *      resets — same as Drake's "reverse on scroll up" behaviour.
 * LAST UPDATED:   2026-05-07 — replaced GSAP with IntersectionObserver
 * ============================================================ */

const ScrollAnimations = (() => {

  let observer = null;

  /**
   * FUNCTION: init
   * WHAT IT DOES: Finds all .scroll-animation elements and sets up
   *               an IntersectionObserver that adds/removes .is-visible
   *               as each element enters or leaves the viewport.
   *               Safe to call multiple times — disconnects any
   *               previous observer first so dynamic content
   *               (timeline items rendered after API calls) gets
   *               picked up on re-call.
   * WHERE CALLED: main.js after sections render, and from section
   *               scripts after dynamic content is inserted.
   */
  function init() {
    // Disconnect any previous observer before creating a new one
    // so we don't double-observe elements
    if (observer) observer.disconnect();

    const elements = document.querySelectorAll('.scroll-animation');
    if (!elements.length) return;

    // rootMargin: "0px 0px -5% 0px" means the trigger fires when
    // the element is 5% above the bottom edge of the viewport —
    // equivalent to Drake's "top bottom+=20%" ScrollTrigger start.
    observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          // Element entered viewport — add class to animate in
          entry.target.classList.add('is-visible');
        } else {
          // Element left viewport — remove class to reset (animate out)
          // This mirrors Drake's "toggleActions: play none none reverse"
          entry.target.classList.remove('is-visible');
        }
      });
    }, {
      rootMargin: '0px 0px -5% 0px',
      threshold: 0,
    });

    elements.forEach(el => observer.observe(el));
  }

  return { init };
})();
