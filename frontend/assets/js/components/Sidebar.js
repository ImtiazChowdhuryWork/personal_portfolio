/* ============================================================
 * FILE: components/Sidebar.js
 * WHAT IT IS:     Sidebar component — highlights active nav link
 *                 as the user scrolls through portfolio sections
 * WHERE USED:     main.js → initialized once on page load
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const Sidebar = (() => {

  /**
   * FUNCTION: init
   * WHAT IT DOES:   Watches which section is currently in view and
   *                 adds the .active class to the matching sidebar nav link.
   *                 Uses IntersectionObserver for performance — no scroll
   *                 event listener needed.
   * WHERE CALLED:   main.js on DOMContentLoaded
   */
  function init() {
    const sections = document.querySelectorAll('section[id]');
    const navLinks = document.querySelectorAll('.right-menu a, .mobile-menu a');

    if (!sections.length || !navLinks.length) return;

    // IntersectionObserver fires when a section enters/leaves the viewport.
    // rootMargin: '-40% 0px -40% 0px' means the section is "active" when
    // it occupies the middle 20% of the viewport.
    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          const id = entry.target.id;
          navLinks.forEach(link => {
            link.classList.toggle('active', link.getAttribute('href') === `#${id}`);
          });
        }
      });
    }, { rootMargin: '-40% 0px -40% 0px' });

    sections.forEach(section => observer.observe(section));

    // Smooth scroll when clicking sidebar or mobile menu links
    navLinks.forEach(link => {
      link.addEventListener('click', (e) => {
        const href = link.getAttribute('href');
        if (href && href.startsWith('#')) {
          e.preventDefault();
          const target = document.querySelector(href);
          if (target) {
            target.scrollIntoView({ behavior: 'smooth', block: 'start' });
            // Close mobile menu if open
            document.getElementById('mobile-menu')?.classList.remove('open');
            document.getElementById('hamburger')?.classList.remove('open');
          }
        }
      });
    });
  }

  return { init };
})();
