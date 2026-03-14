# Runtime Boundaries

## Purpose

This repo now separates runtime-only infrastructure from business modules:

- `internal/modules/*`: business entrypoints and domain logic. Modules must not import `internal/service`.
- `internal/runtimeops`: shared runtime state and tuning, such as Turbo mode and sync ticker hot-reload.
- `internal/platformtools`: platform detection, generated config hints, and PHP config parsing.
- `internal/modules/admin`: admin runtime tooling entrypoints. Keep any tiny runtime glue local to the admin module instead of creating a new top-level bridge package.
- `internal/service`: compatibility exports, legacy/openapi support, and remaining runtime hosts used by bootstrap or legacy paths.

## Rules

1. New business code belongs in `internal/modules/*`, not `internal/service`.
2. If admin tooling needs runtime state, prefer `runtimeops`, `pluginruntime`, `autosync`, or `dbtools`; keep any remaining glue inside `internal/modules/admin`.
3. If a capability is pure parsing/detection logic, prefer `platformtools`.
4. Keep `service` wrappers thin when a capability has already been extracted.

## Current Status

- `internal/modules/*` no longer imports `internal/service`.
- `service/turbo.go` and `service/platform_tools.go` are compatibility wrappers over extracted packages.
- Remaining `service` usage is intentional: bootstrap jobs, legacy/openapi compatibility, and runtime host ownership.
