const ScrollAnimations = (() => {

  let observer = null;

  function init() {
    if (observer) observer.disconnect();

    const elements = document.querySelectorAll('.scroll-animation');
    if (!elements.length) return;

    observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (!entry.isIntersecting) return;

        const el = entry.target;
        el.classList.add('is-visible');

        // Unobserve immediately — animation plays once, never re-triggers.
        // This eliminates the toggle-on/off blinking caused by elements
        // sitting at the viewport boundary during slow scrolling.
        observer.unobserve(el);

        // Release the GPU layer hint after the transition completes
        // so we don't hold compositor layers open unnecessarily.
        el.addEventListener('transitionend', () => {
          el.style.willChange = 'auto';
        }, { once: true });
      });
    }, {
      // 10% of the element must be visible before triggering —
      // prevents firing on a single pixel at the viewport edge.
      threshold: 0.1,
      // Extra 8% bottom margin gives a comfortable buffer so the
      // trigger point is well away from the boundary.
      rootMargin: '0px 0px -8% 0px',
    });

    // Only observe elements that haven't animated yet —
    // avoids re-observing already-visible elements on re-init
    // (which happens when dynamic API content is injected).
    elements.forEach(el => {
      if (!el.classList.contains('is-visible')) {
        observer.observe(el);
      }
    });
  }

  return { init };
})();
