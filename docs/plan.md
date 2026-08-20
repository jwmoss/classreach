# ClassReach CLI implementation plan

## Goal

Build `classreach`, a read-only Go CLI for the ClassReach private API. Jonathan and an OpenClaw
agent will use it instead of the Providence Wilmington ClassReach website.

The first release supports a guardian account and all children visible to that account. It uses
human-readable output by default and stable JSON for agent calls.

## Fixed decisions

- Generate the project from `jwmoss/restctl-template`.
- Use `classreach` as the binary name.
- Configure the ClassReach tenant URL, username, and password in the existing YAML config.
- Use browser-harness only during private API discovery.
- Make runtime requests directly from the Go HTTP client.
- Do not require Chrome or browser-harness at runtime.
- Keep version `0.1.0` read-only, except for requests required to authenticate.
- Do not cache ClassReach response data.
- Download files only after an explicit command.
- Support the Providence Wilmington tenant first while keeping the base URL configurable.
- Include a repository-local ClassReach agent skill for OpenClaw use.

## Phase 1: Discover the private API

Use browser-harness while Jonathan manually signs in and opens each guardian screen. Record the
network requests for:

- authentication and CSRF setup
- guardian identity and visible students
- courses
- messages and attachments
- documents and downloads
- assignments
- grades
- calendar events
- attendance
- announcements
- directory entries
- guardian dashboard data

For each request, record:

- the HTTP method and path
- required headers and cookies
- query parameters
- pagination behavior
- request and response shapes
- file download behavior
- read-state side effects

Store the resulting endpoint matrix in `docs/api-discovery.md`. Use sanitized examples only. Do
not commit HAR files, passwords, cookies, tokens, or real student data.

### Direct-access proof gate

Before resource implementation, prove that a plain Go HTTP client can:

1. Authenticate with the configured username and password.
2. maintain the required cookie and CSRF state.
3. retrieve one guardian resource without Chrome.

Stop if Cloudflare requires a browser at runtime. That result conflicts with the fixed runtime
boundary.

## Phase 2: Implement authentication

Replace the generated token-header flow with the discovered ClassReach login flow.

Use this configuration shape:

```yaml
base_url: https://providencewilmington.classreach.com
origin_host: classreach.azurewebsites.net
username: your-username
password: your-password
```

Keep the template's `0600` config file permissions. Redact the username and password from config,
doctor, trace, and error output.

Use Go HTTP and cookie support. Do not add a browser integration or session cache.

Provide these checks:

```text
classreach login
classreach doctor --json
```

`login` validates the configured credentials. `doctor` validates the configuration,
authentication flow, and basic API access.

## Phase 3: Implement read commands

```text
classreach overview

classreach students list
classreach students get <id>

classreach courses list
classreach courses get <id>

classreach messages list
classreach messages get <id>
classreach messages download <attachment-id>

classreach documents list
classreach documents download <id>

classreach assignments list
classreach assignments get <id>

classreach grades list
classreach calendar list
classreach attendance list

classreach announcements list
classreach announcements get <id>

classreach directory list
classreach directory get <id>

classreach raw get <path>
```

Add only filters that the discovered API supports. Expected filters include:

```text
--student <id>
--course <id>
--term <id>
--since <date>
--until <date>
--unread
--all
```

Keep `raw` limited to GET requests in version `0.1.0`.

## Output and content rules

Preserve the template contracts:

- Human-readable output is the default.
- `--json` provides stable machine output.
- `--plain`, `--no-input`, `--timeout`, and `--trace-http` remain available.
- Primary data goes to stdout.
- Errors and diagnostics go to stderr.
- HTML messages and announcements become readable text in human output.
- JSON keeps useful source fields when the API provides them.
- List and get commands do not download files.
- Message retrieval avoids changing unread state when the API permits it.

## Phase 4: Add the agent skill

Create:

```text
.agents/skills/classreach/SKILL.md
```

Model it on the SkyCLI repository skill. Include:

- a fast path
- configuration and login
- guardian and student selection
- common reads
- messages and attachments
- documents and downloads
- assignments and grades
- calendar and attendance
- announcements and directory
- raw endpoint access
- privacy rules
- troubleshooting
- repository paths

The skill directs agents to use commands such as:

```bash
classreach login
classreach doctor --json
classreach --json --no-input overview
classreach --json --no-input messages list
classreach --json --no-input documents list
classreach --json --no-input assignments list
classreach --json --no-input grades list
```

The skill must enforce these rules:

- Keep credentials and student data out of logs and summaries that do not need them.
- Use explicit download paths.
- Use only read commands in version `0.1.0`.
- Use `raw get` only when no typed command exists.
- Verify the configured tenant and student before reporting results.
- Keep real API responses and downloaded files out of commits.

Add OpenClaw installation instructions to the README.

## Phase 5: Test and verify

Use `httptest.Server` and sanitized fixtures to test:

- login and CSRF handling
- cookies within one process
- authentication failures
- credential redaction
- list and get decoding
- pagination
- message content conversion
- document downloads
- JSON output
- read-only enforcement

Run:

```bash
make check
./skill-checker/scripts/check-skill.sh .agents/skills/classreach
```

Run a manual live smoke test against the Providence tenant after local tests pass.

## Version 0.1.0 completion criteria

Version `0.1.0` is complete when:

- The binary works without Chrome or browser-harness.
- Direct login works with the configured credentials.
- Each listed guardian resource has a working read command.
- Messages and documents support explicit downloads.
- OpenClaw receives stable JSON.
- Commands do not cache student data or perform intentional resource mutations.
- The ClassReach skill passes the repository skill checks.
- `make check` passes without warnings.
