# Changelog

## 0.2.0 - 2026-10-30

### Added

- Completed documents can be archived in Nuage. When the last signer finishes, the stamped PDF
  is published in the background to the Nuage space that shares the Plume space's `facile_id`
  (Nuage creates it on first use), or to a default space for documents with no space. It is
  off unless `NUAGE_URL` and `NUAGE_API_TOKEN` are set, retries up to five times with backoff,
  runs at most four jobs at once, and never blocks or fails signing. The publish carries
  `plume:document:<id>`, so a retry never stores a second copy.
- Spaces have a `facile_id`, the suite-wide key. Existing spaces get one at migration, and a
  unique index enforces it.

## 0.1.0 — 2026-09-20

### Changed

- filet.yml harmonised to the strict suite standard, and the API split to meet
  it: per-route registrars, per-route documentation tables, the documents
  upload machinery in `upload.go`, the spaces member routes in `members.go`,
  and the pdfutil field painters in `stamp.go`
- dev compose: new `docker-compose.dev.yml` overlay publishes the database
  port on the loopback interface only

### Fixed

- pdfutil: restored the `FlattenFields` doc comment mangled by a merge, and
  the overlay temp file is now removed on every error path
- urlsafe: the SSRF guard no longer re-parses its CIDR table per probe; the
  table is covered by a regression test
- auth: the documentation route table appends in the non-nested form
