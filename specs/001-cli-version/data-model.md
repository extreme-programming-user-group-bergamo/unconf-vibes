# Phase 1: Data Model

**Feature**: 001-cli-version

## Entities

### ApplicationMetadata

*Internal configuration struct, not persisted.*

| Field | Type | Description |
|-------|------|-------------|
| `Version` | `string` | Semantic version (e.g., "0.1.0"). Injected via ldflags. |
| `Commit` | `string` | Git commit hash. Injected via ldflags. |
| `Date` | `string` | Build date. Injected via ldflags. |

## Storage Schema

**None.** This feature does not require database persistence.
