# Service Hosts

`internal/service` is no longer the main business layer.
It remains as a compatibility-host package until the last retained bridges are retired.

## What Still Belongs Here

- Compatibility receiver hosts used by sibling files in the same package.
  - Example: `compat_hosts.go`, `admin.go`, `email.go`, `auxiliary.go`
- Plugin compatibility hosts that still back retained cron/runtime bridges.
  - Example: `sdxy_*.go`, `w_*.go`, `xm_*.go`, `ydsj_*.go`, `yongye_*.go`
- Leftover compatibility helpers that are still shared by retained hosts.
  - Example: `helpers.go`, `http_helpers.go`, `site_config.go`

## What Does Not Belong Here

- New business logic
- New module handlers or module services
- New plugin runtime implementations
- New shared helpers unless they are strictly required by retained compatibility hosts

## Current Meaning of This Directory

- `modules/*` is the business layer
- `pluginruntime/*`, `autosync/*`, `dbtools/*`, `runtimeops/*`, `platformtools/*` own new runtime/tooling code
- `legacy/*` owns protocol compatibility
- `service/*` is a shrinking compatibility-host package

## Practical Grouping

- Core compatibility hosts
  - `compat_*.go`, `auth_*.go`, `tenant*.go`, `user_center_*.go`
- Plugin legacy hosts
  - `plugin_host_*.go`, plus related plugin implementation files such as `sdxy_*.go`, `w_*.go`, `xm_*.go`, `ydsj_*.go`, `yfdk_*.go`, `yongye_*.go`
- Shared leftovers
  - `shared_*.go`

## Naming Rule

- `compat_*.go`
  - Thin compatibility hosts and entry wrappers for non-plugin legacy surfaces.
- `plugin_host_*.go`
  - Thin plugin legacy host files that mainly define retained plugin host types or singletons.
- `shared_*.go`
  - Leftover helpers still shared by retained compatibility hosts.

## Exit Rule

Files should leave this directory only when their last external compatibility caller
has been moved to a narrower owner package or retired.
