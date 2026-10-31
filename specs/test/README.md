# API Specs

Each `.feature` file describes how one area of the Sloth API must behave. An agent runs them against a live server with real HTTP requests and reports what passed.

## Running

Give the agent a server URL, a user API key (sent as `X-API-Key`), the files to run, and optionally a tag filter:

> Run specs/test/agents.feature against http://localhost:8000 with API key ck_xxx.

## Rules for the agent

- **Own workspace.** Create `spec-<runId>` before the first scenario and keep everything you create in it (or in workspaces a scenario asks for, also prefixed `spec-<runId>`). Never touch anything you didn't create.
- **Clean up** everything at the end, even after failures, and report what you couldn't delete.
- **Mint keys in the run.** "An access key with only X" means a new access key with exactly those permissions. The full list is `AccessKeyPermissions` in `module/permission.go`.
- **Tags.** `@readonly`, `@mutates` and `@smoke` always run. Skip `@slow`, `@destructive`, `@worker` and `@llm` unless asked. `@llm` needs the server's AI provider configured and makes real model calls. `@admin` needs a platform admin. `@saas` / `@oss` run only on that edition (see `/api/v1/public/action/setup/status`).
- **Routes** come from each feature's description and `openapi.yaml`. A path like `agents/x` is relative to `/api/v1/workspaces/{ws}/`.
- **Then steps are exact.** Status, fields and values must all match. "No X field" means the key is absent. If a step is ambiguous, mark the scenario `unclear`.
- **Scenarios are independent.** Background runs before each one.
- **Rate limits.** The token route and setup are limited per IP, so run the `@slow` rate-limit scenarios last.
- **No secrets** in the report. Use `<redacted>`.

## Report

Write to `specs/reports/<YYYY-MM-DD>-<runId>.md` (gitignored): a header with server, edition, files and tags; a pass / fail / skipped / unclear count; then each failure with its step, request, expected and actual result; then the skipped scenarios with reasons.

## Writing a spec

- One area per file, opening with what it guarantees and which routes it uses.
- Describe behaviour, not implementation. One check per scenario.
- Use `<runId>` in names and `<a random UUID>` for ids that must not exist.
- A spec states what the server **should** do. When they disagree, fix whichever is wrong.
