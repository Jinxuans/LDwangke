# Database Migration Policy

## Purpose

This project keeps two database artifacts on purpose:

- `deploy/init_db.sql`
  - Full schema snapshot for provisioning a brand-new empty database.
- `migrations/core/*.sql`
  - Incremental schema/history patches for upgrading an existing database.

They are not interchangeable.

## Operational Rule

### 1. New empty database

Use `deploy/init_db.sql` to initialize the database to the current baseline.

After that, application startup may still run newer incremental migrations that were added after the snapshot was last refreshed.

### 2. Existing database upgrade

Do not re-import `deploy/init_db.sql`.

Upgrade through `migrations/core/*.sql` only.

The Go service now runs these migrations automatically on startup by default and records applied files in `qingka_schema_migration`.

### 3. Developer workflow

When changing database structure:

1. Add a new numbered migration file under `migrations/core/`.
2. Keep the migration idempotent when practical.
3. After the migration is finalized, sync the same end-state back into `deploy/init_db.sql`.
4. Do not edit old migration files unless the file has never been applied anywhere important.

## Naming Rule

Use monotonic numbered files such as:

- `048_add_supplier_indexes.sql`
- `049_order_status_cleanup.sql`

The startup runner only picks files that begin with a numeric prefix and end with `.sql`.

## Auto Migration

Startup auto migration is controlled by:

- `bootstrap.auto_migrate`
- `bootstrap.migrations_dir`

Environment overrides:

- `GO_API_BOOTSTRAP_AUTO_MIGRATE`
- `GO_API_BOOTSTRAP_MIGRATIONS_DIR`

## Compatibility Note

To avoid replaying old historical seed/fix scripts into long-running production databases, the first automatic migration run treats existing business databases as already baselined through migration `046`, then only auto-applies newer migrations from that point forward.

If the database is missing key tables introduced by later milestones such as `qingka_dynamic_module`, `qingka_platform_config`, `qingka_wangke_sync_config`, `qingka_tenant`, `qingka_ext_menu`, or `qingka_wangke_yfdk_projects`, the auto-baseline is lowered to the last safe version before those tables were introduced so the missing schema can be backfilled automatically.

This means:

- Old live databases are protected from historical reseeding.
- New migrations like `047_*`, `048_*` continue to auto-apply normally.
- Partially upgraded legacy databases can still backfill missing milestone tables before later `ALTER TABLE` migrations run.
