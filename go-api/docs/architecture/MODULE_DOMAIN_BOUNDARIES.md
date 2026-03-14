# Module Domain Boundaries

## Purpose

This project has two different kinds of code under `internal/modules`:

- Core business modules
  - The main business is the normal web-course ordering system and its surrounding account, pricing, tenant, and admin capabilities.
- Plugin modules
  - The rest of the product-specific feature sets are treated as dedicated plugin systems.

This document fixes that distinction so future work does not keep mixing the two.

## Core Business Modules

These modules belong to the main business domain:

- `admin`
- `agent`
- `auth`
- `auxiliary`
- `chat`
- `checkin`
- `class`
- `mail`
- `order`
- `push`
- `supplier`
- `tenant`
- `user`

## Plugin System Modules

These modules are plugin domains, not the core web-course business:

- `appui`
- `paper`
- `sdxy`
- `sxdk`
- `tuboshu`
- `tutuqg`
- `tuzhi`
- `w`
- `xm`
- `ydsj`
- `yfdk`
- `yongye`

See also:

- `docs/plugins/PLUGIN_SURFACE.md`
  - Operational catalog of plugin modules and plugin runtime ownership.

## Shared Module

- `common`
  - Shared helper surface for cross-module reuse.
  - Not a standalone business domain.

## Structural Rules

- New main-business features should go into core modules or new core modules.
- New plugin products should not be added to core modules.
- Plugin runtime hosts belong in `internal/pluginruntime`.
- Core runtime/state hosts belong in narrower purpose-built packages such as `internal/autosync`, `internal/dbtools`, or future core-specific runtime packages.
- Any new directory under `internal/modules` must be classified as `core`, `plugin`, or `shared`.

## Enforcement

- `internal/boundaries/module_domain_test.go`
  - Fails if a new module directory is added without classification.
- `internal/boundaries/service_usage_test.go`
  - Prevents `service` from spreading back into business modules.
