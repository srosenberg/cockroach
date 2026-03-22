## sql/backfill: bugs in `addConstraints`

### 1. No-op validity update for unique-without-index constraints (bug)

At `pkg/sql/backfill.go:696`, the range loop over `scTable.UniqueWithoutIndexConstraints` (a `[]UniqueWithoutIndexConstraint` — a slice of **values**) copies each element into `c`. The assignment at line 706:

```go
c.Validity = descpb.ConstraintValidity_Validating
```

mutates the copy, not the slice element. This is a no-op. The intent is to set the constraint to `Validating` on retry or during rollback of `DROP CONSTRAINT`, but it silently does nothing.

Compare with the FK case (line 643–644) which correctly takes a pointer via index:

```go
for j := range scTable.OutboundFKs {
    def := &scTable.OutboundFKs[j]
```

And the check constraint case (line 607) which works because `Checks` is `[]*TableDescriptor_CheckConstraint` (slice of pointers).

**Fix:** Range by index and take the address:

```go
for i := range scTable.UniqueWithoutIndexConstraints {
    c := &scTable.UniqueWithoutIndexConstraints[i]
```

### 2. Dead code: `fksByBackrefTable` map never used

`addConstraints` (`pkg/sql/backfill.go:579`) builds a `fksByBackrefTable` map (lines 584–592) that is never referenced afterward. This is dead code left behind when commit 68378d6 removed the `WaitToUpdateLeases` loop but did not clean up the map construction.

### 3. Asymmetry with `dropConstraints` and docstring mismatch

Unlike its counterpart `dropConstraints` (line 428), `addConstraints`:

1. Does not have a second read transaction to fetch committed descriptor versions.
2. Returns only `error` instead of `(map[descpb.ID]catalog.TableDescriptor, error)`.
3. Does not fulfill its docstring promise (lines 576–578) of waiting "until the entire cluster is on the new version of the table descriptor."

The backreference tables are correctly updated within the single transaction (lines 680, 687–691) and committed atomically (line 722), so the risk here is latent rather than acute.

### Suggested fix

- **Bug #1:** Change the UWI loop to range by index and take a pointer (matching the FK pattern).
- **Dead code #2:** Remove the unused `fksByBackrefTable` code.
- **Asymmetry #3:** Either update the docstring to reflect actual behavior, or add a second read transaction and return type to match `dropConstraints`.
