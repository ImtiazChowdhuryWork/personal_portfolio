/* ============================================================
 * FILE: pages/portfolio/contact.js
 * WHAT IT IS:     Contact section form handler
 * WHERE USED:     main.js → called on DOMContentLoaded
 * LAST UPDATED:   2026-05-07 — initial creation
 * ============================================================ */

const ContactSection = (() => {

  function init() {
    const form = document.getElementById('contact-form');
    if (!form) return;

    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      Validator.clearErrors(form);

      // Read form values
      const data = {
        name:    form.querySelector('[name="name"]').value,
        email:   form.querySelector('[name="email"]').value,
        type:    form.querySelector('[name="type"]').value,
        content: form.querySelector('[name="content"]').value,
      };

      // Validate
      const { valid, errors } = Validator.validate({
        name:    { required: true, label: 'Name' },
        email:   { required: true, email: true, label: 'Email' },
        content: { required: true, minLength: 10, label: 'Message' },
      }, data);

      if (!valid) {
        Validator.showErrors(form, errors);
        return;
      }

      // Submit
      const btn = form.querySelector('[type="submit"]');
      btn.classList.add('loading');
      btn.disabled = true;

      try {
        await API.post('/messages', data);
        form.reset();
        Toast.show('Message sent! I\'ll get back to you soon.', 'success', 6000);
        const successEl = document.getElementById('contact-success');
        if (successEl) successEl.classList.add('visible');
      } catch (err) {
        Toast.show(err.message || 'Failed to send message. Please try again.', 'error');
      } finally {
        btn.classList.remove('loading');
        btn.disabled = false;
      }
    });
  }

  return { init };
})();
