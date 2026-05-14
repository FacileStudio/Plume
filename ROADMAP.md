# Plume — Roadmap

Self-hosted document signing. Go + SvelteKit.

## Done

- [x] Document CRUD (create, list, view, delete)
- [x] Signer management (add signers with name, email, role)
- [x] Document sending (generate signing tokens, set status to pending)
- [x] Public signing page (token-based, standalone)
- [x] Signature submission with IP and user agent capture
- [x] Document status workflow (draft → pending → completed/declined)
- [x] Dashboard with stats (total documents, pending, completed)
- [x] Profile page (name, email, change password)
- [x] Settings page (webhook configuration)
- [x] OIDC/SSO authentication
- [x] Docker Compose deployment
- [x] Webhook reports with HMAC-SHA256 signing
- [x] Goga font, black & white minimalist UI
- [x] Solar / Iconify icons across UI

## Short-term

- [ ] PDF viewer in template builder (render uploaded PDF pages)
- [ ] Drag-and-drop field placement on PDF pages
- [ ] Signature pad (draw signature on canvas)
- [ ] Email notifications to signers (SMTP integration)
- [ ] Audit trail PDF generation (signed separate document with event log)
- [ ] Field types: signature, initials, text, date, checkbox

## Medium-term

- [ ] Digital signatures (PKI, X.509 certificates, PKCS#7)
- [ ] RFC 3161 timestamp server support
- [ ] Custom PKCS#12 certificate upload
- [ ] Multi-signer ordering (sequential signing flow)
- [ ] Document templates (reusable field layouts)
- [ ] Bulk send from CSV
- [ ] API keys for programmatic access

## Long-term

- [ ] PDF/A-3b compliance
- [ ] AATL certificate support
- [ ] SMS/email verification for signers
- [ ] Conditional fields and formula fields
- [ ] Embeddable signing components (Svelte, React, Vue)
- [ ] Mobile-optimized signing experience
- [ ] Data retention policies
- [ ] Import from DocuSign / HelloSign
