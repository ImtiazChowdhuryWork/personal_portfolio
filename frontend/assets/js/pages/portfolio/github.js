/* ============================================================
 * FILE: pages/portfolio/github.js
 * WHAT IT IS:     GitHub stats section renderer
 * HOW IT WORKS:   Reads the saved profile (already fetched by
 *                 AboutSection) and combines two sources for each
 *                 stat card:
 *                   1. A manual override from the Profile if non-empty
 *                   2. Otherwise, a value computed from the GitHub API
 *                   3. Otherwise, a static fallback so the page never
 *                      shows blanks
 * WHERE USED:     main.js → called on DOMContentLoaded
 * LAST UPDATED:   2026-05-09 — combined dynamic + manual override
 * ============================================================ */

const GithubSection = (() => {

  const FALLBACKS = {
    repos: '20+',
    commits: '500+',
    topLanguage: 'Dart',
    yearsActive: '2.5+',
  };

  // Render the static skeleton so the page has its layout immediately,
  // then asynchronously upgrade values once the profile and API resolve.
  function init() {
    const container = document.getElementById('github-content');
    if (!container) return;

    container.innerHTML = `
      <div class="github-stats-grid scroll-animation" data-animation="fade_from_bottom">
        <div class="github-stat-card">
          <div class="github-stat-icon">📦</div>
          <div class="github-stat-value" id="gh-repos">${FALLBACKS.repos}</div>
          <div class="github-stat-label">Repositories</div>
        </div>
        <div class="github-stat-card">
          <div class="github-stat-icon">✅</div>
          <div class="github-stat-value" id="gh-commits">${FALLBACKS.commits}</div>
          <div class="github-stat-label">Total Commits</div>
        </div>
        <div class="github-stat-card">
          <div class="github-stat-icon">⭐</div>
          <div class="github-stat-value" id="gh-top-language">${FALLBACKS.topLanguage}</div>
          <div class="github-stat-label">Top Language</div>
        </div>
        <div class="github-stat-card">
          <div class="github-stat-icon">🔥</div>
          <div class="github-stat-value" id="gh-years-active">${FALLBACKS.yearsActive}</div>
          <div class="github-stat-label">Years Active</div>
        </div>
      </div>

      <div class="github-embeds scroll-animation" data-animation="fade_from_bottom">
        <img class="github-embed-img" id="gh-embed-stats" alt="GitHub Stats" style="display:none">
        <img class="github-embed-img" id="gh-embed-langs" alt="Top Languages" style="display:none">
      </div>

      <div class="github-link scroll-animation" data-animation="fade_from_bottom">
        <a id="gh-profile-link" href="#" target="_blank" rel="noopener" class="btn btn-outline" style="display:none">
          View GitHub Profile →
        </a>
      </div>
    `;

    hydrate();
  }

  // Pull the profile (already cached by AboutSection.init) and decide what
  // to show in each card. The API call only happens if we have a username.
  async function hydrate() {
    const profile = await waitForProfile();
    const username = (profile && profile.github_username || '').trim();

    if (!username) {
      // No username configured — keep static fallbacks. Manual overrides may
      // still apply if the admin set them on the dashboard.
      applyOverrides(profile);
      return;
    }

    // Fire both API requests in parallel; embed images don't need awaiting
    setEmbeds(username);
    setProfileLink(username);

    // Auto-mode values default to the fallbacks until GitHub responds
    const auto = { ...FALLBACKS };
    try {
      const [user, repos] = await Promise.all([
        fetchUser(username),
        fetchRepos(username),
      ]);
      if (user) {
        if (typeof user.public_repos === 'number') auto.repos = String(user.public_repos);
        if (user.created_at) auto.yearsActive = computeYears(user.created_at);
      }
      if (repos && repos.length) {
        const top = computeTopLanguage(repos);
        if (top) auto.topLanguage = top;
      }
    } catch {
      // Network error / rate limit — silently keep fallbacks
    }

    // Manual overrides (non-empty) win over auto values
    applyOverrides(profile, auto);
  }

  function applyOverrides(profile, auto) {
    auto = auto || { ...FALLBACKS };
    const pick = (override, autoValue) => {
      const o = (override || '').trim();
      return o !== '' ? o : autoValue;
    };
    setText('gh-repos',         pick(profile && profile.github_repos,        auto.repos));
    setText('gh-commits',       pick(profile && profile.github_commits,      auto.commits));
    setText('gh-top-language',  pick(profile && profile.github_top_language, auto.topLanguage));
    setText('gh-years-active',  pick(profile && profile.github_years_active, auto.yearsActive));
  }

  function setEmbeds(username) {
    const stats = document.getElementById('gh-embed-stats');
    const langs = document.getElementById('gh-embed-langs');
    const u = encodeURIComponent(username);
    if (stats) {
      stats.src = `https://github-readme-stats.vercel.app/api?username=${u}&show_icons=true&theme=tokyonight&hide_border=true&bg_color=0a0a0f&title_color=54c5f8&icon_color=54c5f8&text_color=e8e8f0`;
      stats.onerror = () => stats.style.display = 'none';
      stats.style.display = '';
    }
    if (langs) {
      langs.src = `https://github-readme-stats.vercel.app/api/top-langs/?username=${u}&layout=compact&theme=tokyonight&hide_border=true&bg_color=0a0a0f&title_color=54c5f8&text_color=e8e8f0`;
      langs.onerror = () => langs.style.display = 'none';
      langs.style.display = '';
    }
  }

  function setProfileLink(username) {
    const a = document.getElementById('gh-profile-link');
    if (!a) return;
    a.href = `https://github.com/${encodeURIComponent(username)}`;
    a.style.display = '';
  }

  async function fetchUser(username) {
    const r = await fetch(`https://api.github.com/users/${encodeURIComponent(username)}`);
    if (!r.ok) return null;
    return r.json();
  }

  // Sort by recently updated so the top language reflects current focus.
  // 100 is the max page size; portfolios rarely have more public repos.
  async function fetchRepos(username) {
    const r = await fetch(`https://api.github.com/users/${encodeURIComponent(username)}/repos?per_page=100&sort=updated`);
    if (!r.ok) return [];
    return r.json();
  }

  function computeTopLanguage(repos) {
    const counts = {};
    for (const r of repos) {
      if (!r || r.fork) continue; // forks usually aren't your code
      const lang = r.language;
      if (!lang) continue;
      counts[lang] = (counts[lang] || 0) + 1;
    }
    let best = null;
    let bestCount = -1;
    for (const [lang, n] of Object.entries(counts)) {
      if (n > bestCount) { best = lang; bestCount = n; }
    }
    return best;
  }

  function computeYears(createdAt) {
    const created = new Date(createdAt).getTime();
    if (Number.isNaN(created)) return null;
    const years = (Date.now() - created) / (365.25 * 24 * 60 * 60 * 1000);
    if (years < 1) return '<1';
    return years.toFixed(1).replace(/\.0$/, '') + '+';
  }

  // The profile is fetched by AboutSection.init() and stored. Wait briefly
  // for it so this file doesn't depend on init order.
  function waitForProfile(timeoutMs = 3000) {
    return new Promise(resolve => {
      if (typeof Store === 'undefined') return resolve(null);
      const cached = Store.get('profile');
      if (cached) return resolve(cached);
      const unsub = Store.subscribe('profile', (p) => {
        if (typeof unsub === 'function') unsub();
        resolve(p);
      });
      setTimeout(() => resolve(Store.get('profile') || null), timeoutMs);
    });
  }

  function setText(id, value) {
    const el = document.getElementById(id);
    if (el && value !== null && value !== undefined) el.textContent = value;
  }

  return { init };
})();
