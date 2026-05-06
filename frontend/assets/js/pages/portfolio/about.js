/* ============================================================
 * FILE: pages/portfolio/about.js
 * WHAT IT IS:     About section — loads profile data from API
 * WHERE USED:     main.js → called on DOMContentLoaded
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const AboutSection = (() => {

  async function init() {
    try {
      const res = await API.get('/profile');
      const profile = res.data;
      if (!profile) return;

      Store.set('profile', profile);
      updateAboutSection(profile);
      updateHeroSection(profile);
      updateSidebar(profile);
    } catch {
      // Profile API failed — static HTML fallback already in index.html
    }
  }

  function updateAboutSection(p) {
    const bioEl = document.getElementById('about-bio');
    if (bioEl && p.bio) bioEl.textContent = p.bio;

    const availEl = document.getElementById('about-availability');
    if (availEl && p.availability) availEl.textContent = p.availability;

    const photoEl = document.getElementById('about-photo');
    if (photoEl && p.profile_photo) {
      photoEl.src = p.profile_photo;
      photoEl.alt = p.full_name;
    }
  }

  function updateHeroSection(p) {
    const shortBio = document.getElementById('hero-short-bio');
    if (shortBio && p.short_bio) shortBio.textContent = p.short_bio;
  }

  function updateSidebar(p) {
    const nameEl = document.getElementById('sidebar-name');
    if (nameEl && p.full_name) nameEl.textContent = p.full_name;

    const emailEl = document.getElementById('sidebar-email');
    if (emailEl && p.email) emailEl.textContent = p.email;

    const locEl = document.getElementById('sidebar-location');
    if (locEl && p.location) locEl.textContent = p.location;
  }

  return { init };
})();
