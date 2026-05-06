/* ============================================================
 * FILE: pages/portfolio/architecture.js
 * WHAT IT IS:     Architecture Expertise section renderer
 * WHERE USED:     main.js → called on DOMContentLoaded
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const ArchitectureSection = (() => {

  // Static data — architecture knowledge doesn't change via API
  const architectures = [
    {
      num: '01',
      name: 'Clean Architecture',
      diagram: `UI Layer
    ↓
Domain Layer (Use Cases)
    ↓
Data Layer (Repositories)
    ↓
  Database / API`,
      desc: 'Separates the app into layers with clear boundaries. Business logic in the domain layer never depends on UI or data frameworks — making it independently testable and framework-agnostic.',
      projects: ['Project Finder', 'Hiye Health'],
    },
    {
      num: '02',
      name: 'BLoC Pattern',
      diagram: `    Events
      ↓
[BLoC / Cubit]
      ↓
    States
      ↓
    UI Widgets`,
      desc: 'Business Logic Component separates UI from business logic using streams. Events flow in, states flow out. Every state change is explicit, predictable, and testable.',
      projects: ['Hiye Health'],
    },
    {
      num: '03',
      name: 'GetX Pattern',
      diagram: `  Controllers
  (logic + state)
      ↓
  GetX Bindings
  (DI + routing)
      ↓
  Obx Widgets
  (reactive UI)`,
      desc: 'Lightweight state management with built-in routing and dependency injection. Reactive variables (Rx) automatically update the UI without manual setState or streams.',
      projects: ['Project Finder'],
    },
    {
      num: '04',
      name: 'MVVM',
      diagram: `   View (Widget)
       ↕
  ViewModel
  (ChangeNotifier)
       ↕
  Model / Repo`,
      desc: 'Model-View-ViewModel keeps UI code clean by moving all logic into the ViewModel. The View only observes the ViewModel — never makes decisions.',
      projects: ['SperkTech Apps'],
    },
    {
      num: '05',
      name: 'Repository Pattern',
      diagram: `UI / BLoC / GetX
       ↓
   Repository
  (interface)
    ↙      ↘
RemoteDS  LocalDS
(API)   (Hive/SQLite)`,
      desc: 'The Repository acts as the single source of truth, deciding whether to fetch fresh data from the API or serve cached local data — completely transparent to the UI layer.',
      projects: ['Project Finder', 'Hiye Health'],
    },
  ];

  function init() {
    const container = document.getElementById('architecture-grid');
    if (!container) return;

    container.innerHTML = `<div class="arch-grid">${architectures.map(buildCard).join('')}</div>`;
  }

  function buildCard(arch) {
    const projects = arch.projects.map(p => `<span class="arch-project-tag">${p}</span>`).join('');
    return `
      <div class="arch-card">
        <div class="arch-card-header">
          <div class="arch-card-num">${arch.num}</div>
          <h3 class="arch-card-name">${arch.name}</h3>
        </div>
        <pre class="arch-diagram">${arch.diagram}</pre>
        <p class="arch-desc">${arch.desc}</p>
        <div class="arch-projects">${projects}</div>
      </div>
    `;
  }

  return { init };
})();
