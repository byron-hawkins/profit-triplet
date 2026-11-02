# Project conversation history — November 2, 2026

This is a focused transcript summary of today's work on parsing symbolic annotations and proving acyclicity. It records the developer's decisions, the assistant's analyses, and current open issues so another person or agent can continue without treating an incomplete proof as established.

## Annotation grammar and clean parsing

- The developer iteratively updated `model/symbolic-abstraction.bnf` and annotations in `internal/profit-triplet/dag.go`, asking for repeated checks that the source annotations correspond cleanly to the grammar.
- The grammar evolved to support:
  - element, summary, and data-model annotations;
  - tags with function names or signatures and optional element locators;
  - comma-separated annotation labels;
  - explicit pre-condition and post-condition roles;
  - free-form inline Go fragments, multiline comment formatting, and comment closing forms;
  - substitutions with typed values and constraints.
- Earlier parsing reviews identified mismatches at each revision, including missing separators, comment decoration/closing syntax, signature interpretation, and ambiguous or malformed annotations. The developer corrected several of these and clarified that an apparent duplicate `atomic.Int32` annotation was a temporary reference copy, not an intended annotation.
- Later checks judged the annotations close to parseable, while noting that BNF correspondence is a syntax question and does not establish semantic truth or successful compilation.

## External abstraction boundary

- The developer explicitly requested that abstract models of external components not be required to reflect the concrete implementation. Proofs should stop at the annotated abstraction for all project proofs, unless a separate refinement/conformance proof is requested.
- Updated `model/proof-generation.md` accordingly:
  - external data models and operation summaries define the proof interface;
  - check models for internal consistency, typing, and sufficiency;
  - do not inspect dependency implementation to prove the abstract contract;
  - report `UNKNOWN` or a specification bug if the abstraction is missing or insufficient.
- The developer asked whether `slices.Contains` can be used as standard-library semantics. It is locally available and can be treated as a documented standard-library operation, but its argument must be a slice. For the vector struct model, the type-compatible expression is over `.elements`.
- The developer clarified the vector abstraction: `vector.Vector[*Item]` is modeled as a struct containing `elements []*Item`, not as a raw slice. This resolved the earlier ambiguity about whether `NewVector` should return a modeled array or a struct.

## Vector model and semantics

- The developer iteratively fixed Vector annotations after proof checks:
  - data model: `elements []*Item`;
  - `PushBack`: assigns `append(v.elements, item)` back to `v.elements`;
  - `Empty`: has a boolean result and checks `len(v.elements) == 0`;
  - `NewVector`: returns a pointer to a modeled Vector struct with an empty `elements` slice.
- Once the constructor return type and struct literal were fixed, the assistant confirmed the Vector summaries were internally consistent under the declared abstraction. No dependency-source refinement was required or performed.
- `go test ./internal/profit-triplet` passed during the latest proof attempt; this establishes compilation/tests for that package, not the truth of annotation contracts.

## Substitution applicability semantics

- The developer added `<maybe-invariant-list>` to `<substitution>` and asked whether its invariant constrains applicability of the containing substitution.
- The assistant confirmed the intended reading: a substitution case is usable only if its own value constraints and attached preconditions hold; such preconditions do not apply to sibling cases or become global function preconditions.
- Updated `model/proof-generation.md` to make this a formal proof rule. Attached postconditions describe the selected case's result and are checked after evaluating it.
- The developer then requested another acyclicity proof attempt based on the updated semantics and Vector fixes.

## Acyclicity proof attempts and current status

- The core local proof idea remains:
  - `Item.Link` requires `to.out.Empty()` and `from != to`.
  - It adds the edge `from → to` and the reverse incoming reference.
  - Since `to` has no outgoing edges, the new edge cannot close a cycle.
  - If every call site establishes this precondition and all graph mutations happen through `Link`, induction over `GraphWalk` preserves acyclicity.
- The developer clarified that the proof need not model `interval.SearchTree` item-by-item. An abstract history such as `EverInserted []*Item` can provide provenance: query results refer only to previously inserted item identities. This remains an abstract contract, not a model of the tree's internal structure.
- The latest proof attempt treated external models abstractly and found:
  - **Pass:** Vector `NewVector`, `PushBack`, and `Empty` are now type-consistent within their struct abstraction.
  - **Pass conditionally:** The local `Item.Link` argument is sound under its stated precondition.
  - **Open:** `AllIntersections` lacks a valid empty-history case; the nonempty substitution bounds do not cover all valid slices, so query behaviors are under-specified.
  - **Open:** `Item.AllIn` uses Vector `Size` and `At`, for which the abstract model currently lacks operation semantics.
  - **Open:** No abstract constructor contract initializes `EverInserted`; no complete `Insert`/`Delete` history semantics establish how it changes or whether updates occur only on success.
  - **Open:** `Item#constructor` is annotated as `return Item{}`, which does not describe the actual initialized fields or establish empty vectors/freshness.
  - **Open:** `GraphWalk#InsertItem` requires `item.Price.Base < item.Price.Bound`; that condition is not established for arbitrary inputs.
- Result remains **UNKNOWN / proof blocked**, not a counterexample to acyclicity. The package test passing is not a proof.

## Developer's current direction

- The developer said the proof should be less thorough and that some things need simplification. No specific proof obligations have yet been removed or replaced; keep the current proof result explicitly unknown until the intended simplified scope/rules are clarified.
- The developer requested a new dated conversation history, resulting in this file.

## Files relevant to continuing

- `model/symbolic-abstraction.bnf` — current annotation grammar.
- `model/proof-generation.md` — project-wide external abstraction boundary and proof process, including substitution applicability.
- `model/dag.proof-acyclic.md` — prior detailed conditional proof ledger; it may now be more detailed than the developer wants.
- `internal/profit-triplet/dag.go` — current annotations and implementation under analysis.
- `model/dag.cfg.md` and `model/cfg-generation-rules.md` — generated per-function CFGs and reusable graphing conventions.

## Methodological summary

The project is using an agentic proof workflow:

1. Parse annotations according to the evolving BNF.
2. Derive per-function control-flow graphs, with call sites referencing separate callee graphs.
3. Use Go source for project-owned control flow and state mutations, while treating external component models as axiomatic abstractions.
4. Turn each proof into explicit sub-goals for types, contracts, call preconditions, data provenance, and loop induction.
5. Fail closed: unsupported or insufficient annotations produce `UNKNOWN` or a specification bug, never an assumed proof.
6. Simplify the proof scope only when the developer explicitly chooses which details can be abstracted away.

The most important distinction is between a true mathematical argument and evidence that its premises hold. Here the local `Link` lemma is understandable, but the current query-history contracts do not yet justify all premises needed at its call sites.

## Later work and proof updates (afternoon/evening)

The following items were added after the initial summary above. Where they
conflict with its earlier “current status” bullets, this later addendum
records the subsequent annotation and proof state.

- Regenerated `model/doc/dag.cfg.md` after model documents were split into
  `model/meta`, `model/proof`, `model/bug`, and `model/doc`. The refreshed CFG
  includes the then-current vector and atomic summaries, `AllIn`, and
  separate call-site references for `GraphWalk` and `GraphStep`.
- Created `model/doc/model-check-results.md` to summarize proof documents,
  bug reports, and supporting test output. The test mutation campaign starts
  from 10 generated inputs, creates 10 copies with either a relevant price
  mutation or an input swap, and tests the original plus all copies. A later
  captured run failed on `input-8` (the eighth mutated copy) with an
  equal-endpoint interval at price `162`; the log does not reveal which
  mutation was applied.
- `GraphWalk` gained a boolean precondition constraining every input price to
  `1..9999`. The initial version was statement syntax rather than a boolean
  expression; it was subsequently changed to an immediately invoked
  `func() bool` loop. Under that precondition, the constructor establishes
  `Base` in `[1,9999]` and `Bound == math.MaxInt32`.
- Added `model/proof/fresh-item-valid.md`, proving the
  `GraphWalk#InsertItem` range condition under the input-range precondition
  and the previously established link-freshness fact. The proof applies the
  `GraphStep#MatchTruncate` alias to the matched item's bound and uses
  `match != item` to show the current new item's price is unaffected.
- `ItemTruncate` became proof-reachable through a `call` label. Its truncate
  assignment sets the matched item's `Bound` to the query price; the
  subsequent `ItemTruncate#Insert` precondition requires `Bound > Base`.
  The earlier report used a broad query abstraction and a `Base > a`
  counterexample. The `AllIntersections` postcondition was then changed to
  require returned items have `Base <= a`, allowing item bases below the
  query price but also allowing equality.
- The final counterexample to `ItemTruncate#Insert` uses duplicate prices:
  inputs `[20, 30, 20]`. The abstract query for the third item may return the
  first item with `Base == a == 20`; the caller guard `a < match.Bound` holds,
  but truncation sets `Bound == Base == 20`, refuting the strict precondition.
  The corrected report is `model/bug/invalid-item-rescope.md`. Initial
  `GraphWalk#InsertItem` validity does not prevent this because the
  `ItemTruncate` mutation changes an object already referenced in
  `EverInserted`.
- Updated `model/doc/model-check-results.md` to list the new range proof and
  the corrected equal-price bug separately. The earlier entries are grouped
  under an “Existing results” section explicitly marked as associated with
  git branch `iteration/bug-hunt-1`, not top-of-trunk content.

## Latest proof state

- **Established conditionally:** `GraphWalk#InsertItem`'s precondition, via
  `model/proof/fresh-item-valid.md`, assuming its input precondition and
  match freshness.
- **Refuted in the current annotated abstraction:** the precondition on
  `ItemTruncate#Insert`, via
  `model/bug/invalid-item-rescope.md`, because the query postcondition permits
  `match.Price.Base == a`.
- **Still distinct:** the prior `Item.out` acyclicity argument depends on
  freshness; it is not itself a proof that every interval truncation remains
  a valid nonempty span.
