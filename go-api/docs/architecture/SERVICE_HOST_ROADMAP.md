# Service Host Roadmap

## Purpose

`internal/service` is no longer the business center.
What remains are a few narrow host packages that still bridge into `service`.

This file defines whether each host should keep shrinking, how far it is worth going, and what should explicitly stop.

## Current Approved Hosts

### `internal/pluginruntime/cron_bridge.go`

Current role:

- Plugin-only compatibility bridge.
- Hosts the remaining plugin cron jobs.

Current `service` bridge surface:

- `RunYDSJCron`
- `RunWCron`
- `RunXMCron`
- `RunYongyeCron`
- `RunSDXYCron`

Migration target:

- Split by plugin-runtime concern, not by arbitrary file count.
- Recommended next extraction units:
  - `internal/pluginruntime/cron`

Stop condition:

- Stop once plugin runtime is isolated from business code and no longer leaks into `bootstrap` or `app`.
- Do not force-delete `service` implementations if they are still the only runtime host for a plugin and there is no operational gain.
- Treat `sdxy / w / xm / ydsj / yongye` as explicit cron-bridge owned plugins until each one has a dedicated runtime owner or is retired.

### `internal/autosync`

Current role:

- Auto-sync runtime state/config host for the main system.

Current status:

- Completed extraction.
- `internal/autosync` now owns config persistence, runtime state, and sync execution needed by auto-sync scheduling.
- It no longer imports `internal/service`.

Stop condition:

- Keep `internal/autosync` stable as the owner of main-system auto-sync behavior.
- Do not move this logic back into `service` or into plugin runtime packages.

### `internal/dbtools/compat.go`

Current role:

- DB compatibility inspection/fix owner.

Current status:

- Completed extraction.
- `internal/dbtools` now owns schema definitions, schema inspection/fix execution, email template seeding, and default-admin compatibility repair.
- It no longer imports `internal/service`.

Stop condition:

- Keep `dbtools` as the owner of DB compatibility inspection/fix.
- Do not move this operational tooling back into `service`.

### `internal/dbtools/sync.go`

Current role:

- External DB sync owner.

Current status:

- Completed extraction.
- `internal/dbtools` now owns external DB connection, schema probing, auto-add-column handling, and row sync execution.
- It no longer imports `internal/service`.

Stop condition:

- Keep `dbtools` as the owner of external DB sync tooling.
- Do not move this operational toolchain back into `service`.

## Priority Order

1. `internal/pluginruntime/cron_bridge.go`

Reasoning:

- `dbtools/compat` is complete and should now remain stable.
- `dbtools/sync` is complete and should now remain stable.
- `autosync` is already complete and should now be treated as a stable owner.
- `pluginruntime/cron_bridge` is now the only remaining non-test host still bridging into `service`; it is tied to legacy plugin lifecycle behavior, so it should be split carefully, not rushed.

## Explicit Non-Goals

- Do not keep moving code out of `service` just to make the directory smaller.
- Do not split plugin modules further unless the split improves runtime ownership or retirement planning.
- Do not move legacy compatibility code into plugin modules.

## Completion Signal

This line of work is effectively complete when:

- Core business stays out of `service`
- Plugin runtime stays out of core business packages
- Remaining `service` imports are only in approved host packages
- Each remaining host has an explicit owner and stop condition
