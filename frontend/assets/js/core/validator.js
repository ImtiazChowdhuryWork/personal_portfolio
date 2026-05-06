/* ============================================================
 * FILE: core/validator.js
 * WHAT IT IS:     Frontend form validation helpers
 * WHERE USED:     contact form, login form
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const Validator = (() => {

  function isEmail(val) {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(val.trim());
  }

  function isEmpty(val) {
    return !val || val.trim() === '';
  }

  function minLength(val, min) {
    return val && val.trim().length >= min;
  }

  /**
   * Validates a form by checking rules on each field.
   * @param {object} rules - { fieldName: { required, email, minLength, label } }
   * @param {object} data  - { fieldName: value }
   * @returns {object} { valid: boolean, errors: { fieldName: errorMsg } }
   */
  function validate(rules, data) {
    const errors = {};
    for (const [field, rule] of Object.entries(rules)) {
      const val = data[field] || '';
      const label = rule.label || field;

      if (rule.required && isEmpty(val)) {
        errors[field] = `${label} is required`;
      } else if (rule.email && val && !isEmail(val)) {
        errors[field] = `Please enter a valid email address`;
      } else if (rule.minLength && val && !minLength(val, rule.minLength)) {
        errors[field] = `${label} must be at least ${rule.minLength} characters`;
      }
    }
    return { valid: Object.keys(errors).length === 0, errors };
  }

  /**
   * Shows validation errors on a form's input elements.
   * Adds .is-invalid class and inserts a .form-error element.
   * @param {HTMLFormElement} form
   * @param {object} errors - { fieldName: errorMessage }
   */
  function showErrors(form, errors) {
    // Clear previous errors first
    form.querySelectorAll('.is-invalid').forEach(el => el.classList.remove('is-invalid'));
    form.querySelectorAll('.form-error').forEach(el => el.remove());

    for (const [field, msg] of Object.entries(errors)) {
      const input = form.querySelector(`[name="${field}"]`);
      if (input) {
        input.classList.add('is-invalid');
        const err = document.createElement('p');
        err.className = 'form-error';
        err.textContent = msg;
        input.parentNode.insertBefore(err, input.nextSibling);
      }
    }
  }

  function clearErrors(form) {
    form.querySelectorAll('.is-invalid').forEach(el => el.classList.remove('is-invalid'));
    form.querySelectorAll('.form-error').forEach(el => el.remove());
  }

  return { isEmail, isEmpty, minLength, validate, showErrors, clearErrors };
})();
