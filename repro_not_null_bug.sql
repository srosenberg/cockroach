-- Repro for NOT NULL constraint bug in runSchemaChangesInTxn (backfill.go:2741).
-- The `continue` at line 2741 only continues the inner loop, causing the
-- temporary check constraint to be re-added and the column to remain NULL.

DROP TABLE IF EXISTS bug_repro;
DROP TABLE IF EXISTS bug_repro_correct;

-- Correct behavior: non-transactional DDL.
CREATE TABLE bug_repro_correct (x INT);
ALTER TABLE bug_repro_correct ALTER COLUMN x SET NOT NULL;
-- Expected: x INT8 NOT NULL, no check constraint.
SHOW CREATE TABLE bug_repro_correct;

-- Buggy behavior: transactional DDL on a new table.
SET autocommit_before_ddl = off;
BEGIN;
CREATE TABLE bug_repro (x INT);
ALTER TABLE bug_repro ALTER COLUMN x SET NOT NULL;
COMMIT;
-- Expected: x INT8 NOT NULL, no check constraint.
-- Actual:   x INT8 NULL with CONSTRAINT x_auto_not_null CHECK (x IS NOT NULL).
SHOW CREATE TABLE bug_repro;

DROP TABLE bug_repro;
DROP TABLE bug_repro_correct;
