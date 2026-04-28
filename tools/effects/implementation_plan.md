# Effects Tool — Implementation Plan

A small Go AST walker that mechanically derives an upper bound on the
read/write heap effects of Go functions, modeled on the `ftpt` /
`R(E)` machinery from Rosenberg's Region Logic dissertation
(*Region Logic: Local Reasoning for Java Programs and its Application*,
2011, §2.4–2.7 and §3.3.7–3.3.10).

Concurrency is **out of scope**. The tool produces a sound (over-)
approximation of the heap regions a function may read or write.

## Effect Vocabulary

Each effect is `kind + path` where `kind ∈ {rd, wr, send, recv}` and
`path` is a textual location:

| Form          | Meaning                              |
|---------------|--------------------------------------|
| `x`           | local / package-scope variable       |
| `x.f`         | field selection                      |
| `x[*]`        | any element of slice/map/array `x`   |
| `*x`          | dereference                          |
| `alloc`       | allocation on the heap (`new`, `make`, `&T{}`) |

The tool currently treats `send` / `recv` symmetrically with `rd` / `wr`
on the channel value; channel ordering / blocking semantics are not modeled.

## What Is Implemented

### 1. Syntax-directed `ftpt` over expressions
`ftpt(e ast.Expr) effectSet` covers:
`BasicLit, Ident, SelectorExpr, IndexExpr, IndexListExpr, SliceExpr,
StarExpr, UnaryExpr (incl. & on composite literals), BinaryExpr,
ParenExpr, CallExpr, TypeAssertExpr, CompositeLit, KeyValueExpr,
FuncLit, Ellipsis`.

### 2. Statement walker
`stmt()` covers all statement forms (assign / inc-dec / if / for /
range / switch / type-switch / select / send / defer / go / branch /
labeled / decl / block). `lvalEffect()` decomposes l-values into the
write atom plus the reads needed to resolve the path.

### 3. Interprocedural propagation (intra-package)
- `collectFuncs()` builds `funcInfo` per `*ast.FuncDecl`
  (`recvName`, `formals`, `effects`, `opaque`).
- `calleesOf()` walks each body to collect intra-package call edges.
- `topoOrder()` runs Kahn's algorithm in reverse-topological order;
  any function left in a cycle is marked `opaque` (recursion).
- `inlineCallee()` builds a substitution map
  (`recvName → exprText(recv)`, `formals[i] → exprText(args[i])`)
  and rewrites the callee's path roots via `substPath()`.
- Generic methods are canonicalized through `*types.Func.Origin()`
  on both the call-site selection and the declaration side, so
  instantiated and uninstantiated `Func` objects map to the same node.

### 4. Builtin specials (`handleBuiltin`)
Type-correct heap atoms are emitted for:
`new, make` (→ `wr alloc`), `copy` (→ `wr dst[*]`, `rd src[*]`),
`append` (→ `wr alloc`, `rd dst[*]`), `delete` (→ `wr m[*]`),
`clear` (→ `wr x[*]`), `len, cap, min, max, complex, real, imag,
panic, print, println, recover, close`.

### 5. Filtering
The `-heap` flag suppresses pure local-only atoms (paths that don't
touch a field, dereference, index, or `alloc`), so the output focuses
on heap-visible effects.

### 6. Driver
`main()` loads packages via `golang.org/x/tools/go/packages`
(`NeedTypes | NeedSyntax | NeedTypesInfo`), processes them in
topological order, and prints effects per function in source order.

### 7. Atom subsumption / canonicalization
After each function's effect set is computed, `canonicalize()` runs
two passes:

1. **Slice-op stripping** — `s[i:j]` produces a slice header pointing
   to the same backing array as `s`, so `s[i:j][*]` and `s[*]` denote
   the same memory region. Every path is rewritten to remove embedded
   slice expressions (`[a:b]`, `[:b]`, `[a:]`, `[:]`, `[a:b:c]`),
   identified by a top-level `:` inside a bracket group. Index
   expressions `[i]` (no top-level colon) are preserved — they read a
   distinct element, possibly of a different type (`[][]int`).
2. **Bracket-peel subsumption** — `<k> X[*]` is dropped when
   `<k> Y[*]` is present and `Y` is reachable from `X` by peeling
   trailing balanced `[...]` segments. Catches residual nesting that
   step 1 doesn't normalize (e.g. `s[i][*]` when `s[*]` is present).

Both passes are per-kind. They cleaned up the previously visible
substitution artifacts where a callee's `rd dAtA[*]` got rewritten to
`rd dAtA[:size][*]` after the caller passed `dAtA[:size]`.

### 8. SliceExpr footprint
`ftpt(SliceExpr)` reads only the source slice header (and the bound
expressions) — it does **not** add an element-read atom. Sub-slicing
in Go is header construction; element reads happen later when the
sub-slice is indexed or copied from. This is contrary to an earlier,
over-eager version that added `rd <X>[*]` for every slice expression.

### 9. Compact, grouped output
`renderEffects()` formats a function's effect set for human reading.
By default, atoms whose paths share a leading identifier (paths of
the form `<id>.<tail>`) are grouped under that pivot — e.g.
`rd r.buffer, rd r.head, rd r.tail` becomes `rd r.{buffer, head, tail}`.
Grouping is single-level (split on the first `.`); paths with no
top-level `.` (singletons like `wr alloc`, `*p`, `buf[*]`) print on
their own line. Within a group, tails sort lexicographically; groups
larger than 8 items truncate to 7 entries plus `... (+N more)`.
A `-full` flag prints one atom per line (the previous format).
The full set is always retained internally; truncation is purely
presentational.

## Verified Examples

- `pkg/util/fast_int_map.go` — mixed value/pointer receivers, range
  over map, slice access. `(FastIntMap) GetDefault` correctly inlines
  `Get`'s reads.
- `pkg/util/every_n.go` — trivial struct + mutex.
- `pkg/util/ring/ring_buffer.go` — generic `Buffer[T]` with many
  methods. `Reserve` propagates `resize`'s effects through the call site
  (including `wr alloc`); `copy()` emits the right element-level atoms.

## Known Limitations / Reasonable Next Steps

Listed in rough order of impact on output quality.

### (3) Freshness (`fr G`) — biggest noise reduction
The dissertation's seq rule cancels `wr {x}'f` when paired with
`fr {x}` produced by a fresh allocation in the same body. Without
this, every function that uses a builder / copy helper carries write
effects on the helper's local buffers.

Concretely visible today: `resize` and `maybeGrow` emit
`wr newBuffer[*]` and `wr newBuffer[cap(r.buffer)-r.head:][*]`
because `copyTo`'s `wr dst[*]` is substituted with `newBuffer`,
even though `newBuffer` is freshly allocated and not yet escaped.

**Sketch:** in `stmt()`, when we see
`x := make(...) | &T{...} | new(T) | <fresh-call>`, record `x ∈ Fresh`.
Invalidate on any write that stores `x` into a non-fresh location,
return of `x`, or pass of `x` to an opaque callee. After computing the
function's effect set, drop write atoms whose root path is in `Fresh`.

### (4) Local-shadow false positives
`isLocalIdent()` filters by string only. When a local variable shadows
a package-level constant of the same name (e.g. `numVals`), the const's
effects survive the filter. Fix: resolve the actual `*types.Object` per
ident occurrence and decide based on its scope.

### (5) Receiver rendering as `this`
Methods print effects as `r.field` because `r` is the source-level
receiver name. For composable specs we want a uniform `this.field`
rendering so substitution at call sites is symmetric across callers
that name the receiver differently. Mechanical: rewrite `recvName`
to a fixed sentinel before storing; substitute back at call sites.

### (6) `make(map)` vs `make(slice)` distinction
Both currently emit `wr alloc`. Cosmetic; could distinguish to
support type-aware downstream analysis (e.g. "this function only
allocates maps").

### (7) Cross-package propagation
Calls to functions in packages outside the load set are opaque.
A natural extension is to compute effects bottom-up across a package
DAG (via `packages.Load` with `NeedDeps`) and persist a small per-
package effect summary, similar to how the SSA package caches
function summaries.

### (8) Interface dispatch
Calls through interface method sets are opaque. Could be improved by
collecting all concrete implementors in the load set and joining
their effects (sound but coarse), or by using `golang.org/x/tools/go/
callgraph/{vta,rta}` to refine the target set.

### (9) Pointer / aliasing precision
Currently a write to `*p` is recorded as `wr *p`, with no attempt to
relate `p` to anything it might alias. Region Logic's region
abstraction is the principled way in; for a Go-flavored quick win,
we could integrate `golang.org/x/tools/go/pointer` (Andersen-style
points-to) and use it to canonicalize l-value roots.

### (10) Output formats
Today the tool prints a flat per-function listing. Useful follow-ups:
- A `-json` mode that emits machine-readable summaries.
- A `-diff` mode that compares effect sets across two revisions
  (for review: "this PR added a write to `r.buffer`").
- An `-annotate` mode that writes `// effects: ...` comments above
  each function.

### (11) Closures
`FuncLit` is currently treated as a plain expression — its body is not
walked into the enclosing function's effects, and captured variables
are not tracked. A minimal improvement: when a FuncLit is immediately
called, walk it inline; otherwise, treat its captures as `rd` (the
closure may read them later).

### (12) ~~Atom subsumption / canonicalization~~ — DONE
Implemented as item 7 above (slice-op stripping + bracket-peel).

### (13) ~~Bare slice-expression artifacts from substitution~~ — DONE
Slice-op stripping in item 7 collapsed these (`rd dAtA[:size]` → `rd dAtA`,
`wr dAtA[:size][*]` → `wr dAtA[*]`, etc.).

### (14) Type-conversion expressions appearing as paths
Patterns like `rd uint64(len(m.AddressField))` survive because a
type-cast around an `len()` call gets used as the textual path of an
l-value or r-value at a call site. The conversion itself reads
nothing (it's just a representation change). Possible cleanup: when
recording a path that starts with `<TypeName>(...)`, peel the
conversion and use the inner expression's path. Low impact (visible
mainly in generated pb.go files), so deprioritize.

### (15) Callee-local rewrite (let-binding inlining)
Substitution at call sites only rewrites the receiver and formal
parameters. Callee-locals like `tailElements := r.buffer[r.head:]`
inside `copyTo` propagate verbatim into the caller's atoms, producing
meaningless paths like `rd tailElements[*]` in `(*Buffer[T]) all`.

**Approach (pre-inline rewrite):** before storing a function's effect
set in `funcInfo.effects`, walk the function body to collect
single-assignment local bindings of the form `x := <expr>` where
`<expr>` references only caller-visible names (formals, receiver,
package-level names). Build a callee-local substitution map
`x → <expr-text>`. Rewrite the function's own effects through this
map (using the existing `substAtom`). Then the slice-op stripping
from item 7 collapses the result onto a base path that already
exists, and dedup happens automatically.

For `tailElements`: `tailElements → r.buffer[r.head:]`, so
`rd tailElements[*]` rewrites to `rd r.buffer[r.head:][*]`,
strips to `rd r.buffer[*]`, dedupes against the existing
`rd r.buffer[*]`. Net effect: leaked-local atoms vanish.

**Constraints to respect:**
- Only single-assignment locals. Skip names that are reassigned,
  appear in a range/for `:=`, or are written through.
- Skip locals whose RHS references other unresolved callee-locals
  (would just shift the leak); a fixed-point pass can chain bindings.
- Watch for shadowing: only consider the binding visible at the use
  site; for the simple "single-assignment within the function" case
  this is automatic.

## Downstream Consumers

### Deep static analysis skill
`.claude/skills/deep-static-analysis/SKILL.md` performs evidence-backed
bug analysis of a target function. Mechanically derived effects could
augment several phases:

- **Phase 2 (Function Trace Table)** — the `Side effects` column is
  exactly what the effects tool produces. Today it's manually
  reconstructed; with effects it's a lookup, with `file:line` evidence
  for free.
- **Phase 3 (Dataflow Ledger)** — `Modified at` / `Used at` /
  `Never modified` claims need `file:line` evidence per VALUE. Effects
  give a mechanical "every function that writes `r.foo`" list.
- **Phase 3 (Concurrency checklist)** — "verify every read and write
  of mutex-guarded fields holds the lock". Effects enumerate every
  `wr r.mu.<field>` site; cross-referencing against `Lock`/`Unlock`
  call sites becomes a list-vs-list check rather than ad-hoc grep.
- **Phase 3 (Range-loop value-copy)** — effects already report writes
  through range copies; direct candidate enumeration.
- **Phase 5 (Divergence claims)** — claims of the form "this function
  doesn't mutate X" can be backed (or refuted) by checking that no
  `wr X` appears in the effect set. Useful especially for refutations.

**Caveats to disclose to the skill's analyst:**
- Effects are **flow-insensitive within a function** (union over all
  paths). Path-sensitive claims like "skipped on one early return"
  need a per-path effect refinement we don't yet produce.
- **Cross-package calls are opaque** (only argument footprints).
- **Interface dispatch is opaque** (no points-to / implementor
  enumeration).

**Prerequisites before integrating:** finish items (3) freshness and
(15) callee-local rewrite. Without those, output noise (spurious
`wr <fresh>[*]` and `rd <callee-local>[*]` atoms) would dilute the
value of the side-effect column and produce false leads in the
ledger.

**Sketch of integration:** invoke `tools/effects` on the target's
package as part of context gathering before Phase 2. Inject the
target's effect summary plus those of its key callees into the
analyst's working set. Cite atoms (e.g. `wr r.head`) alongside
`file:line` references in the ledger and divergence claims.

## Future Applications

Beyond the deep-static-analysis skill, mechanically derived effects
unlock several light-weight bug-finding and code-quality applications.
Heavyweight verification (proof-carrying code, full Region Logic
verification) is **not** in scope; everything below is a "lint-style"
or "review-augmentation" use of effects as a precise, mechanically
derived data source.

Ordered by approximate cost-benefit:

### A. Read-only / purity attestation (low cost, high leverage)
Methods named `Get*`, `View*`, `Snapshot*`, `Iterator*`, etc. should
have an empty `wr` set (modulo `wr alloc` for fresh return values).
A single offending atom is either a contract violation or a hidden
side effect worth reviewing. One-shot exhaustive audit; finds real
bugs in already-shipped code.

### B. PR diff-of-effects (low cost, high reviewer value)
For each PR, run effects on `before` and `after`, diff per-function
sets. Surface to the reviewer: *"this PR added writes to
`r.committed` from `(*Replica).RefreshLease`, intended?"* Catches
accidental scope creep — a "refactor-only" PR that adds writes to
new state is a red flag.

### C. Risk-weighted coverage (low cost, strategic)
Coverage × effects: rank functions by `criticality(effect set) ×
inverse(coverage)`. Criticality comes from a curated allowlist of
"sensitive" paths (`txn.{committed,timestamp}`, `lease.{epoch,...}`,
`desc.{generation}`, persisted protobuf fields). Reframes "improve
coverage" from the impossible "cover everything" to a few hundred
high-risk under-tested functions. Changes how testing time is spent.

### D. Refactor impact analysis (trivial cost, on-demand)
"What breaks if I rename / split / move field `r.buffer`?" — query
effects for any atom mentioning the field. Today this is grep with
high noise (comments, unrelated identifiers); effects give the
precise list of affected functions.

### E. Lock-discipline verifier (medium cost, high leverage)
For every struct with `mu sync.Mutex`, enumerate fields whose path
starts `mu.`. Assert every function with `wr/rd r.mu.x` either
(a) calls `r.mu.Lock()` at entry, or (b) is only called by functions
in (a). Mechanical, exhaustive, finds real bugs. Requires
lock-acquisition tracking but otherwise straightforward.

### F. Linter false-positive suppression (medium cost)
- `unused` / `structcheck`: a field flagged as unused is genuinely
  dead iff no function's effect set mentions it. Confirms the
  no-references direction; doesn't help with reflection / proto.
- `nilness`: a `wr r.foo` along all paths between entry and a use
  site justifies treating `r.foo` as non-nil; conversely, no `wr`
  along the path strengthens a nilness warning.
- Custom CockroachDB linters (e.g. `redactcheck`): refine "is this a
  field of a sensitive type" by watching `wr` paths to redactable types.

### G. Migration / mixed-version safety hint (medium cost)
Any function with `wr` whose root path resolves to a serialized type
(descriptor, span, raft entry, persisted protobuf field) is a
candidate that may need a `clusterversion` gate. Produces a
reviewable shortlist; expect false positives but coverage is high.

### H. Idempotency hint for retriable paths (high cost, high value)
A function with `wr X` (X observable beyond the function) followed by
a possible error return is potentially non-idempotent under retry.
Combined with retry classification, surfaces candidates for
retry-safety review. Needs control-flow + error-flow tracking that
the current tool doesn't yet have.

### I. Documentation augmentation (low cost, hygiene win)
Auto-generate `// effects: rd r.{a,b}, wr r.{c}` annotations above
exported methods. Doesn't catch bugs directly but raises contract
explicitness and makes drift visible: when a method is changed to
write a new field, the annotation needs updating, prompting review.

### J. Property / differential testing input generation (high cost, long-term)
Use the effect set as a coverage objective: generate inputs that
maximize write-set diversity across functions. Reframes coverage
from "lines hit" to "memory regions touched" — better proxy for
behavior coverage. Best paired with a fuzzer.

### Cost-benefit summary

| Application | Bug-finding leverage | Implementation effort |
|---|---|---|
| A. Read-only attestation | High (one-shot) | Low |
| B. PR diff-of-effects | Medium-high | Low |
| C. Risk-weighted coverage | High (strategic) | Low |
| D. Refactor impact | High (per-refactor) | Trivial |
| E. Lock-discipline verifier | High | Medium |
| F. Linter FP suppression | Medium | Medium |
| G. Migration safety hint | Medium | Medium |
| H. Idempotency hint | Medium-high | High |
| I. Doc augmentation | Low (bugs), high (hygiene) | Low |
| J. Property testing | High (long-term) | High |

**Suggested first cuts**: A (read-only attestation) and B (PR
diff-of-effects) for quickest wins; C (risk-weighted coverage) for
the strategic shift in how testing effort is allocated.

## Non-Goals

- Concurrency / happens-before / data-race reasoning.
- Termination / total correctness.
- Reasoning about effects of `unsafe`, cgo, or reflection.
- Producing proofs — only specifications (upper bounds).
