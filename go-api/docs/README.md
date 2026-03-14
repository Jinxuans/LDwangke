# Docs Index

## Purpose

This directory is organized by concern so project structure and governance
documents are no longer mixed together in one flat list.

## Layout

- `architecture/`
  - Refactor boundaries, runtime ownership, legacy surface, and service-host decisions.
- `plugins/`
  - Plugin catalog, runtime ownership matrix, and plugin-side governance.
- `database/`
  - Database schema references.
- `operations/`
  - Operational notes and optimization checklists.
- `archive/`
  - Historical migration and compatibility analysis kept only for reference.

## Start Here

- `architecture/MODULE_DOMAIN_BOUNDARIES.md`
  - Core business vs plugin system split.
- `architecture/COMPATIBILITY_BOUNDARIES.md`
  - Where `service` and `legacy` are still intentionally retained.
- `architecture/SERVICE_HOST_ROADMAP.md`
  - Final service-host roadmap and stop conditions.
- `plugins/PLUGIN_SURFACE.md`
  - Active plugin catalog and runtime ownership.
- `plugins/PLUGIN_RUNTIME_MATRIX.md`
  - Runtime entrypoints and retirement conditions for plugin runtime.

## High-Value References

- `database/DB-SCHEMA.md`
- `operations/OPTIMIZATION.md`
- `architecture/LEGACY_SURFACE.md`
- `architecture/RUNTIME_BOUNDARIES.md`
