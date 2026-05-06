/* ============================================================
 * FILE: components/Navbar.js
 * WHAT IT IS:     Mobile navigation bar component
 * WHERE USED:     main.js → initialized once on page load
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const Navbar = (() => {

  /**
   * FUNCTION: init
   * WHAT IT DOES: Wires up the hamburger button to open/close
   *               the mobile fullscreen menu with animation.
   *               Also closes the menu when a link is clicked
   *               (smooth scroll to section).
   * WHERE CALLED: main.js on DOMContentLoaded
   */
  function init() {
    const hamburger = document.getElementById('hamburger');
    const mobileMenu = document.getElementById('mobile-menu');

    if (!hamburger || !mobileMenu) return;

    hamburger.addEventListener('click', () => {
      hamburger.classList.toggle('open');
      mobileMenu.classList.toggle('open');
      // Prevent body scroll when menu is open
      document.body.style.overflow = mobileMenu.classList.contains('open') ? 'hidden' : '';
    });

    // Close the menu when any link is clicked
    mobileMenu.querySelectorAll('a').forEach(link => {
      link.addEventListener('click', () => {
        hamburger.classList.remove('open');
        mobileMenu.classList.remove('open');
        document.body.style.overflow = '';
      });
    });

    // Close menu when clicking outside
    document.addEventListener('click', (e) => {
      if (!hamburger.contains(e.target) && !mobileMenu.contains(e.target)) {
        hamburger.classList.remove('open');
        mobileMenu.classList.remove('open');
        document.body.style.overflow = '';
      }
    });
  }

  return { init };
})();
