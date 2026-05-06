/* ============================================================
 * FILE: pages/portfolio/services.js
 * WHAT IT IS:     "What I Offer" section renderer (static content)
 * WHERE USED:     main.js — services section is static (no API call)
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const ServicesSection = (() => {

  const services = [
    {
      icon: '📱',
      title: 'Mobile App Development',
      desc: 'I build production-ready Flutter apps for iOS and Android from scratch to App Store launch. Clean code, clean architecture, real results.',
    },
    {
      icon: '🔧',
      title: 'App Maintenance & Updates',
      desc: 'I maintain, debug, and improve existing Flutter apps. Performance optimization, dependency upgrades, new feature integration.',
    },
    {
      icon: '🔌',
      title: 'API & Service Integration',
      desc: 'I integrate REST APIs, Firebase, payment gateways (Stripe, RevenueCat), Socket.IO, WebRTC, and third-party SDKs into Flutter apps.',
    },
  ];

  function init() {
    const container = document.getElementById('services-grid');
    if (!container) return;

    container.innerHTML = services.map(s => `
      <div class="card">
        <div class="card-icon">${s.icon}</div>
        <h3 class="card-title">${s.title}</h3>
        <p class="card-desc">${s.desc}</p>
      </div>
    `).join('');
  }

  return { init };
})();
