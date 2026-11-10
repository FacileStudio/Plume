# Erasure

When Registre erases a person, Plume anonymises the account and keeps the legal records.
A document that was sent for signature or signed is evidence. It stays as it was, with its
signers, its fields, its files and its audit trail, including every name and email printed
in it at the time.

## How a notice arrives

Registre's signed `user.deleted` push (`POST /api/registre/deletions`) and the deletion-list
poll both reach `porte/directory`'s receiver. It calls `spaces.Deletions.ErasePerson`, then
porte's `pg.Store.EraseIdentity`. A notice for a subject Plume does not know, or one already
erased, changes nothing. The provider matches with or without a trailing slash.

`ErasePerson` does three things, in this order:

1. **Freeze.** For every completed or declined document the person owns, it writes the
   audit trail PDF to `UPLOAD_DIR/audit/` if no current copy is there. The trail prints the
   owner's name and email from the `users` row, so it is written while that row still exists.
   A failure stops the erasure, and the receiver retries the notice later.
2. **Erase,** in one transaction through `porte/kit`'s `EraseAccount`, under the kit's lock.
3. **Remove files** of the erased drafts and the uploaded avatar, after the commit.

"Sent" means any status other than `draft`: `pending`, `completed` or `declined`.

## Field by field

| Table | Rows | What happens |
|---|---|---|
| `users` | the person's account | **Deleted.** No foreign key needs it: `documents.owner_id` has none, and the kept documents keep the old id. Every column goes: `email`, `name`, `locale`, `reminder_interval_days`, `avatar_url`, `avatar_source`, `avatar_upload_path`, `oidc_picture_url`, `oidc_subject`, `password_hash`, `oidc_access_token`, `oidc_refresh_token`, `oidc_token_expiry`, `profile_synced_at`, `created_at` |
| `porte_identities` | the person's | **Deleted** (cascade from `users`): provider, subject and the sealed Registre tokens |
| `porte_sessions` | the person's | **Deleted** (cascade): browser sessions and every API token, Ardoise's included |
| `porte_login_codes`, `porte_magic_links` | the person's | **Deleted** (cascade) |
| `smtp_configs` | owned by the person | **Deleted**, password included |
| `webhooks` | owned by the person | **Deleted**, secret included |
| `space_members` | the person's memberships | **Deleted** (cascade) |
| `spaces` | owned, unlinked, with no other member | **Deleted** when no document is left in it after the drafts go. **Kept** otherwise, with `owner_id` set to 0, because a sent document names it |
| `spaces` | owned, with other members or linked to Registre | **Kept**, `owner_id` passes to the longest-standing owner or admin, else the longest-standing member, else 0 |
| `clients` | owned by the person, named by no sent document | **Deleted**. A draft that named one gets `client_id` NULL |
| `clients` | owned by the person, named by a sent document | **Kept** unchanged: they are the counterparty of a legal record |
| `documents` | drafts, personal or in a space only the person belonged to | **Deleted**, with their `signers` and `fields`, the original and flattened PDFs, and any cached audit trail and certificate |
| `documents` | drafts in a space others use | **Kept**, `owner_id` passes to the space's successor as above; an `external_ref` that clashes with the successor's own is cleared |
| `documents` | sent: `pending`, `completed`, `declined` | **Kept byte-identical**, every column, `owner_id` and `updated_at` included |
| `signers`, `fields` | of a sent document | **Kept byte-identical**: names, emails, tokens, IP addresses, user agents, timestamps and signature values |
| `signers` | the person signing someone else's document | **Kept byte-identical**: it is that document's audit trail |
| `directory_deletion_state` | one row | Updated by the receiver; holds no personal data |

| File under `UPLOAD_DIR` | What happens |
|---|---|
| Original, flattened `_signed.pdf`, `audit/audit_<id>.pdf`, `certificates/cert_<id>.pdf` of a sent document | **Kept byte-identical** |
| `audit/audit_<id>.pdf` of a completed or declined document that had none | **Written** before the erasure (the freeze) |
| The same files of an erased draft | **Deleted** |
| The uploaded avatar at `avatar_upload_path` | **Deleted** |
| The owner's upload directory `<id>/` | Removed only when empty |

## What still names the person afterwards

Only the legal records: the sent documents' rows, their signers and fields, their PDFs, and
their audit trails. Nothing else in Plume's database contains the person's email, name or
Registre subject. `TestErasePersonLeavesEverySentDocumentByteIdentical` in
`apps/api/modules/spaces/erase_records_test.go` proves it: it scans every text column of every
table for the three values and snapshots the kept rows and files before and after.

## Known limits

- **A pending document of an erased owner** can still be signed through its link. Its
  regenerated audit trail then prints an empty owner, because the trail is regenerated after
  each signature and the owner's row is gone. Completed and declined documents are frozen.
- **No mail follows an erased owner's document.** Their SMTP settings are gone, so no
  reminder, completion notice or next-signer invitation is sent, and their webhooks are gone,
  so Ardoise hears nothing about it. Nuage archiving is skipped: no Registre identity is left.
- **Other people's data is not touched.** Another account's drafts or contacts that name the
  erased person stay: they belong to that account.
