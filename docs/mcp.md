# MCP server

Plume serves a remote MCP server at `POST /mcp` (tronc/mcp) when `MCP_RESOURCE` is set. It takes
Registre access tokens bound to that resource, with the scopes `plume.read` and `plume.write`. A
person's first call creates their account, as a first web login does. Tool names and schemas are a
contract with MCP clients: add fields, never rename or reshape one.

## Tools

| Tool | Scope | Confirmed | What it does |
|---|---|---|---|
| `plume_list_documents` | read | no | A scope's documents, 25 a page, with counts and spaces |
| `plume_get_document` | read | no | One document with its client, signers and fields |
| `plume_list_clients` | read | no | The person's clients, 50 a page |
| `plume_create_document` | write | no | Upload a PDF (3 MB at most) as a draft with 1 to 20 signers |
| `plume_update_document` | write | no | Rename a draft, change its order or its client |
| `plume_add_signer` | write | no | Add a signer, approver or viewer to a draft |
| `plume_remove_signer` | write | no | Take a signer and their fields off a draft |
| `plume_save_client` | write | no | Create a client or change some of its fields |
| `plume_send_document` | write | yes | Email the signers of a draft |
| `plume_remind_signer` | write | yes | Email a signer their link again |
| `plume_delete_document` | write | yes | Delete a document, its signers, fields and file |
| `plume_delete_client` | write | yes | Delete a client; its documents stay, with no client |
| `plume_delete_field` | write | yes | Delete one field from a draft |

Every tool calls the same service method as the matching REST route, so spaces and ownership hold
as they do in the web app. A confirmed tool asks the person through MRTR when the client can ask.
Otherwise it answers `pending_confirmation` with a `confirm_url` to `/pending`, where only the
person's own browser session can confirm (porte `mcpauth.RequireOwnSession`). Signing tokens and
signing links never leave through MCP.

## Left out on purpose

| Not exposed | Why |
|---|---|
| Voiding or cancelling a sent document | Plume has no void: no REST route and no web action. A sent document can be deleted with `plume_delete_document`, which is confirmed. A void tool waits for the product feature. |
| Creating, moving or resizing fields | Placing a field is visual work on the PDF page. Drafts get a default signature box per signer, and the returned `url` opens the field editor. `plume_delete_field` covers removing a box. |
| Signing, declining, the signing file and status | These are the signer's routes, authorised by the signer's private token. An agent acting for the owner must never sign for someone. |
| SMTP settings and webhooks | They hold credentials (SMTP password, webhook secret) and decide where Plume sends mail and events. Change them in Settings. |
| API tokens | A token minted through MCP would outlive the MCP grant. Mint them in Settings > API. |
| Spaces: create, rename, leave, delete, members | Registre owns spaces when OIDC is configured, and Plume answers 409 on linked spaces. `plume_list_documents` lists the person's spaces. |
| Certificate and audit-trail PDFs | Binary files. `plume_get_document` returns the document's `url` in Plume, where both download. |
| Verifying a signed PDF | Public, anonymous and rate-limited at `/verify`; it needs no account. |
