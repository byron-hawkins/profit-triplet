# Project conversation history — abstract collections

Brief summary of the November 27, 2026 discussion. For current syntax and
rules, consult [`symbolic-abstraction.bnf`](../model/meta/symbolic-abstraction.bnf)
and [`symbolic-abstraction-rules.md`](../model/meta/symbolic-abstraction-rules.md).

## Abstract collection (AC) annotations

- The BNF now distinguishes abstract sets and lists, supports collection
  instantiation annotations, and includes `\` as a state operator. An
  integrity review identified a malformed backslash literal and the
  example singleton `{i}` not being a Go expression; these were corrected
  in the grammar/source. The latest BNF check found no remaining issue in
  these additions.
- `@=` declares an AC associated with each instance of the annotated Go
  type. `@==` associates a struct field with the AC as its concrete
  instantiation. Placement requirements for these annotations are in the
  semantic rules.
- Intended AC state begins empty unless an ACI supplies its initial value.
  Set correspondence ignores order and duplicates; list correspondence
  preserves sequence order and multiplicity. ACIs are optional.
- State annotations are interpreted as updates of the form
  `AC = AC operator operand`. Operand kind is determined from its fully
  resolved static type, not guessed from runtime behavior:
  compatible elements act as singletons; supported arrays, slices, or
  explicitly modeled vectors act as collections. Maps and other aggregate
  types are excluded. Unresolved types are insufficiently defined;
  resolved but unsupported/incompatible types are rejected.
- Set operators use set semantics. For lists, union appends, intersection
  retains the longest common prefix, and removal removes every occurrence
  of each operand element. Element equality uses Go `==`, requiring
  comparable element types.
- The semantics were added to `symbolic-abstraction-rules.md`; no grammar
  change was needed for that documentation update.

## `dag.go` instance universe

The current example declares `I` as a set of `*Item` on `PriceSpanSet`,
associates the `all` field with it, and annotates `NewItem`'s append as
`I ∪ {i}`. Under the AC semantics, this defines the per-instance universe
of items constructed through `NewItem`, including repeated `GraphWalk`
calls on the same instance. Closure of `Item.twin` within that universe is
a separate invariant, not implied by collection membership.

No proof of tree/twin exclusivity or implementation conformance was
undertaken in this discussion.
