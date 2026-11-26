# Project conversation history — recent induction work

This is a focused summary of the recent conversation, approximately the last
four days, recorded November 25, 2026. Exact dates of individual discussions
are not available here. Earlier background is in the existing dated history
files in this directory. The source and annotations have been edited
repeatedly; observations below refer to the revision discussed, not
necessarily to every later revision.

## Annotation grammar and model references

- The annotation BNF in `model/meta/symbolic-abstraction.bnf` now supports
  optional induction goal tags on abstract models and invariants:
  `[<induction-goal-tag>]`, where the tag is an alphanumeric identifier.
  Invariants have a mode, `(?)` or `(!)`, followed by `pre-condition` or
  `post-condition` and an optional tag.
- An abstract model can also specify an equivalence set, for example
  `TwinPricesUniquelyChained[TPUC](set *PriceSpanSet) <=> {TUP,TCSP,TCU}`.
  The intended role of the listed tags is described below; the syntax alone
  does not implement a proof.
- Indentation distinguishes invariants attached to the enclosing annotation
  from those attached to a substitution. The BNF's `<newline>` includes
  the comment decoration `* `; summary invariants require two additional
  spaces and substitution invariants require four additional spaces.
  Source whitespace before the `*` is allowed.
- A standalone abstract-model invocation in an invariant or model alias
  uses its name without `@`. A call inside a Go-like fragment uses `@`,
  such as `@TwinChainSamePrice(node)`. Model references resolve within the
  same source file, and the static abstract-model call graph must be
  checked for cycles before proof progress.
- Go-expression arguments must be in scope and side-effect-free, with
  pointer dereferences and collection indexing justified on every relevant
  value-flow path. Failures to resolve value flow are not proofs.
  `model/meta/symbolic-abstraction-rules.md` records these rules.

## Loop conditions and induction intent

- The user specified that a `loophead` precondition is evaluated once
  before loop entry and its postcondition once at each loop exit, rather
  than implicitly on every iteration. The assistant added this rule to
  `model/meta/symbolic-abstraction-rules.md` and linked it from
  `model/meta/proof-generation.md`. The recorded rule includes exits via
  `break`, `return`, and panic; the details of evaluating conditions on
  panic and with scoped variables may need further clarification.
- The user's intended per-iteration inductive conditions are instead placed
  inside the loop, usually on its first statement. An entry/exit condition
  alone is not a loop invariant.
- The user proposed the following proof discipline:
  1. An `(!)pre-condition` refers to an explicitly established, logically
     corresponding earlier `(?)` pre- or postcondition and fails if that
     fact was not established or no longer holds.
  2. An `(!)post-condition` refers to a corresponding `(?)pre-condition`,
     usually, perhaps necessarily, on the same statement, and fails if the
     precondition cannot be found or does not continue to hold.
  3. A `(?)` condition establishes its fact by analyzing possible execution
     paths up to the annotated point, respecting loop-head rules. A `(*)`
     mode was suggested for loop internals if needed, but is not in the BNF.
  4. Conditions carrying an induction tag are steps for the abstract model
     carrying the same tag. The proof should fail if those steps cannot
     establish the model, even if another proof exists; ideally, missing
     steps could be suggested.
  5. A model with an equivalence set should be proved strictly from the
     listed sub-goals, with comparable failure reporting.
- These are the user's proposed semantics, not a fully specified or
  implemented matching and proof system. Questions such as logical
  correspondence, ambiguity, control-flow provenance, and preservation
  still need precise definitions.

## Model and graph discussions

- Reviews covered the search-tree operation summaries, twin chains,
  truncation, item identity and graph traversal. `NewItem` was changed to
  return a pointer and record that same object in `set.all`; pointer receiver
  methods can be called on an addressable stack value without writing
  `&stack.push(...)`. The model's statically sized trio has exactly three
  elements, so accesses to its three corresponding slice positions do not
  need an additional length premise.
- A package imported only for use in comments is unused by Go. The user
  raised the related performance concern that modeling a FIFO queue by
  repeatedly shifting a slice could make symbolic checking expensive; a
  queue model should preserve the intended FIFO behavior without assuming
  that Go accepts a comment-only import.
- Reachability was renamed to `PurchaseStepwiseReachable` and
  `ProxyPurchaseStepwiseReachable` to express finding a path over properly
  linked purchasable steps. A found path is sufficient; unrelated paths
  need not be checked. Global graph well-formedness is a separate goal, and
  `GreaterRootsGraphUnreachable` is intended to depend on it. A greedy
  search still needs a completeness argument under that well-formedness
  property before failure can establish that no suitable path exists.
- Abstract external data models and operation summaries are the current
  proof interface; they are not automatically proofs of the concrete
  library implementations. Insufficient premises, unsupported semantics,
  or unresolved safety obligations should yield `UNKNOWN` or a
  specification error, not an assumed proof. An abstract counterexample
  must not be called reachable from valid inputs without a reachability
  argument.

## Proof-plan discussion

The user asked whether a parser and symbolic checker could produce a
source-linked intermediate file that, together with the Go source, gives a
complete proof plan for tagged models, potentially consumable by an
open-source proof assistant. Why3 and Rocq were discussed as possible
backends. A credible translation would need explicit obligations,
source locations, control-flow and loop coverage, provenance, external-model
assumptions, and a clearly stated translator trust boundary. No such
translator, proof assistant integration, or machine-checked proof was
created.

## Current experiment: `TreeUniquePrices` (`TUP`)

The user requested this gated procedure:

1. Verify every `(?)` condition tagged `TUP`.
2. Only if all succeed, identify a minimal logically corresponding set of
   `(?)` conditions for each `(!)` condition; fail on ambiguity.
3. Only if all correspondences are identified, verify that each fact
   remains valid at the `(!)` location.
4. Only then prove the `TreeUniquePrices(set)` conditions in `GraphWalk`:
   its loop-entry `(?)` precondition, the inner `(!)` precondition across
   iterations using the tagged steps, and the loop-exit `(!)` postcondition.

The last response attempted the first gate informally. It treated the tree
insert membership postcondition and `HasAlterEgo(match, item)` after the
twin assignment as locally established, conditional on their operations
and relevant premises. It left the elided item's tree exclusion, merged
chain tree exclusion, and `GraphStep` return condition `UNKNOWN`, because
the needed preservation and cross-iteration facts were not established.
These are not machine-checked results, and `UNKNOWN` is not a counterexample.
The later gates were not performed.

The previous response was **not a complete inventory** of the tagged
conditions: it omitted the `MatchProxy#Insert` precondition
`!set.Contains(match)` already present at that point. Moreover, the source
has since changed: a fresh inventory on November 25 also shows additional
`TUP` conditions at `GraphStep#TreeQuery`, `GraphStep#SpliceTwinsLeft`, and
`MatchProxy#Insert`. The last includes `match != item` inside `MatchProxy`,
whose parameters are named `match` and `proxy`; its identifier scope needs
checking before attempting the proof. Re-read the current source rather
than treating either inventory as exhaustive.

The last response also reported insufficient indentation on three
`GraphWalk` conditions. The current source subsequently shows the required
extra spaces on those lines; recheck the actual revision rather than
repeating the old finding. The earlier conditional proof claims should be
reassessed only after checking syntax, static model-call cycles, safety and
operation preconditions, normal versus exceptional exits, and all current
tagged conditions. A passing Go test does not validate annotation comments.

## Next step

Re-read the current BNF, semantic rules, and `dag.go`, inventory all `TUP`
conditions, resolve syntax and scope blockers, and retry the user's first
gate. Distinguish proved, refuted, and unknown obligations, and stop before
the later gates unless the preceding gate actually succeeds.
