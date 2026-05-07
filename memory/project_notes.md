---
name: Open in Mail button issue
description: The "Open in Mail" mailto button triggers a macOS system popup that the user cannot interact with
type: project
---

The dashboard reply feature has two buttons: "Send Reply" (SMTP) and "Open in Mail" (mailto link).

Clicking "Open in Mail" triggers a macOS system popup asking the user to select apps (Mail, Contacts, Calendar, Notes) to use with their Google account. The user is unable to interact with this popup.

**Why:** macOS is trying to set up a Google account with native apps — this is a system-level issue, not related to our code.

**Current status:** Unresolved. User has not confirmed a fix yet.

**How to apply:** Do not suggest using the "Open in Mail" button until the user confirms it works. Always direct the user to use the "Send Reply" button (SMTP-based) in the dashboard for replying to messages. The SMTP reply requires the Gmail App Password to be set in backend/.env — user has not yet confirmed this is done.
