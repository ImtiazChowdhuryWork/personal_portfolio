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
      updateFooter(profile);
      applyStatusBehavior(profile);
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

    // Show CV buttons only if a CV exists AND the admin hasn't hidden the button.
    // Click fires a fire-and-forget counter ping so the dashboard sees download stats.
    const visible = p.cv_visible !== false;
    ['cv-download-btn-1', 'cv-download-btn-2'].forEach(id => {
      const btn = document.getElementById(id);
      if (!btn) return;
      if (p.cv_file && visible) {
        btn.href = p.cv_file;
        btn.setAttribute('download', p.cv_file.split('/').pop());
        btn.style.display = 'inline-flex';
        btn.addEventListener('click', recordCVDownload, { once: false });
      } else {
        btn.style.display = 'none';
      }
    });
  }

  function recordCVDownload() {
    // Fire-and-forget — never block the actual download.
    try { API.post('/profile/cv/download', {}); } catch { /* ignore */ }
  }

  function updateHeroSection(p) {
    const shortBio = document.getElementById('hero-short-bio');
    if (shortBio && p.short_bio) shortBio.textContent = p.short_bio;

    // Hero subtitle pill resolution:
    //   1. If `hero_subtitle` is set, use it as a template and resolve placeholders
    //   2. Otherwise, fall back to "Say Hi from {nickname || first_name}, {title}"
    const fullName = (p.full_name || '').trim();
    const firstName = fullName.split(/\s+/)[0] || '';
    const nickname = (p.nickname || '').trim() || firstName;
    const title = (p.title || '').trim();
    const custom = (p.hero_subtitle || '').trim();
    const subtitleText = custom
      ? custom
          .replace(/\{nickname\}/g, nickname)
          .replace(/\{name\}/g, fullName)
          .replace(/\{title\}/g, title)
      : `Say Hi from ${nickname}, ${title}`;

    const subtitleTextEl = document.getElementById('hero-subtitle-text');
    if (subtitleTextEl) subtitleTextEl.textContent = subtitleText;

    // Hero heading — render from line1/line2/line3 + highlights, or override
    const heading = document.getElementById('hero-title');
    if (heading) {
      const html = renderHeroHeadingHTML({
        line1:      p.hero_heading_line1,
        line2:      p.hero_heading_line2,
        line3:      p.hero_heading_line3,
        highlights: p.hero_heading_highlights,
        override:   p.hero_heading_override,
      });
      if (html) heading.innerHTML = html;
    }
  }

  // Renders the hero heading HTML from profile fields. Mirrors the helper in
  // dashboard.html so dashboard preview and live portfolio always match.
  function renderHeroHeadingHTML({ line1, line2, line3, highlights, override }) {
    const escapeHTML = (s) => String(s == null ? '' : s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;').replace(/'/g, '&#39;');
    const escapeRegex = (s) => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    // ASCII-safe markers (avoiding \x00 which some HTML parsers strip)
    const HOPEN  = '__HL_OPEN_KP__';
    const HCLOSE = '__HL_CLOSE_KP__';
    const finalize = (s) => escapeHTML(s)
      .split(HOPEN).join('<span class="highlight">')
      .split(HCLOSE).join('</span>');

    if (override && override.trim()) {
      const marked = override.replace(/\*([^*\n]+)\*/g, HOPEN + '$1' + HCLOSE);
      return marked.split('\n').map(finalize).join('<br>');
    }
    const words = (highlights || '').split(',').map(w => w.trim()).filter(Boolean)
      .sort((a, b) => b.length - a.length);
    const markup = (line) => {
      if (!line) return '';
      let raw = line;
      words.forEach(w => {
        const re = new RegExp(escapeRegex(w), 'gi');
        raw = raw.replace(re, m => HOPEN + m + HCLOSE);
      });
      return finalize(raw);
    };
    return [line1, line2, line3].map(markup).filter(Boolean).join('<br>');
  }

  function updateSidebar(p) {
    const nameEl = document.getElementById('sidebar-name');
    if (nameEl && p.full_name) nameEl.textContent = p.full_name;

    const designationEl = document.getElementById('sidebar-designation');
    if (designationEl && p.title) designationEl.textContent = p.title;

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

  // Footer tagline: "{title} · {location}", with each side optional.
  function updateFooter(p) {
    const tagline = document.getElementById('footer-tagline');
    if (tagline) {
      const parts = [p.title, p.location].map(v => (v || '').trim()).filter(Boolean);
      if (parts.length) tagline.textContent = parts.join(' · ');
    }

    // Copyright resolution chain:
    //   sidebar = sidebar_copyright_text || copyright_text || auto-fallback
    //   footer  = footer_copyright_text  || copyright_text || auto-fallback
    // All three custom fields support {year} and {name} placeholders.
    const fullName = (p.full_name || '').trim();
    const fallback = `© {year} {name}. All Rights Reserved.`;

    const main         = (p.copyright_text         || '').trim();
    const sidebarRaw   = (p.sidebar_copyright_text || '').trim() || main || fallback;
    const footerRaw    = (p.footer_copyright_text  || '').trim() || main || fallback;
    const builtWithRaw = (p.footer_built_with      || '').trim();

    const resolve = (text) => String(text)
      .replace(/\{year\}/g, new Date().getFullYear())
      .replace(/\{name\}/g, fullName);

    const sidebarCopy = document.getElementById('sidebar-copy');
    if (sidebarCopy) sidebarCopy.textContent = resolve(sidebarRaw);

    const footerCopy = document.getElementById('footer-copy');
    if (footerCopy) footerCopy.textContent = resolve(footerRaw);

    const builtWithEl = document.getElementById('footer-built-with');
    if (builtWithEl) {
      if (builtWithRaw) {
        builtWithEl.textContent = resolve(builtWithRaw);
        builtWithEl.style.display = '';
      } else {
        builtWithEl.style.display = 'none';
      }
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

  // Per-status behaviour for hire-related buttons across the portfolio.
  // Drives:
  //   - Sidebar Hire Me button (label + disabled state)
  //   - Hero CTA buttons (which show)
  //   - Contact "Job" / "Freelance" cards (which show)
  //   - Contact form type select (which options remain)
  //   - Status banner above the contact cards (text + color)
  function applyStatusBehavior(p) {
    const status = (p.availability || 'Open to Work').trim();

    const config = {
      'Open to Work': {
        sidebarLabel: 'Hire Me!',
        sidebarDisabled: false,
        showHeroJob: true,
        showHeroFreelance: true,
        showContactJob: true,
        showContactFreelance: true,
        formTypes: ['Job Opportunity', 'Freelance', 'Other'],
        bannerClass: 'is-open',
        bannerIcon: 'ph-handshake',
        bannerText: 'Open to new opportunities — let’s talk.',
      },
      'Available for Freelance': {
        sidebarLabel: 'Start a Project',
        sidebarDisabled: false,
        showHeroJob: false,
        showHeroFreelance: true,
        showContactJob: false,
        showContactFreelance: true,
        formTypes: ['Freelance', 'Other'],
        bannerClass: 'is-freelance',
        bannerIcon: 'ph-rocket-launch',
        bannerText: 'Currently taking on freelance projects.',
      },
      'On Vacation': {
        sidebarLabel: 'Currently Away',
        sidebarDisabled: true,
        showHeroJob: false,
        showHeroFreelance: false,
        showContactJob: false,
        showContactFreelance: false,
        formTypes: ['Other'],
        bannerClass: 'is-away',
        bannerIcon: 'ph-tree-palm',
        bannerText: 'Currently on vacation. Feel free to send a message — I’ll respond when I’m back.',
      },
    }[status] || null;
    if (!config) return;

    // Sidebar Hire Me button
    const sidebarBtn = document.getElementById('sidebar-hire-btn');
    const sidebarLabel = document.getElementById('sidebar-hire-btn-label');
    if (sidebarLabel) sidebarLabel.textContent = config.sidebarLabel;
    if (sidebarBtn) sidebarBtn.classList.toggle('is-disabled', config.sidebarDisabled);

    // Hero CTAs
    const heroJob       = document.getElementById('hero-cta-job');
    const heroFreelance = document.getElementById('hero-cta-freelance');
    if (heroJob)       heroJob.style.display       = config.showHeroJob       ? '' : 'none';
    if (heroFreelance) heroFreelance.style.display = config.showHeroFreelance ? '' : 'none';

    // Contact cards
    const cardJob       = document.getElementById('contact-card-job');
    const cardFreelance = document.getElementById('contact-card-freelance');
    if (cardJob)       cardJob.style.display       = config.showContactJob       ? '' : 'none';
    if (cardFreelance) cardFreelance.style.display = config.showContactFreelance ? '' : 'none';

    // Contact form type — filter <option>s to those listed in config.formTypes
    const typeSelect = document.getElementById('contact-type');
    if (typeSelect) {
      Array.from(typeSelect.options).forEach(opt => {
        const keep = config.formTypes.some(t =>
          opt.value === t || opt.value.startsWith(t) || t.startsWith(opt.value));
        opt.hidden = !keep;
        opt.disabled = !keep;
      });
      // If the currently selected option got hidden, switch to the first kept one
      if (typeSelect.options[typeSelect.selectedIndex]?.hidden) {
        const firstKept = Array.from(typeSelect.options).find(o => !o.hidden);
        if (firstKept) typeSelect.value = firstKept.value;
      }
    }

    // Status banner above contact cards
    const banner = document.getElementById('status-banner');
    const bannerText = document.getElementById('status-banner-text');
    const bannerIcon = document.getElementById('status-banner-icon');
    if (banner && bannerText && bannerIcon) {
      banner.className = 'status-banner ' + config.bannerClass;
      bannerIcon.className = 'ph ' + config.bannerIcon;
      bannerText.textContent = config.bannerText;
      banner.style.display = 'flex';
    }
  }

  return { init };
})();
