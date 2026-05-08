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
      updateSocialLinks(profile);
    } catch {
      // Profile API failed — static HTML fallback already in index.html
    }
  }

  function updateAboutSection(p) {
    const bioEl = document.getElementById('about-bio');
    if (bioEl && p.bio) bioEl.textContent = p.bio;

    const availEl = document.getElementById('about-availability');
    if (availEl && p.availability) availEl.textContent = p.availability;

    // About section uses its own dedicated photo
    const photoEl = document.getElementById('about-photo');
    if (photoEl && p.about_photo) {
      photoEl.src = p.about_photo;
      photoEl.alt = p.full_name;
      photoEl.style.display = 'block';
      // Hide the fallback div that onerror may have shown
      const fallback = photoEl.nextElementSibling;
      if (fallback) fallback.style.display = 'none';
    }

    // Show CV buttons only if a CV has been uploaded, hide otherwise
    ['cv-download-btn-1', 'cv-download-btn-2'].forEach(id => {
      const btn = document.getElementById(id);
      if (!btn) return;
      if (p.cv_file) {
        btn.href = p.cv_file;
        btn.setAttribute('download', p.cv_file.split('/').pop());
        btn.style.display = 'inline-flex';
      } else {
        btn.style.display = 'none';
      }
    });
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

    // Update sidebar profile photo
    if (p.profile_photo) {
      const imgEl         = document.getElementById('sidebar-photo');
      const placeholderEl = document.querySelector('.sidebar-avatar-placeholder');
      if (imgEl) {
        imgEl.src           = p.profile_photo;
        imgEl.style.display = 'block';
      }
      if (placeholderEl) placeholderEl.style.display = 'none';
    }
  }

  // Wires every <a data-social="..."> on the page to the matching profile field.
  // If a value is empty the link stays hidden, so the page never shows broken icons.
  function updateSocialLinks(p) {
    document.querySelectorAll('[data-social]').forEach(el => {
      const platform = el.dataset.social;
      let url = '';
      if (platform === 'whatsapp') {
        const num = (p.whatsapp || '').replace(/\D/g, '');
        url = num ? `https://wa.me/${num}` : '';
      } else {
        url = (p[platform] || '').trim();
      }
      if (url) {
        el.href = url;
        el.style.display = '';
      } else {
        el.style.display = 'none';
      }
    });
  }

  return { init };
})();
