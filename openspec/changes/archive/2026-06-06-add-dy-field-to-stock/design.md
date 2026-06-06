## Context

The API uses a clean architecture (domain → application → infrastructure/adapters). The `Stock` struct is the single source of truth; GORM handles persistence via auto-migrate. Adding a new field requires propagating it through every layer: domain, DTO, handler, service, and repository (auto-migrate handles the DB column).

Current fields: `Ticker`, `Nome`, `Setor`, `Nota`, `PrecoAtual`, `VariacaoHoje`. The frontend already renders and submits `dy`; without this change the value is silently dropped.

## Goals / Non-Goals

**Goals:**
- Add `DY float64` to the `Stock` domain model, persisted via GORM auto-migrate
- Expose `dy` in all request/response DTOs
- Propagate `dy` in `CreateStock` and `UpdateStock` handlers and the service's update merge
- Accept `sort=dy` in `ListStocks` (descending)

**Non-Goals:**
- Computing `dy` automatically from dividends data
- Changing API versioning or authentication
- Modifying any other field or endpoint behavior

## Decisions

**D1 — float64, optional, defaults to 0.0**
`dy` is an optional field (not all stocks pay dividends). Using `float64` with a zero default keeps backward compatibility: existing DB rows get `0.0` without a migration script. Validation requires `dy >= 0` (negative yield is meaningless here).

**D2 — GORM auto-migrate handles the column**
No manual SQL migration is needed. GORM's `AutoMigrate` adds the `dy` column with a zero default on the next startup. Rollback: dropping the column is a manual DBA operation; application rollback just ignores the column.

**D3 — sort=dy added alongside nota/variacao**
The handler already validates the `sort` query parameter against an allowlist. Adding `"dy"` to that list and the repository's `ORDER BY` switch is sufficient; no new abstraction needed.

## Risks / Trade-offs

- [Risk] Zero default for `dy` on existing rows is semantically ambiguous (unknown vs. truly 0%) → Accepted; the frontend can distinguish by showing "—" for new vs. legacy records if needed, but the API does not need to model this distinction.
- [Risk] GORM auto-migrate runs on startup; a failed migration blocks startup → Low probability for an additive column with a zero default.

## Migration Plan

1. Merge and deploy the updated service.
2. On startup, GORM `AutoMigrate` adds `dy NUMERIC DEFAULT 0` to the `stocks` table.
3. Existing rows default to `dy = 0.0`.
4. No seed or backfill script required.

**Rollback**: revert the deployment; the extra `dy` column is harmless to the previous binary (GORM ignores unknown columns during reads).
