# Project conversation history — November 3, 2026

This is a focused transcript summary of today's work on the evolving
`dag.go` model, proof refresh, branch-specific findings, and formal
purchasability goals. It records established results separately from open
proof goals.

## Proof refresh after `dag.go` changes

- The developer requested that every proof in `model/proof/` be retried in
  dependency order after significant source changes, stopping at the first
  unprovable goal.
- Current proof files were:
  - `model/proof/linked-item-fresh.md`
  - `model/proof/dag-acyclic.md`
  - `model/proof/fresh-item-valid.md`
- The dependency order used was:
  1. `linked-item-fresh.md` (no proof dependency within this directory);
  2. `dag-acyclic.md` (depends on freshness);
  3. `fresh-item-valid.md` (depends on freshness).
- All three goals were judged provable under explicit assumptions and their
  files were rewritten:
  - **Freshness:** `AllIntersections` selects only entries from
    `EverInserted`; each `AllIn` predecessor is an earlier graph item; the
    current object is fresh and has empty `out`. The argument accounts for
    `GraphStep`'s boolean result, conditional `GraphWalk#InsertItem`,
    `ItemTruncate#Insert` re-appending an existing match, and `Join` changing
    only twin links. It assumes fresh object identity for each `NewItem`.
  - **Acyclicity:** induction over all `out` appends; each occurs through
    `Item.Link`, whose destination is the fresh empty-out current item and
    whose source is distinct. `Join`, price truncation, and tree/history
    effects do not mutate `out`. This depends on the freshness proof.
  - **Insert validity:** the current query alias filters matches by
    `match.Base <= query` and `match.Bound >= query`. The truncation branch
    requires `query > match.Base`, so the new bound is strictly greater than
    the base; if `FindSameInterval` reports a twin, insertion is skipped.
    The current item's price remains its constructor values, so
    `GraphWalk#InsertItem` is valid whenever reached. This proof depends on
    freshness and the `GraphWalk` input-range precondition.
- These are proofs under the project's annotation semantics and premises,
  not refinement proofs of external library implementations. In particular,
  callers must establish `GraphWalk`'s preconditions, and the fresh-object
  premise must be accepted or separately justified.

## CFG and results index refreshed

- Regenerated `model/doc/dag.cfg.md` from the latest annotated `dag.go`. It
  now covers `Item.Join`, the loop-based filtered `AllIntersections`
  summary, boolean-returning `GraphStep`, conditional graph insertion, and
  the two `ItemTruncate` paths.
- Updated `model/doc/model-check-results.md` to summarize the refreshed
  proofs. Earlier findings were moved to historical results and the
  `invalid-item-rescope.md` equal-price bug was assigned to branch
  `iteration/bug-hunt-2`, not `iteration/bug-hunt-1`. The original
  unrestricted-input insert bug and mutation-test output remain associated
  with `iteration/bug-hunt-1`.

## Index uniqueness and non-concurrency

- The developer added a `GraphWalk` precondition intended to leave enough
  `int32` index space for one `NewItem` call per input element, including
  items later elided by `GraphStep`.
- The bound was discussed in terms of the entry counter and `len(input)`;
  use of `atomic.Int32.Load()` and compatible integer arithmetic was
  recommended. Symbolic proof still needs a model for `Load()`, and the bound
  relies on no concurrent counter changes during `GraphWalk`.
- The developer added the source note `/* Note: unsafe for concurrent
  execution */` on `PriceSpanSet`. The conversation distinguished this
  documentation contract from synchronization: the atomic counter does not
  make the tree and graph safe for concurrent use.

## Purchasability goals

- The developer introduced a `Purchasable(items []*Item)` function. It
  returns false for an empty slice, true for a singleton, otherwise sorts
  the slice in place by index and checks that adjacent `Price.Base` values
  strictly increase.
- Created `model/proof/purchasable-goals.md` with readable mathematical
  definitions for:
  - the final `Item.out` graph and directed paths;
  - `Purchasable(S)` as a pairwise index/base-price ordering condition;
  - the goal that every subset of size at least two drawn from a directed
    path is purchasable, with an equivalent pairwise path condition;
  - a possible incomparable-node/non-purchasability statement, explicitly
    noting that non-reachability alone does not logically imply
    non-purchasability.
- The developer clarified the intended idea: every pair of purchasable items
  should fall together on some directed `Item.out` path, and that relation
  is itself part of the proof goal. This is not yet a proven property. It
  differs from assuming that roots reach every vertex: common reachability
  from roots gives a vertex universe but does not ensure every pair shares a
  directed path.
- The model-level vertex universe remains to be finalized. A vector of roots
  can define it as the closure of `Item.out` reachability from those roots,
  provided that closure covers every instantiated item, including elided
  items. This would avoid requiring a runtime collection for every
  mathematical subset.
- Further definition work is pending before attempting a proof. In
  particular, specify exactly whether the pair claim is:
  - for every pair `a,b` in the final universe, `Purchasable({a,b})`
    implies that `a` and `b` lie together on a directed path; or
  - an equivalent characterization involving purchasability of every
    size-at-least-two subset of each path.
  Also define whether a path may have length zero, how the root closure is
  represented, and how the in-place sorting behavior of `Purchasable` is
  handled in a proof.

## Current status

- Refreshed results: link freshness, `Item.out` acyclicity, and reachable
  insert-precondition proofs, all conditional on documented premises.
- Historical branch-specific reports remain explicitly categorized in
  `model/doc/model-check-results.md`.
- The purchasability statements are definitions/goals only; no
  `Purchasable` postcondition or proof has yet been added.
- Next step: finish selecting the exact quantifiers and root/universe
  abstraction, then revise `purchasable-goals.md` to match that agreed
  statement before attempting proof or counterexample generation.

## Twin-chain and `Item.Join` follow-up

- Rechecking `linked-item-fresh.md` after the twin-link loop changed found
  that freshness of the current destination alone was insufficient: every
  source reached by following `twin` also had to be shown distinct from the
  current item.
- The freshness proof was updated with a uniform-`Price.Base` invariant for
  twin chains. Under the current `AllIntersections` and `Find` aliases, the
  two `Item.Join` call paths join equal-base items. In the twin-link loop,
  the match base is strictly less than the current item's base, so no
  reached twin can be the current item. This establishes the `Item.Link`
  source-distinctness obligation at that loop, conditional on the stated
  identity and alias premises.
- A separate twin-chain acyclicity goal is necessary for proving that
  twin traversals terminate. A prior `Item.Join` contract attempt was not
  provable; its precondition failed at a reachable `MatchTruncate` call in
  the then-current behavior.
- `GraphStep#FauxMatch` was subsequently added to skip a match whose bound
  equals the current base. Under the present annotations this removes the
  direct self-join route in `MatchTruncate`: a non-skipped match has
  `match.Bound > at`, and `Find(match.Base, at)` occurs before mutation, so
  it cannot return that same match. The other join with the fresh current
  item is also not a direct self-join. This does not establish the general
  `Item.Join` precondition or twin-chain acyclicity.
- `PriceSpanSet.roots` does not, by itself, prove the `Item.Join`
  precondition. Root coverage does not rule out an argument that already
  has a non-nil twin or establish that a splice preserves an acyclic chain.

## Abstract-model BNF and proof status

- The symbolic-abstraction grammar now declares
  `<abstract-model-annotation>` using `<abstract-model-name>`, optional
  relation punctuation, a newline, and multiline `<go+>` content. The
  `dag.go` declarations `PriceBounds` and `InputSizeBounds` use multiline
  bodies, and `GraphWalk` refers to both names as preconditions.
- The naming rules in
  `model/meta/symbolic-abstraction-rules.md` specify that `*-name`
  nonterminals stand for freely chosen identifier spellings: declarations
  introduce fresh names and later occurrences are references to those
  declarations. They also clarify that ordinary BNF cannot express this
  cross-occurrence freshness/resolution constraint by itself.
- A parsing attempt accepted the targeted abstract-model declaration
  shape and name references after the syntax update, but did not confirm
  whole-grammar coherence: `<element-id>` remains undefined in the visible
  BNF, and the rules do not yet define how free variables in a named model
  body are bound at use sites.
- The recommended direction for resolving the latter issue was explicit
  model parameters and argument substitutions (e.g.
  `PriceBounds(container, sequence)` referenced as
  `PriceBounds(set, input)`), rather than implicitly capturing every
  identifier in scope. This makes dependencies, arity, and types explicit.
  The developer agreed callers are already handled by Go call semantics
  and deferred the design decision for later.
- No proofs of `PriceBounds` or `InputSizeBounds` were completed. The
  immediate blocker is finalizing the grammar/model binding semantics; after
  that, each model's guarantees and caller obligations can be checked.

## Resume point

- Resume by discussing whether/how to parameterize abstract-model
  declarations and references, including arity and type rules, while
  retaining Go's ordinary function-call semantics for callers.
- Once the syntax and name/binding rules are agreed, re-parse the
  annotations and assess whole-BNF coherence before attempting the two
  named precondition proofs.
- Keep the twin-chain acyclicity obligation distinct from direct
  self-join avoidance; neither roots coverage nor `FauxMatch` alone proves
  general acyclicity.
