/* ============================================================
 * FILE: core/utils.js
 * WHAT IT IS:     General utility helper functions
 * WHERE USED:     Any JS file that needs these helpers
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const Utils = (() => {

  // Tracks per-element rAF ids so a fresh animation cancels any in-flight one
  // on the same element (prevents two loops fighting over textContent when the
  // profile fetch updates a stat mid-animation).
  const _animationFrames = new WeakMap();

  /**
   * Animates a number counter from 0 to the target value.
   * Used for the hero stats section (2.5+ years, 5+ apps etc.)
   * @param {HTMLElement} el - the element to update
   * @param {number} target - the final number to count to
   * @param {number} duration - how long the animation takes in ms
   * @param {string} suffix - appended after the number ('+', '%', etc.)
   * @param {number} decimals - decimal places to render (1 → 0.0, 0.5, 1.0…)
   */
  function animateCounter(el, target, duration = 1800, suffix = '', decimals = 0) {
    const existing = _animationFrames.get(el);
    if (existing) cancelAnimationFrame(existing);

    const targetNum = parseFloat(target) || 0;
    const start = performance.now();
    const format = (n) => decimals > 0 ? n.toFixed(decimals) : String(Math.round(n));
    const update = (time) => {
      const elapsed = time - start;
      const progress = Math.min(elapsed / duration, 1);
      // easeOutQuart — gentler tail than cubic, no harsh snap at the end
      const eased = 1 - Math.pow(1 - progress, 4);
      const current = eased * targetNum;
      el.textContent = format(current) + suffix;
      if (progress < 1) {
        _animationFrames.set(el, requestAnimationFrame(update));
      } else {
        // Land exactly on the target so rounding doesn't leave us at "2.4+"
        el.textContent = format(targetNum) + suffix;
        _animationFrames.delete(el);
      }
    };
    _animationFrames.set(el, requestAnimationFrame(update));
  }

  /**
   * Formats a date string for display.
   * @param {string} dateStr - ISO date string or "Jan 2024"
   * @returns {string} human-readable date
   */
  function formatDate(dateStr) {
    if (!dateStr) return 'Present';
    return dateStr;
  }

  /**
   * Generates initials from a full name for avatar placeholders.
   * "Chowdhury Md. Imtiazul Islam" → "CI"
   * @param {string} name
   * @returns {string} 1-2 uppercase initials
   */
  function getInitials(name) {
    if (!name) return '?';
    const parts = name.trim().split(' ').filter(p => p.length > 0);
    if (parts.length === 1) return parts[0][0].toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }

  /**
   * Truncates a string to a max length and adds "..." if needed.
   * @param {string} str
   * @param {number} maxLen
   * @returns {string}
   */
  function truncate(str, maxLen = 100) {
    if (!str || str.length <= maxLen) return str || '';
    return str.slice(0, maxLen).trim() + '...';
  }

  /**
   * Debounces a function — waits for 'delay' ms of inactivity
   * before calling it. Prevents rapid-fire calls on scroll/resize.
   * @param {function} fn
   * @param {number} delay ms
   * @returns {function}
   */
  function debounce(fn, delay = 200) {
    let timer;
    return (...args) => {
      clearTimeout(timer);
      timer = setTimeout(() => fn(...args), delay);
    };
  }

  /**
   * Checks if an element is currently visible in the viewport.
   * Used to trigger animations when the user scrolls to a section.
   * @param {HTMLElement} el
   * @param {number} offset - px from bottom of viewport to trigger
   * @returns {boolean}
   */
  function isInViewport(el, offset = 100) {
    const rect = el.getBoundingClientRect();
    return rect.top < (window.innerHeight - offset) && rect.bottom > 0;
  }

  /**
   * Parses a JSON string array safely.
   * Returns [] if the string is null, empty, or invalid JSON.
   * @param {string|Array} val
   * @returns {Array}
   */
  function parseArray(val) {
    if (!val) return [];
    if (Array.isArray(val)) return val;
    try { return JSON.parse(val); }
    catch { return []; }
  }

  return { animateCounter, formatDate, getInitials, truncate, debounce, isInViewport, parseArray };
})();
