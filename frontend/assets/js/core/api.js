/* ============================================================
 * FILE: core/api.js
 * WHAT IT IS:     Central API client — ALL HTTP requests to the
 *                 backend go through this file only.
 * WHY IT EXISTS:  Centralizing API calls means:
 *                 1. JWT token is automatically added to every
 *                    protected request — no handler needs to
 *                    remember to set the Authorization header
 *                 2. 401 responses automatically redirect to login
 *                 3. Base URL is set once — easy to change in production
 *                 4. Error handling is consistent across all requests
 * WHERE USED:     Every page/component that needs backend data
 * IF REMOVED:     All API calls break — nothing can talk to the backend
 * HOW TO USE:
 *   const data = await API.get('/projects');
 *   const created = await API.post('/projects', { name: 'My App' });
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const API = (() => {
  // ─── Configuration ─────────────────────────────────────────
  // BASE_URL is relative — the Go backend serves BOTH the API and the frontend
  // files on the same port, so no absolute URL or CORS issues.
  // In production just deploy the same Go binary and this still works.
  const BASE_URL = '/api/v1';

  /**
   * FUNCTION: getToken
   * WHAT IT DOES: Retrieves the JWT access token from localStorage.
   *               The token is stored there after successful login.
   * WHERE CALLED: request() — on every API call
   * RETURNS: {string|null} — the token string or null if not logged in
   */
  function getToken() {
    return localStorage.getItem('portfolio_token');
  }

  /**
   * FUNCTION: request
   * WHAT IT DOES: The core fetch wrapper that:
   *               1. Constructs the full URL (BASE_URL + endpoint)
   *               2. Adds Content-Type: application/json header
   *               3. Adds Authorization: Bearer <token> if logged in
   *               4. Sends the request and parses the JSON response
   *               5. If the response is 401 (token expired), clears
   *                  localStorage and redirects to /login.html
   *               6. Throws an error for non-OK responses so callers
   *                  can catch them with try/catch
   * WHERE CALLED: API.get(), API.post(), API.put(), API.delete()
   * PARAMETERS:   @param {string} endpoint - the API path e.g. '/projects'
   *               @param {object} options - fetch options (method, body, etc.)
   * RETURNS:      @returns {Promise<object>} - the parsed JSON APIResponse
   */
  async function request(endpoint, options = {}) {
    // ─── Step 1: Build request headers ───────────────────────
    const headers = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    // ─── Step 2: Add JWT token if we have one ────────────────
    // The Authorization header is required for all [protected] routes.
    // Public routes (GET /projects, POST /messages) ignore it.
    const token = getToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    // ─── Step 3: Send the request ─────────────────────────────
    let response;
    try {
      response = await fetch(`${BASE_URL}${endpoint}`, {
        ...options,
        headers,
      });
    } catch (networkErr) {
      // Network error means the server is unreachable (offline, server down)
      throw new Error('Cannot connect to server. Check your internet connection or try again later.');
    }

    // ─── Step 4: Handle 401 Unauthorized ─────────────────────
    // A 401 means the JWT token has expired or is invalid.
    // We clear the saved token and redirect to the login page
    // so the user can get a fresh token.
    if (response.status === 401) {
      localStorage.removeItem('portfolio_token');
      // Only redirect to login if we are on a dashboard page
      if (window.location.pathname.includes('dashboard')) {
        window.location.href = '/login.html';
      }
      throw new Error('Session expired. Please log in again.');
    }

    // ─── Step 5: Parse the JSON response ─────────────────────
    const data = await response.json();

    // ─── Step 6: Throw on API errors ─────────────────────────
    // The backend always returns { success: true/false, message: '...' }
    // If success is false, we throw so callers can show the error message
    if (!response.ok || !data.success) {
      throw new Error(data.message || `Request failed with status ${response.status}`);
    }

    return data;
  }

  // ─── Public methods ──────────────────────────────────────────
  // These are the only functions other files should call.

  /**
   * FUNCTION: API.get
   * WHAT IT DOES: Sends a GET request — used for fetching data
   * EXAMPLE:      const res = await API.get('/projects');
   *               const projects = res.data;
   */
  async function get(endpoint) {
    return request(endpoint, { method: 'GET' });
  }

  /**
   * FUNCTION: API.post
   * WHAT IT DOES: Sends a POST request with a JSON body — used for creating data
   * EXAMPLE:      const res = await API.post('/messages', { name, email, content });
   */
  async function post(endpoint, body) {
    return request(endpoint, {
      method: 'POST',
      body: JSON.stringify(body),
    });
  }

  /**
   * FUNCTION: API.put
   * WHAT IT DOES: Sends a PUT request — used for updating existing data
   * EXAMPLE:      await API.put('/projects/1', updatedProject);
   */
  async function put(endpoint, body) {
    return request(endpoint, {
      method: 'PUT',
      body: JSON.stringify(body),
    });
  }

  /**
   * FUNCTION: API.delete
   * WHAT IT DOES: Sends a DELETE request — used for removing data
   * EXAMPLE:      await API.delete('/projects/1');
   */
  async function del(endpoint) {
    return request(endpoint, { method: 'DELETE' });
  }

  /**
   * FUNCTION: API.upload
   * WHAT IT DOES: Uploads a file using FormData (not JSON).
   *               Used for profile photo, app screenshots, CV PDF.
   * EXAMPLE:
   *   const formData = new FormData();
   *   formData.append('file', fileInput.files[0]);
   *   formData.append('folder', 'images');
   *   const res = await API.upload('/upload', formData);
   *   const url = res.data.url;
   */
  async function upload(endpoint, formData) {
    const token = getToken();
    const headers = {};
    if (token) headers['Authorization'] = `Bearer ${token}`;

    // Note: Do NOT set Content-Type when sending FormData.
    // The browser sets it automatically with the correct boundary string
    // that multipart/form-data requires. Setting it manually breaks uploads.
    const response = await fetch(`${BASE_URL}${endpoint}`, {
      method: 'POST',
      headers,
      body: formData,
    });

    const data = await response.json();
    if (!response.ok || !data.success) {
      throw new Error(data.message || 'Upload failed');
    }
    return data;
  }

  // Expose only the public interface — internal functions are private
  return { get, post, put, delete: del, upload };
})();
