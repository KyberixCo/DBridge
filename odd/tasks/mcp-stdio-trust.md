# MCP STDIO trust and reliability

Locator: `odd/tasks/mcp-stdio-trust.md`.

## Objective and authorization

Make dbridge's MCP server more useful, safe, and reliable in VS Code, as requested by the user. Scope: MCP tool metadata and guidance, advertised resource support, query limits, and error handling. VS Code terminal approvals for Python run by the agent are controlled by VS Code and remain outside server control.

## Problem and evidence

- `tools/list` has no behavior annotations, so VS Code asks to approve even the clearly read-only tools.
- The server advertises resources but does not implement `resources/read`.
- `max_rows <= 0` can cause the Oracle layer to use 500 rows, exceeding a smaller configured policy limit; switching connections discards persistence errors; PL/SQL error formatting can dereference a nil result.
- `oracle_query` can allow writes under custom/full policy and Oracle SELECT can have side effects. It must not be declared unconditionally read-only.

## Constraints and delivery

- Preserve the configured SQL and PL/SQL security policy; do not claim that annotations enforce safety.
- Keep transport compatible with VS Code and avoid relying on Python or shell wrappers.
- TDD: off; source: no explicit project or session TDD setting found. Functional runner: `go test ./...`.
- RDD: disabled by explicit global setting (`gentle-ai review mode status`); ordinary verification applies.
- Delivery strategy: ask-on-risk. Forecast: about 300 authored changed lines, below the approximate 400-line chain threshold. Running count: about 185 after MCP-1 (source, test, and documentation lines; task tracker excluded).
- Work-unit commits use Conventional Commit messages on `feat/mcp-stdio-trust`; no push or PR is authorized.

## Tasks

- [x] **MCP-1** Publish accurate read-only tool annotations and server guidance, and make advertised resources readable or stop advertising them. Route: delegated direct. Evidence: mapping required 4+ files; implementation spans server, tests, and docs. Outcome: three read-only tools annotated; 2025-03-26 negotiation supports annotations while retaining 2024-11-05 clients; server instructions added; resources/read returns schema metadata only for advertised URIs; CodeGraph index excluded from Git. Checks: `go test ./...` passed outside sandbox (macOS keychain and Go cache require it); `git diff --check` passed. Commit: `41ef0c4` (`feat(mcp): describe safe tools and serve schema resources`). RDD: disabled/unmanaged.
- [ ] **MCP-2** Harden tool arguments and failure paths without loosening DB permissions. Route: delegated direct. Evidence: implementation spans server, security, tests, and docs. Acceptance: nonpositive row counts cannot exceed policy; switching surfaces persistence errors; PL/SQL errors cannot panic on nil results; descriptions match actual limits; read-only SQL rejects multiple statements; focused tests and `go test ./...` pass. Commit: pending. RDD: disabled/unmanaged.

## Progress and next step

MCP-1 committed and verified. Delegate MCP-2.
