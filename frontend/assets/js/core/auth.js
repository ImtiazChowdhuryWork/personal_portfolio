/* ============================================================
 * FILE: core/auth.js
 * WHAT IT IS:     Authentication state management
 * WHY IT EXISTS:  Dashboard pages need to verify the user is logged
 *                 in before loading. This file checks the stored
 *                 JWT and redirects to login if it's missing.
 * WHERE USED:     dashboard.html includes this to guard the page
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const Auth = (() => {

  const TOKEN_KEY = 'portfolio_token';

  /** Saves the JWT token to localStorage after login. */
  function setToken(token) {
    localStorage.setItem(TOKEN_KEY, token);
  }

  /** Retrieves the stored JWT token. Returns null if not logged in. */
  function getToken() {
    return localStorage.getItem(TOKEN_KEY);
  }

  /** Removes the JWT token (logout). */
  function clearToken() {
    localStorage.removeItem(TOKEN_KEY);
  }

  /** Returns true if a token exists (user appears to be logged in). */
  function isLoggedIn() {
    return !!getToken();
  }

  /**
   * FUNCTION: requireAuth
   * WHAT IT DOES: Called at the top of dashboard pages. If no token
   *               is found, redirects immediately to login.html.
   *               If a token exists, verifies it with the backend.
   *               On 401 response, clears the stale token and redirects.
   */
  async function requireAuth() {
    if (!isLoggedIn()) {
      window.location.href = '/login';
      return false;
    }
    // Verify the token is still valid with the backend
    try {
      await API.get('/auth/me');
      return true;
    } catch {
      clearToken();
      window.location.href = '/login';
      return false;
    }
  }

  /**
   * FUNCTION: login
   * WHAT IT DOES: Calls the login API and stores the token on success.
   * @param {string} email
   * @param {string} password
   * @returns {Promise<object>} user data
   */
  async function login(email, password) {
    clearToken(); // remove any stale token before requesting a new one
    const res = await API.post('/auth/login', { email, password });
    setToken(res.data.token);
    return res.data.user;
  }

  /**
   * FUNCTION: logout
   * WHAT IT DOES: Calls the logout endpoint and clears the local token,
   *               then redirects to login page.
   */
  async function logout() {
    try { await API.post('/auth/logout', {}); } catch { /* ignore */ }
    clearToken();
    window.location.href = '/login';
  }

  return { setToken, getToken, clearToken, isLoggedIn, requireAuth, login, logout };
})();
