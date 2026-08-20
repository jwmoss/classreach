---
name: classreach
description: "Use when Jonathan or Amanda asks 'what did the school send,' 'show my ClassReach messages,' 'what is on the school calendar,' 'find a school document,' or asks about Providence students, courses, assignments, grades, attendance, announcements, directories, and guardian data. Runs the read-only classreach CLI with safe agent output."
---

# classreach

**UTILITY SKILL:** one ClassReach read.

**INVOKES:** `classreach`.

**USE FOR:** Providence guardian, academic, message, document, calendar, and directory data.

**DO NOT USE FOR:** writes or non-ClassReach research.

Prefer typed commands with `--json --no-input`.

## Fast path

```bash
classreach login
classreach doctor --json
classreach --json --no-input overview
```

## Rules

- Keep secrets and private data out of output and Git commits.
- Resolve IDs with lists. Download identified files to explicit paths.
- Use `raw get` only when no typed command exists.
- `messages get` marks an unread thread as read.

## Examples

```bash
classreach --json --no-input students list
classreach --json --no-input courses list --student <student-id>
classreach --json --no-input assignments list --student <student-id> --section <section-id>
classreach --json --no-input grades list --student <student-id>
classreach --json --no-input attendance list --student <student-id> --section <section-id>
classreach --json --no-input messages list
classreach --json --no-input messages get <thread-id>
classreach --json --no-input documents list --folder <folder-id>
classreach --json --no-input calendar list --start YYYY-MM-DD --end YYYY-MM-DD
classreach --json --no-input directory families --search <name>
```

Download a known file:

```bash
classreach --no-input messages download <thread-id> <file-id> --output <path>
classreach --no-input documents download <document-id> --folder <folder-id> --output <path>
```

## Troubleshooting

For login failures, update the config. Never request a password in chat.
Verify filters for missing data. Use only known GET endpoints with `raw get`.

## Done when

The command succeeds with JSON and returns only the requested data.
