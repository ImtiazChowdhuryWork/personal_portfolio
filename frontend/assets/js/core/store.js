/* ============================================================
 * FILE: core/store.js
 * WHAT IT IS:     Simple reactive state store for the frontend
 * WHY IT EXISTS:  Multiple components need to share and react to
 *                 the same data (e.g. projects list used by both
 *                 the Apps Showcase and the thumbnail strip).
 *                 The Store ensures data is fetched once and shared,
 *                 and components re-render when data changes.
 * WHERE USED:     All page scripts call Store.set/get/subscribe
 * IF REMOVED:     Components must each manage their own state —
 *                 leads to duplicate API calls and inconsistent UI
 * HOW TO USE:
 *   Store.set('projects', data);
 *   const projects = Store.get('projects');
 *   Store.subscribe('projects', (data) => { renderProjects(data); });
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const Store = (() => {
  // ─── Internal state ───────────────────────────────────────────
  // The actual data lives here — outside the module, only accessible
  // through get/set methods to prevent accidental direct mutation
  const _state = {};

  // Listeners maps a key → array of callback functions.
  // When Store.set('projects', data) is called, all callbacks
  // registered for 'projects' are called with the new data.
  const _listeners = {};

  /**
   * FUNCTION: set
   * WHAT IT DOES: Saves data under a key in the store, then notifies
   *               all components that subscribed to that key.
   * WHERE CALLED: Page scripts after API calls return data
   * PARAMETERS:   @param {string} key - the data key e.g. 'projects'
   *               @param {*} value - the data to store
   * EXAMPLE:      Store.set('profile', profileData);
   */
  function set(key, value) {
    _state[key] = value;
    // Notify all subscribers about the new value
    if (_listeners[key]) {
      _listeners[key].forEach(cb => cb(value));
    }
  }

  /**
   * FUNCTION: get
   * WHAT IT DOES: Retrieves the current value stored under a key.
   *               Returns undefined if the key has never been set.
   * WHERE CALLED: Components that need data synchronously
   * PARAMETERS:   @param {string} key - the data key
   * RETURNS:      @returns {*} - the stored value
   * EXAMPLE:      const projects = Store.get('projects');
   */
  function get(key) {
    return _state[key];
  }

  /**
   * FUNCTION: subscribe
   * WHAT IT DOES: Registers a callback that will be called whenever
   *               the value at 'key' changes via Store.set().
   *               If data already exists, the callback fires immediately
   *               with the current value — useful for components that
   *               mount after data is already loaded.
   * WHERE CALLED: Components on initialization to react to data changes
   * PARAMETERS:   @param {string} key - the data key to watch
   *               @param {function} callback - called with new value on change
   * EXAMPLE:
   *   Store.subscribe('projects', (projects) => {
   *     AppShowcase.render({ apps: projects });
   *   });
   */
  function subscribe(key, callback) {
    if (!_listeners[key]) {
      _listeners[key] = [];
    }
    _listeners[key].push(callback);

    // If data already exists, fire immediately so the component
    // doesn't have to wait for the next Store.set() call
    if (_state[key] !== undefined) {
      callback(_state[key]);
    }
  }

  /**
   * FUNCTION: clear
   * WHAT IT DOES: Removes a key from the store. Called on logout
   *               to clear auth-related state.
   * PARAMETERS:   @param {string} key - the key to remove
   */
  function clear(key) {
    delete _state[key];
  }

  return { set, get, subscribe, clear };
})();
