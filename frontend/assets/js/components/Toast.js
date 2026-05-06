/* ============================================================
 * FILE: components/Toast.js
 * WHAT IT IS:     Toast notification component
 * WHY IT EXISTS:  Shows success/error feedback after actions
 *                 (form submitted, API error, etc.) without page reload
 * WHERE USED:     Any page script that needs user feedback
 * HOW TO USE:
 *   Toast.show('Project created!', 'success');
 *   Toast.show('Something went wrong', 'error');
 *   Toast.show('Check your email', 'info');
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const Toast = (() => {
  // Create the container on first use
  let container;

  function getContainer() {
    if (!container) {
      container = document.createElement('div');
      container.id = 'toast-container';
      document.body.appendChild(container);
    }
    return container;
  }

  const icons = {
    success: '✓',
    error:   '✕',
    warning: '⚠',
    info:    'ℹ',
  };

  /**
   * FUNCTION: show
   * WHAT IT DOES: Creates and displays a toast notification.
   *               Auto-dismisses after 'duration' ms.
   * @param {string} message - the text to show
   * @param {string} type    - 'success' | 'error' | 'warning' | 'info'
   * @param {number} duration - auto-dismiss delay in ms (default 4000)
   */
  function show(message, type = 'info', duration = 4000) {
    const c = getContainer();
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `
      <span class="toast-icon">${icons[type] || icons.info}</span>
      <div class="toast-body">
        <p class="toast-msg">${message}</p>
      </div>
      <button class="toast-close" aria-label="Close">×</button>
    `;

    c.appendChild(toast);
    // Trigger the slide-in animation on next frame
    requestAnimationFrame(() => toast.classList.add('show'));

    // Auto-dismiss
    const timer = setTimeout(() => dismiss(toast), duration);

    // Manual close button
    toast.querySelector('.toast-close').addEventListener('click', () => {
      clearTimeout(timer);
      dismiss(toast);
    });
  }

  function dismiss(toast) {
    toast.classList.remove('show');
    // Remove from DOM after the slide-out animation finishes
    setTimeout(() => toast.remove(), 400);
  }

  return { show };
})();
