/* ============================================================
 * FILE: main.js
 * WHAT IT IS:     Portfolio page orchestrator — the entry point
 *                 for index.html's JavaScript
 * LAST UPDATED:   2026-05-07 — scroll animations now run after all
 *                 sections (including API-driven) are fully rendered
 * ============================================================ */

document.addEventListener('DOMContentLoaded', async () => {

  Sidebar.init();
  Navbar.init();

  const pageLoader = document.getElementById('page-loader');

  // Static sections — render immediately, no API needed
  HeroSection.init();
  ArchitectureSection.init();
  ServicesSection.init();
  GithubSection.init();
  ContactSection.init();

  // API-driven sections — wait for all to finish so dynamic content
  // (timeline items, tech stack bars, etc.) is in the DOM before
  // ScrollAnimations picks them up
  await Promise.allSettled([
    AboutSection.init(),
    TechStackSection.init(),
    AppsSection.init(),
    ExperienceSection.init(),
  ]);

  // Init scroll animations AFTER everything is rendered so the
  // IntersectionObserver observes all elements including dynamic ones
  ScrollAnimations.init();

  // Hide the page loader
  if (pageLoader) {
    pageLoader.classList.add('hidden');
    setTimeout(() => pageLoader.remove(), 600);
  }
});
