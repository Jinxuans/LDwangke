# Internal Layout

## Purpose

`internal/` is split by role, not by historical layer count.
Use this file as the short directory map for development.

## Main Areas

- `modules/`
  - Main business modules and plugin modules.
- `shared/`
  - Cross-module shared helpers and repositories.
- `routes/`
  - Route registration only.
- `bootstrap/`
  - Startup jobs and process wiring.
- `app/`
  - Application assembly and shutdown hooks.
- `legacy/`
  - Compatibility-only protocol surfaces.
- `pluginruntime/`
  - Plugin-side runtime hosts.
- `autosync/`
  - Main-system auto-sync runtime.
- `dbtools/`
  - DB compatibility and external DB sync tooling.
- `runtimeops/`
  - Runtime tuning and performance profiles.
- `platformtools/`
  - Platform detection and PHP config parsing.
- `service/`
  - Legacy compatibility host package. Not for new business logic.
  - See `service/README.md` for the remaining host categories.
  - File prefixes now follow `compat_`, `plugin_host_`, and `shared_` to make the leftovers readable.

## Current Structural Rule

- New business logic goes into `modules/*`.
- New plugin runtime logic goes into `pluginruntime/*` or a narrower owner package.
- `service/` is no longer a general-purpose destination.
