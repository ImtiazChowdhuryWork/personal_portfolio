/* ============================================================
 * FILE: main.js
 * WHAT IT IS:     Portfolio page orchestrator — the entry point
 *                 for index.html's JavaScript
 * WHY IT EXISTS:  Initializes all page sections in the correct order
 *                 after the DOM has fully loaded. This is the only
 *                 JS file that index.html calls directly — all
 *                 other JS is imported as modules below.
 * WHERE USED:     index.html → <script src="assets/js/main.js">
 * IF REMOVED:     None of the dynamic sections load (skills, apps, etc.)
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

document.addEventListener('DOMContentLoaded', async () => {

  // ─── Step 1: Initialize UI chrome ──────────────────────────
  // Sidebar active link tracking and mobile hamburger menu
  Sidebar.init();
  Navbar.init();

  // ─── Step 2: Show the page loader ──────────────────────────
  // The loader is shown until all critical data is loaded.
  // It's removed in Step 6 below.
  const pageLoader = document.getElementById('page-loader');

  // ─── Step 3: Initialize all portfolio sections ─────────────
  // Some sections (hero, architecture, services) are static and fast.
  // Others (tech stack, apps, experience) fetch from the API.
  // We run them in parallel where possible for speed.

  // Static sections — no API calls needed
  HeroSection.init();
  ArchitectureSection.init();
  ServicesSection.init();
  GithubSection.init();
  ContactSection.init();

  // API-driven sections — fetch data in parallel
  await Promise.allSettled([
    AboutSection.init(),       // Loads profile — updates hero, sidebar, about
    TechStackSection.init(),   // Loads skills — renders animated bars
    AppsSection.init(),        // Loads projects — renders slider + phone mockups
    ExperienceSection.init(),  // Loads experience — renders timeline
  ]);

  // ─── Step 4: Start scroll animations ───────────────────────
  // Simple AOS-like scroll observer for [data-aos] elements
  initScrollAnimations();

  // ─── Step 5: Hide the page loader ──────────────────────────
  if (pageLoader) {
    pageLoader.classList.add('hidden');
    // Remove from DOM after transition to free memory
    setTimeout(() => pageLoader.remove(), 600);
  }
});

/**
 * FUNCTION: initScrollAnimations
 * WHAT IT DOES: Adds the .aos-animate class to elements with [data-aos]
 *               attributes when they scroll into the viewport.
 *               This triggers the CSS transition defined in base.css.
 */
function initScrollAnimations() {
  const elements = document.querySelectorAll('[data-aos]');
  if (!elements.length) return;

  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        // Small delay offset based on data-aos-delay attribute
        const delay = parseInt(entry.target.dataset.aosDelay) || 0;
        setTimeout(() => {
          entry.target.classList.add('aos-animate');
        }, delay);
        observer.unobserve(entry.target); // Animate only once
      }
    });
  }, { threshold: 0.1 });

  elements.forEach(el => observer.observe(el));
}
