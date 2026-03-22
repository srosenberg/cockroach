---
name: deep-static-analysis
description: Perform repo-wide semi-formal static bug analysis of a target function or method. Use for under-tested code when you want evidence-backed bug candidates, dataflow tracing, and ranked findings without executing code.
argument-hint: "[file:function | package.symbol | symbol]"
disable-model-invocation: true
context: fork
agent: Explore
---

# Deep Static Analysis

Use this skill to perform **deep, repo-wide static analysis** of a target function or method.

This is **not** a PR review and **not** a snippet review.
You have access to the **entire repository**, and you must use that advantage.

Your job is to decide whether the target is likely buggy, fragile, under-specified, or relying on hidden invariants, and to produce only **evidence-backed** findings.

Target:
$ARGUMENTS

ultrathink

## Operating mode

- Perform **static** analysis only.
- Do **not** run tests, binaries, servers, or mutate files.
- Use code search, file reading, package docs, nearby tests, comments, and design docs.
- Prefer **fewer high-confidence findings** over many speculative ones.
- If the evidence is weak, say so clearly.

## Non-negotiable rules

1. Never claim behavior without citing concrete `file:line` evidence or a directly derived local consequence.
2. Every finding must be a **proof obligation** with all of:
   - precondition / triggering state
   - control-flow path
   - dataflow path
   - exact divergence point
   - externally visible bad outcome
3. Label each result as exactly one of:
   - `VERIFIED BUG CANDIDATE`
   - `PLAUSIBLE RISK`
   - `DISPROVED HYPOTHESIS`
4. Before finalizing any bug candidate, actively search for counterevidence:
   - guards in callers
   - invariants in comments/docs
   - type/system constraints
   - wrapper semantics
   - retries / sanitization / normalization
   - tests that already cover the scenario
5. If you cannot close the proof obligation, downgrade the claim.

## Repository-wide mindset

Because you have the full codebase, do not stop at the target function.
You must inspect, as needed:

- direct callers
- transitive callers that constrain inputs
- direct callees and helpers
- interface definitions and alternate implementations
- nearby tests and datadriven cases
- package comments and design docs
- related state structs, enums, settings, and error types

For CockroachDB-style distributed Go code, pay extra attention to:

- transaction state and retry safety
- idempotency across retries / partial failure
- context propagation / cancellation
- deferred cleanup and early returns
- error wrapping, classification, and swallowing
- stale metadata, leases, caches, or descriptors
- timestamp / MVCC / boundary assumptions
- span / key / range boundary handling
- goroutine / channel lifecycle
- feature flags, version gates, and cluster settings
- ownership / aliasing of mutable objects and buffers

## Required analysis process

### Phase 0: Resolve the target precisely

Resolve the exact symbol named by `$ARGUMENTS`.

Produce:

- canonical symbol name
- file and line range
- package
- receiver/type if any
- directly related tests
- closest interfaces / implementations / wrappers

If the argument is ambiguous:
- resolve the most likely match
- list 1-2 rejected alternatives
- explain why your choice is the best match

### Phase 1: Reconstruct the contract

Read the full target and enough surrounding context to answer:

- What is the apparent purpose of this function?
- What inputs are assumed valid?
- What outputs / side effects are promised?
- What invariants are expected before entry and after return?
- What failure modes are intended vs unintended?

Write these as formal premises:

`PREMISE C1: ...`
`PREMISE C2: ...`

Only include premises supported by evidence from code/comments/tests/docs.

### Phase 2: Build the interprocedural function trace

Construct a `FUNCTION TRACE TABLE` for the target and the small set of functions that actually matter.

Use this format:

| Function/Method | File:Line | Inputs | Outputs | Side effects | Why relevant | Evidence |
|---|---|---|---|---|---|---|

Include:
- target function
- key callees
- key callers
- wrappers that transform arguments/errors/state
- defer paths if semantically important

Do not include irrelevant helpers.

### Phase 3: Perform explicit dataflow analysis

Track only the variables / objects / state that matter to correctness.

Required categories to consider:
- input parameters
- receiver fields
- context and cancellation state
- transactions / descriptors / leases / timestamps / spans / keys if relevant
- errors
- mutable collections / pointers / aliases
- booleans or enums that gate behavior
- resources requiring cleanup

Create a `DATAFLOW LEDGER` using this format:

`VALUE V1: <name>`
- Created/received at: `file:line`
- Validated at: `file:line` or `NONE`
- Modified at: `file:line(s)` or `NEVER MODIFIED`
- Aliased/stored into: `file:line(s)` or `NONE`
- Used at: `file:line(s)`
- Escapes beyond function?: `YES/NO`
- Invariant expected: `...`
- Threats to invariant: `...`

Be explicit about:
- where a value is first trusted
- whether it is revalidated after transformation
- whether a stale / nil / zero / partially initialized value can propagate
- whether cleanup is coupled to the success path only

#### Go-specific value semantics checklist

For every `for _, v := range collection` loop where `v` is **mutated**:

1. Determine the element type of `collection`:
   - `[]T` (value type): `v` is a **copy** — mutations are lost unless written back via index
   - `[]*T` (pointer type): `v` is a pointer copy — mutations via `v.Field = ...` are visible
2. If `v` is mutated and `collection` is a value-type slice, flag as a divergence candidate
3. Compare with sibling loops in the same function that iterate over similar slices — inconsistent patterns (some using index, some using range copy) are a strong signal
4. Check whether the mutation is the **intent** (e.g., setting validity, status) vs incidental

### Phase 4: Enumerate path conditions

Build a `PATH MATRIX`.

For each meaningful path through the function, include:

| Path ID | Guard / precondition | Key branches taken | Important calls | Return / side effect | Why this path matters |
|---|---|---|---|---|---|

You must include:
- happy path
- all meaningful early returns
- error paths
- defer-dependent paths
- boundary / zero / nil paths if plausible
- retry or reentry paths if plausible

Do not just paraphrase code.
State the actual semantic difference between paths.

### Phase 5: Identify divergences from the reconstructed contract

For each suspicious mismatch, write a formal divergence claim:

`CLAIM D1: At file:line, <actual behavior> contradicts PREMISE Cn because <reason>.`

Every claim must include:
- exact code location
- the premise it contradicts
- the triggering input/state
- the downstream consequence

Good divergence examples:
- validation exists for one path but not another
- error is converted into success or less severe class
- cleanup is skipped on one early return
- state mutation occurs before possible failure and is not rolled back
- retry causes duplicate/non-idempotent side effect
- caller assumes callee preserves invariant but callee weakens it
- stale cached state can bypass a required refresh
- a nil/zero/default value silently means “success” or “use prior state”
- mutation on a range-loop copy of a value type (lost write)
- inconsistent iteration patterns across sibling loops (index vs copy) on similar slice types

### Phase 6: Attack your own hypothesis

For every candidate bug, perform an `ALTERNATIVE HYPOTHESIS CHECK`:

- What evidence would make this *not* a bug?
- Which callers or tests suggest the precondition is impossible?
- Is there a surrounding invariant that makes the path unreachable?
- Is the behavior intentional and documented?
- Is there compensating cleanup or retry logic elsewhere?

Use this format:

`ALT H1: If the candidate were false, we would expect ...`
- Searched:
- Found:
- Conclusion: `REFUTED` / `PARTIALLY REFUTED` / `NOT REFUTED`

A finding cannot be labeled `VERIFIED BUG CANDIDATE` unless you attempted to refute it.

### Phase 7: Rank the findings

Produce at most 5 ranked findings.

For each finding use this exact structure:

#### Rank N — <short title>
- Classification: `VERIFIED BUG CANDIDATE` / `PLAUSIBLE RISK` / `DISPROVED HYPOTHESIS`
- Confidence: `high` / `medium` / `low`
- Primary location: `file:line`
- Triggering precondition:
- Dataflow path:
- Control-flow path:
- Divergence claim(s):
- Counterevidence checked:
- Likely user-visible effect:
- Why this matters in production:
- Minimal validating test to add:
- Minimal fix direction:

A finding is only `high` confidence if:
- the path is concrete
- the dataflow is explicit
- the divergence point is precise
- counterevidence has been checked

## Final output format

Return the report in exactly this order:

1. `Target Resolved`
2. `Contract Premises`
3. `Function Trace Table`
4. `Dataflow Ledger`
5. `Path Matrix`
6. `Divergence Claims`
7. `Alternative Hypothesis Checks`
8. `Ranked Findings`
9. `Bottom Line`

## Bottom Line requirements

The `Bottom Line` must answer:

- Is there a likely real bug here?
- If yes, what is the single best candidate?
- What assumption had to fail for that bug to exist?
- What is the one test that would most efficiently validate or falsify it?

## Quality bar

Do not stop at “this looks suspicious.”
Do not stop at the crash site / error site if an earlier root cause may exist.
Do not infer semantics from names alone.
Do not trust one local snippet when repository context can disprove it.

Your standard is:
**premises -> trace -> dataflow -> divergence -> attempted refutation -> ranked conclusion**
