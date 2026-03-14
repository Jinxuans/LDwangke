# Plugin Runtime Matrix

## Purpose

This file is the operating reference for plugin runtime ownership.
It answers three concrete questions:

- Which plugins still have background runtime behavior
- Where that runtime starts today
- Under what condition the runtime may leave the cron compatibility bridge

## Current Matrix

| Plugin | Owner | Runtime Entry | Current State | Exit Condition |
| --- | --- | --- | --- | --- |
| `sdxy` | `cron_bridge` | `pluginruntime/cron_bridge.go:RunSDXYCron` | Active plugin; cron still delegated to legacy runtime in `service` | Give SDXY a dedicated pluginruntime cron owner or retire the plugin |
| `w` | `cron_bridge` | `pluginruntime/cron_bridge.go:RunWCron` | Active plugin; cron still delegated to legacy runtime in `service` | Give W a dedicated pluginruntime cron owner or retire the plugin |
| `xm` | `cron_bridge` | `pluginruntime/cron_bridge.go:RunXMCron` | Active plugin; cron still delegated to legacy runtime in `service` | Give XM a dedicated pluginruntime cron owner or retire the plugin |
| `ydsj` | `cron_bridge` | `pluginruntime/cron_bridge.go:RunYDSJCron` | Active plugin; cron still delegated to legacy runtime in `service` | Give YDSJ a dedicated pluginruntime cron owner or retire the plugin |
| `yongye` | `cron_bridge` | `pluginruntime/cron_bridge.go:RunYongyeCron` | Active plugin; cron still delegated to legacy runtime in `service` | Give Yongye a dedicated pluginruntime cron owner or retire the plugin |
| `longlong` | `pluginruntime` | `pluginruntime/longlong.go` | Fully migrated out of `service` | Keep as pluginruntime-owned runtime |
| `hzw_socket` | `pluginruntime` | `pluginruntime/socket.go` | Fully migrated out of `service` | Keep as pluginruntime-owned runtime |
| `simple-thread` | `pluginruntime` | `pluginruntime/simplethread.go` | Fully migrated out of `service` | Keep as pluginruntime-owned runtime |

## Decision Rule

- If a plugin runtime is still routed through `cron_bridge.go`, treat it as compatibility-retained runtime, not as unowned code.
- Do not continue migration only for symmetry.
- Migrate a cron bridge plugin only when one of these becomes true:
  - Its cron logic gains a dedicated owner in `internal/pluginruntime`
  - Its runtime can be retired with the plugin itself

## Operational Meaning

- `cron_bridge` means "still active, but legacy runtime-hosted"
- `pluginruntime` means "owned by extracted runtime packages"
- `module` means "no separate runtime host is required"
