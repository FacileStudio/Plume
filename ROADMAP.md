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

## Short-term

- [ ] PDF viewer (render uploaded PDF pages in browser)
- [ ] Drag-and-drop field placement on PDF pages
- [ ] Signature pad (draw signature on canvas)
- [ ] Audit trail PDF generation (event log per document)
- [ ] Field types: initials, text, date, checkbox (signature exists)
- [ ] Download signed document with embedded field values

## Medium-term

- [ ] Document templates (reusable field layouts)
- [ ] Multi-signer ordering (sequential signing flow enforcement)
- [ ] Bulk send from CSV (batch create documents + signers)
- [ ] API keys for programmatic access
- [ ] Digital signatures (PKI, X.509, PKCS#7)
- [ ] RFC 3161 timestamp server support
- [ ] Custom PKCS#12 certificate upload
- [ ] Reminders (automatic re-send for pending signers)

## Long-term

- [ ] PDF/A-3b archival compliance
- [ ] AATL certificate support
- [ ] SMS verification for signers
- [ ] Conditional fields and formula fields
- [ ] Embeddable signing components (iframe / Svelte / React)
- [ ] Mobile-optimized signing experience
- [ ] Data retention policies and auto-cleanup
- [ ] Import from DocuSign / HelloSign
- [ ] Branding customization (logo, colors in emails and signing page)
