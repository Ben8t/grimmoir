# AGENT.md

## Development Philosophy: Red/Green TDD

Use **red/green TDD** for all code changes.

Write the automated tests first, confirm that they **fail** (red), then iterate on the implementation until the tests **pass** (green). Never skip the red phase — if you do, you risk building a test that passes already, failing to exercise and confirm the new implementation.

This protects against writing code that doesn't work, building code that is unnecessary, and ensures a robust automated test suite that guards against future regressions.

Reference: [Simon Willison — Red/Green TDD](https://simonwillison.net/guides/agentic-engineering-patterns/red-green-tdd/)

---

## Code Quality Standards

All code produced must meet the following bar:

1. **The code works.** It does what it's meant to do, without bugs.
2. **We know the code works.** We've taken steps to confirm to ourselves and to others that the code is fit for purpose.
3. **It solves the right problem.**
4. **It handles error cases gracefully and predictably.** It doesn't just consider the happy path. Errors should provide enough information to help future maintainers understand what went wrong.
5. **It's simple and minimal.** It does only what's needed, in a way that both humans and machines can understand now and maintain in the future.
6. **It's protected by tests.** The tests show that it works now and act as a regression suite to avoid it quietly breaking in the future.
7. **It's documented at an appropriate level,** and that documentation reflects the current state of the system. If the code changes an existing behavior, the existing documentation must be updated to match.
8. **The design affords future changes.** Maintain YAGNI — code with added complexity to anticipate future changes that may never come is often bad code — but also don't write code that makes future changes much harder than they should be.
9. **All relevant "-ilities"** — accessibility, testability, reliability, security, maintainability, observability, scalability, usability — the non-functional quality measures appropriate for the particular class of software being developed.

---

## MCP Agent Mail: Coordination for Multi-Agent Workflows

### What it is

A mail-like layer that lets coding agents coordinate asynchronously via MCP tools and resources. Provides identities, inbox/outbox, searchable threads, and advisory file reservations, with human-auditable artifacts in Git.

### Why it's useful

- Prevents agents from stepping on each other with explicit file reservations (leases) for files/globs.
- Keeps communication out of your token budget by storing messages in a per-project archive.
- Offers quick reads (`resource://inbox/...`, `resource://thread/...`) and macros that bundle common flows.

### How to use effectively

#### 1) Same repository

- **Register an identity:** call `ensure_project`, then `register_agent` using this repo's absolute path as `project_key`.
- **Reserve files before you edit:** `file_reservation_paths(project_key, agent_name, ["src/**"], ttl_seconds=3600, exclusive=true)` to signal intent and avoid conflict.
- **Communicate with threads:** use `send_message(..., thread_id="FEAT-123")`; check inbox with `fetch_inbox` and acknowledge with `acknowledge_message`.
- **Read fast:** `resource://inbox/{Agent}?project=<abs-path>&limit=20` or `resource://thread/{id}?project=<abs-path>&include_bodies=true`.
- **Tip:** set `AGENT_NAME` in your environment so the pre-commit guard can block commits that conflict with others' active exclusive file reservations.

#### 2) Across different repos in one project (e.g., Next.js frontend + FastAPI backend)

- **Option A (single project bus):** register both sides under the same `project_key` (shared key/path). Keep reservation patterns specific (e.g., `frontend/**` vs `backend/**`).
- **Option B (separate projects):** each repo has its own `project_key`; use `macro_contact_handshake` or `request_contact`/`respond_contact` to link agents, then message directly. Keep a shared `thread_id` (e.g., ticket key) across repos for clean summaries/audits.

### Macros vs granular tools

- **Prefer macros** when you want speed or are on a smaller model: `macro_start_session`, `macro_prepare_thread`, `macro_file_reservation_cycle`, `macro_contact_handshake`.
- **Use granular tools** when you need control: `register_agent`, `file_reservation_paths`, `send_message`, `fetch_inbox`, `acknowledge_message`.

### Common pitfalls

- **"from_agent not registered":** always `register_agent` in the correct `project_key` first.
- **"FILE_RESERVATION_CONFLICT":** adjust patterns, wait for expiry, or use a non-exclusive reservation when appropriate.
- **Auth errors:** if JWT+JWKS is enabled, include a bearer token with a `kid` that matches server JWKS; static bearer is used only when JWT is disabled.

## Check Deployment on Vercels

Use the vercel cli installed to check and deploy changes if necessary

## Local CLI Update After Changes

When code changes affect the `grim` CLI or TUI, always rebuild and reinstall the local binary so manual testing uses the latest behavior.

Use:

```bash
go test ./...
go build -o ./bin/grim ./cmd/grim
install -m 755 ./bin/grim /opt/homebrew/bin/grim
```

Then verify:

```bash
grim --help
```
