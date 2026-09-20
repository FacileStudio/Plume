# Changelog

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
