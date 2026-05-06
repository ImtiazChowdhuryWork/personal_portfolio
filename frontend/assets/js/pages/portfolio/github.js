/* ============================================================
 * FILE: pages/portfolio/github.js
 * WHAT IT IS:     GitHub stats section renderer
 * WHERE USED:     main.js → called on DOMContentLoaded
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const GithubSection = (() => {

  const GITHUB_USERNAME = 'imtiazchowdhury'; // Update with real username

  function init() {
    const container = document.getElementById('github-content');
    if (!container) return;

    container.innerHTML = `
      <div class="github-stats-grid">
        <div class="github-stat-card">
          <div class="github-stat-icon">📦</div>
          <div class="github-stat-value">20+</div>
          <div class="github-stat-label">Repositories</div>
        </div>
        <div class="github-stat-card">
          <div class="github-stat-icon">✅</div>
          <div class="github-stat-value">500+</div>
          <div class="github-stat-label">Total Commits</div>
        </div>
        <div class="github-stat-card">
          <div class="github-stat-icon">⭐</div>
          <div class="github-stat-value">Dart</div>
          <div class="github-stat-label">Top Language</div>
        </div>
        <div class="github-stat-card">
          <div class="github-stat-icon">🔥</div>
          <div class="github-stat-value">2.5+</div>
          <div class="github-stat-label">Years Active</div>
        </div>
      </div>

      <div class="github-embeds">
        <img class="github-embed-img"
          src="https://github-readme-stats.vercel.app/api?username=${GITHUB_USERNAME}&show_icons=true&theme=tokyonight&hide_border=true&bg_color=0a0a0f&title_color=54c5f8&icon_color=54c5f8&text_color=e8e8f0"
          alt="GitHub Stats"
          onerror="this.style.display='none';">
        <img class="github-embed-img"
          src="https://github-readme-stats.vercel.app/api/top-langs/?username=${GITHUB_USERNAME}&layout=compact&theme=tokyonight&hide_border=true&bg_color=0a0a0f&title_color=54c5f8&text_color=e8e8f0"
          alt="Top Languages"
          onerror="this.style.display='none';">
      </div>

      <div class="github-link">
        <a href="https://github.com/${GITHUB_USERNAME}" target="_blank" rel="noopener" class="btn btn-outline">
          View GitHub Profile →
        </a>
      </div>
    `;
  }

  return { init };
})();
