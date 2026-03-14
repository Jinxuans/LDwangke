# Plugin Surface

## Purpose

The normal web-course ordering flow is the core business.
The product-specific systems below are treated as plugin domains.

This file is the operational catalog for those plugin modules.

## Plugin Catalog

| Module | Status | Runtime | Owner | Notes |
| --- | --- | --- | --- | --- |
| `appui` | Active | No | `module` | AppUI product plugin |
| `paper` | Active | No | `module` | Paper generation plugin |
| `sdxy` | Active | Yes | `cron_bridge` | SDXY plugin with background sync runtime |
| `sxdk` | Active | No | `module` | SXDK plugin |
| `tuboshu` | Active | No | `module` | Tuboshu writing plugin |
| `tutuqg` | Active | No | `module` | TutuQG plugin |
| `tuzhi` | Active | No | `module` | TuZhi plugin |
| `w` | Active | Yes | `cron_bridge` | W plugin with background sync runtime |
| `xm` | Active | Yes | `cron_bridge` | XM plugin with background sync runtime |
| `ydsj` | Active | Yes | `cron_bridge` | YDSJ plugin with background sync runtime |
| `yfdk` | Active | No | `module` | YFDK plugin |
| `yongye` | Active | Yes | `cron_bridge` | Yongye plugin with background sync runtime |

## Status Semantics

- `active`
  - Plugin module is still registered in current routes and remains part of the live plugin surface.
- `compat_retained`
  - Plugin module is kept only for compatibility or migration reasons and should be a retirement candidate.

Current repository state:

- All plugin modules listed above are still `active`.
- Compatibility-retained surfaces currently live mainly under `internal/legacy/*`, not under `internal/modules/*`.

## Runtime Boundary

Plugin runtime entrypoints currently live in:

- `internal/pluginruntime/cron_bridge.go`
- `internal/pluginruntime/socket.go`
- `internal/pluginruntime/longlong.go`
- `internal/pluginruntime/simplethread.go`

Current split:

- `cron_bridge.go`
  - Compatibility bridge only.
  - Still delegates `sdxy / w / xm / ydsj / yongye` cron loops to legacy plugin runtime in `service`.
- `socket.go`
  - Owns HZW socket runtime.
- `longlong.go`
  - Owns LongLong runtime and CLI tooling.
- `simplethread.go`
  - Owns simple-thread sync runtime.

The following plugins are intentionally still on the cron compatibility bridge:

- `sdxy`
- `w`
- `xm`
- `ydsj`
- `yongye`

Detailed runtime ownership matrix:

- `docs/plugins/PLUGIN_RUNTIME_MATRIX.md`

Reason:

- These loops are still tightly coupled to legacy plugin runtime implementations.
- They are active plugin surfaces, not abandoned dead code.
- Further migration should only happen if a specific plugin runtime gains a clear operational owner.

Retirement / migration rule by owner:

- `module`
  - Default owner. Retire only after routes and product obligations are removed.
- `cron_bridge`
  - Special compatibility owner. Move off the bridge only when that plugin's cron has a dedicated `pluginruntime` owner or the whole plugin is retired.

Current protection:

- `internal/boundaries/plugin_cron_bridge_test.go`
  - Locks which plugins are allowed to remain on the cron bridge.
- `internal/pluginruntime/cron_bridge_test.go`
  - Locks the pluginruntime-to-service cron delegation layer.

## Governance

- New plugin modules must be added to `internal/modulemeta/domains.go`.
- New plugin modules must also be added to `internal/modulemeta/plugins.go`.
- If a plugin exposes HTTP routes, its `PluginSpec` must declare `HasRoutes=true`.
- If a plugin owns background jobs or plugin-side runtime tooling, its `PluginSpec` must declare `HasPluginRuntime=true`.
- Every plugin must declare `Status` as either `active` or `compat_retained`.
- Every plugin must declare `RuntimeOwner` as either `module` or `cron_bridge`.
- Every plugin must declare a non-empty `RetirementRule`.
- If a plugin still depends on `pluginruntime/cron_bridge.go`, it must be listed in `modulemeta.CronBridgePlugins`.
