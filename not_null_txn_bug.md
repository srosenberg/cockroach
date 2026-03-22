## sql: ALTER COLUMN SET NOT NULL on table created in same transaction produces wrong schema

When `ALTER COLUMN SET NOT NULL` is applied to a table created in the same transaction, the column remains `NULL` in the descriptor and a residual `CHECK (col IS NOT NULL)` constraint is left behind instead.

### Root cause

In `runSchemaChangesInTxn` (`pkg/sql/backfill.go:2730-2744`), the `continue` at line 2741 only continues the inner `for i := range tableDesc.Checks` loop, not the outer `for _, c := range constraintAdditionMutations` loop. After the inner loop exits, line 2744 (`tableDesc.Checks = append(...)`) executes unconditionally, re-adding the temporary check constraint that was just removed.

Introduced in commit 83093002560b (`sql: enforce ALTER COLUMN SET NOT NULL for tables created in same txn`, Nov 2024).

### Impact

- Column descriptor says `NULL` but inserts of NULL are rejected by the check constraint — schema inspection is misleading.
- Wrong SQLSTATE on NULL violation: `23514` (check_violation) instead of `23502` (not_null_violation). Applications checking error codes will behave incorrectly.
- ORMs and migration tools that inspect the schema will think the column is nullable.

### Repro

```sql
SET autocommit_before_ddl = off;
BEGIN;
CREATE TABLE bug_repro (x INT);
ALTER TABLE bug_repro ALTER COLUMN x SET NOT NULL;
COMMIT;
SHOW CREATE TABLE bug_repro;
```

**Actual:**
```sql
CREATE TABLE public.bug_repro (
  x INT8 NULL,
  ...
  CONSTRAINT x_auto_not_null CHECK (x IS NOT NULL)
);
```

**Expected:**
```sql
CREATE TABLE public.bug_repro (
  x INT8 NOT NULL,
  ...
  -- no check constraint
);
```

Contrast with the non-transactional case which works correctly:
```sql
CREATE TABLE bug_repro_correct (x INT);
ALTER TABLE bug_repro_correct ALTER COLUMN x SET NOT NULL;
SHOW CREATE TABLE bug_repro_correct;
-- x INT8 NOT NULL, no check constraint
```

### Fix direction

The `continue` at `backfill.go:2741` needs to target the outer loop (e.g., via a labeled `continue`) so that line 2744 is skipped for NOT NULL column constraints.
