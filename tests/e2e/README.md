# ClassReach end-to-end flows

Run `npm ci --ignore-scripts`, then `npm run test:e2e` or `make test-e2e`.
Use Node 22.12 or newer and the Go version from `go.mod`.

TesterArmy `e2e` 0.17.0 runs 15 flow tests against the compiled CLI.
Each test creates a local HTTP fixture, an isolated home, and a private config.
The child process receives only explicit test settings and platform variables.
Each process has a five-second deadline. Each test has a 30-second deadline.
The suite uses no model, browser, external credentials, or production service.
Linux CI runs the suite after the native checks.

## Command coverage

| Command leaf | Observable checks |
| --- | --- |
| `config init` | Password stdin, CRLF, private modes, overwrite refusal, force, default home path, persisted login, dry-run |
| `config show` | Redaction, actual path, environment values, flag precedence |
| `login` | Form credentials, anti-forgery token, cookie session, denied credentials, challenges, missing session, missing config credentials |
| `doctor` | Authenticated quick view, missing model, malformed data |
| `overview` | Requested week, student/announcement counts, JSON and text |
| `students list/get` | IDs and names, alias, selected student, unknown student, text and JSON |
| `courses list/get` | Student filter, empty filter result, section lookup, unknown section, text and JSON |
| `grades list` | Selected student, numeric grade, text and JSON |
| `assignments list/get` | Student/section URL, escaped HTML model, assignment selection, missing assignment, malformed model |
| `attendance list` | Embedded model, records, marking labels, malformed model |
| `announcements list` | JSON, HTML removal, entity decoding, text |
| `calendar list` | Date query, event identity, text and JSON, invalid range |
| `notifications counts` | Term query, unknown provider fields, large integer bytes, non-JSON rejection |
| `directory list/families` | Default directory/year discovery, explicit IDs, paging, search, sort, academic levels, text and JSON |
| `messages list` | Cookie/token handshake, paging/filter request, thread identity, paging summary, JSON |
| `messages get` | Thread request, parsed messages, decoded body, JSON |
| `messages download` | Attachment lookup, exact binary bytes, fallback URL, missing file, force, JSON receipt |
| `documents list` | Folder query, folder/file output, JSON |
| `documents download` | Exact bytes, private file, overwrite refusal, symlink refusal, force, missing document, fallback URL, HTTP failure and truncated body without file changes |
| `agenda download` | Exact ZIP bytes, extracted PDF, private file, overwrite/force, unsafe traversal, invalid ZIP |
| `raw get` | Repeated/encoded query, exact binary stdout, JSON integer bytes, trace redaction, redirect refusal, non-JSON rejection |
| `version` and `--version` | JSON/text/plain output, subcommand flag, environment format and flag precedence |
| `completion` | Actual Bash/Zsh/Fish/PowerShell script output, invalid shell |

The suite checks every authenticated leaf and `config init` with `--dry-run`.
It asserts zero network requests and unchanged config/output files.
Invalid flags, arguments, dates, paging values, and environment values return usage errors before authentication.
HTTP status failures, refused connections, malformed responses, and deadlines return errors through stderr.

## Scope limits

Fixtures prove local process behavior against the stated response contracts.
They do not prove the current ClassReach tenant contract or actual school records.
Real login challenges, Azure origin routing, provider permissions, and external file hosts require separate authorized live checks.
Message reads can mark threads as read. This suite does not access live messages.
Native Go tests retain response-size, expanded ZIP-size, file-race, and transport routing checks.
The E2E suite supports macOS and Linux. CI runs it on Linux; Windows retains the native Go checks.
Reports appear in the ignored `.e2e/` directory. No test credentials are real.
