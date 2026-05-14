# Plume — Roadmap

Self-hosted document signing. Go + SvelteKit. DocuSeal alternative, single-tenant.

## Done

- [x] Document CRUD (create, list, view, delete)
- [x] PDF upload with persistent filesystem storage (named Docker volume)
- [x] PDF file serving (GET /documents/{id}/file)
- [x] Signer management (add signers with name, email, role)
- [x] Document sending (generate signing tokens, set status to pending)
- [x] Public signing page (token-based, standalone)
- [x] Signature submission with IP and user agent capture
- [x] Document status workflow (draft -> pending -> completed/declined)
- [x] SMTP configuration (per-user, with test button)
- [x] Email notifications to signers (signing invitation with link)
- [x] Dashboard with stats (total documents, pending, completed)
- [x] Profile page (name, email, change password)
- [x] Settings page (SMTP config, webhook configuration)
- [x] OIDC/SSO authentication
- [x] Docker Compose deployment (db + api + client, named volumes)
- [x] Webhook dispatching with HMAC-SHA256 signing
- [x] Minimalist black & white UI, Solar / Iconify icons
- [x] Webhook dispatch on document/signer events (sent, signed, declined, completed)
- [x] Owner notification emails (signed/declined events)
- [x] Audit trail in API (IP address + user agent on signer responses)
- [x] PDF viewer on signing page (pdfjs-dist, side-by-side layout)
- [x] Field CRUD API (create, update, delete fields on draft documents)
- [x] Drag-and-drop field placement on PDF pages (field editor with cross-page drag)
- [x] Field labels and renaming in field editor
- [x] Signer color-coded field overlays (per-signer palette)
- [x] Field overlays on signing page with scroll-to-field on input focus
- [x] Signature certificate PDF generation (lazy, on-demand download)
- [x] Delete confirmation modal (shadcn alert dialog)
- [x] Field types: signature, text, date, checkbox

## Short-term

- [ ] Signature pad (draw signature on canvas)
- [ ] Download signed document with embedded field values (stamp onto original PDF)
- [ ] Audit trail PDF generation (full event log per document)
- [ ] Document templates (reusable field layouts)
- [ ] Reminders (automatic re-send for pending signers)

## Medium-term

- [ ] Multi-signer ordering (sequential signing flow enforcement)
- [ ] Bulk send from CSV (batch create documents + signers)
- [ ] API keys for programmatic access
- [ ] Digital signatures (PKI, X.509, PKCS#7)
- [ ] RFC 3161 timestamp server support
- [ ] Custom PKCS#12 certificate upload
- [ ] Branding customization (logo, colors in emails and signing page)

## Long-term

- [ ] PDF/A-3b archival compliance
- [ ] AATL certificate support
- [ ] SMS verification for signers
- [ ] Conditional fields and formula fields
- [ ] Embeddable signing components (iframe / Svelte / React)
- [ ] Mobile-optimized signing experience
- [ ] Data retention policies and auto-cleanup
- [ ] Import from DocuSign / HelloSign
