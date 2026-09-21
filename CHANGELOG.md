# Changelog

## Unreleased

## v0.4.0 - 2026-09-20

- Reject cross-origin authentication redirects, HTTPS downgrades, and unsafe base URLs.
- Require an authenticated session and check the guardian API in `doctor`.
- Write private files atomically and reject destination symlinks or unwanted overwrites.
- Validate command input before login and make dry runs perform no network requests or file changes.
- Fix version flags, usage exit codes, effective config diagnostics, and JSON download output.
- Preserve large numbers in raw JSON and reject non-JSON responses in explicit JSON mode.
- Limit response and agenda archive sizes.
- Add notification counts for an explicit academic term.
- Document current endpoint coverage, discovery gaps, and native calendar subscriptions.

## 0.2.0 - 2026-08-19

- Add weekly agenda downloads as extracted PDF files or a raw ZIP archive.
- Preserve exact response bytes from `raw get` for binary endpoints.

## 0.1.2 - 2026-08-19

- Fix message attachment downloads when ClassReach returns `FileDownloadLink`.

## 0.1.0 - 2026-08-19

- Add direct ClassReach web authentication through the tenant Azure origin.
- Add guardian overview, student, course, assignment, grade, and attendance reads.
- Add message, document, announcement, calendar, and directory reads.
- Add explicit message attachment and school document downloads.
- Add stable JSON output and the ClassReach agent skill.
