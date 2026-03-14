# Compatibility Boundaries

## Purpose

Structural refactor is complete. `internal/modules/*` no longer imports `internal/service`.
This document defines where `service` and `legacy` usage is still intentional, and where new usage is no longer allowed.

## Approved `service` hosts

Only the following locations may directly import `go-api/internal/service`:

- `internal/pluginruntime/cron_bridge.go`
  - Plugin runtime host for plugin cron compatibility bridges.
  - `autosync` now owns its own config/state/runtime implementation and no longer imports `service`.
  - `dbtools/compat` now owns its own schema inspection/fix implementation and no longer imports `service`.
  - `dbtools/sync` now owns its own external DB sync implementation and no longer imports `service`.
  - `pluginruntime` itself now owns HZW socket, LongLong runtime, and simple-thread runtime implementation; only the plugin cron bridge still imports `service`.

## Forbidden new usage

The following areas must not add new direct `service` imports:

- `internal/modules/*`
- New business handlers
- New repositories or shared domain helpers
- New route packages outside the existing compatibility hosts

If a new business feature needs logic that currently lives in `service`, extract it into:

- `internal/modules/*` for domain logic
- `internal/runtimeops` for runtime state/tuning
- `internal/platformtools` for parsing/detection
- `internal/autosync` for auto-sync runtime state/config hosts
- `internal/dbtools` for DB compatibility and external DB sync tooling
- `internal/pluginruntime` for plugin-system runtime hosts

## `legacy` boundary

`internal/legacy/*` remains intentionally active for protocol compatibility:

- `internal/legacy/openapi`
  - Keeps old OpenAPI-compatible request/response shapes.
- `internal/legacy/php`
  - Keeps PHP bridge and reverse proxy behavior.
- `internal/legacy/module`
  - Keeps legacy module-style routing compatibility.

These packages are compatibility surfaces, not the place for new product logic.

## Current decision

- Keep `legacy/*` and the approved `service` hosts stable.
- Keep `service` imports out of `app`, `bootstrap`, `routes`, `legacy`, and normal business modules; they now go through explicit owner packages.
- Do not continue broad “move code out of service” work unless a concrete runtime host is being extracted.
- Prefer tests and explicit boundaries over further structural churn.

Related roadmap:

- `docs/architecture/SERVICE_HOST_ROADMAP.md`
  - Defines what remains in each approved host, whether it should keep shrinking, and where to stop.
