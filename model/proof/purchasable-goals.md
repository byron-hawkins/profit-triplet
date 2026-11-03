# Formal goals for purchasability

This document defines candidate post-state properties for `GraphWalk`. These
are proof goals, not claims that the current implementation establishes them.

## Abstract final graph

Fix one `PriceSpanSet` and a state reached after `GraphWalk` returns normally.
Model its `Item.out` graph as `G = (V, E)`, where:

- `V` is the finite set of `Item` identities belonging to this graph;
- `E` contains `(u, v)` exactly when `v` occurs in `u.out.elements`.

The vertex universe `V` is a model-level definition. It must include every
item intended to count as a graph node, including an item elided from the
tree by `GraphStep` if it is still considered part of the graph. The current
Go fields do not provide a canonical enumeration of all such identities.
Therefore a proof must define a projection/ghost universe for `V`; using
`tree.EverInserted` as `V` is valid only if the intended universe is exactly
that history.

Define `u ↝ v` to mean there is a directed `out` path from `u` to `v`.
Paths are finite sequences `P = (v₀, …, vₖ)` with `k ≥ 0` and
`(vᵢ, vᵢ₊₁) ∈ E` for each `0 ≤ i < k`. Let `V(P)` be the set of vertices in
that sequence. Since the graph is intended to be a DAG, a path cannot repeat
a vertex.

## Purchasable predicate

For a finite set `S ⊆ V` with at least two distinct items, define:

> **Purchasable(S)** iff for every `a, b ∈ S`,
> `a.index < b.index` implies `a.Price.Base < b.Price.Base`.

Equivalently, no pair in `S` satisfies
`a.index < b.index` and `a.Price.Base ≥ b.Price.Base`.

The requested goals only use sets of size at least two, so they do not
depend on the current Go function's special cases where `Purchasable` returns
false for an empty slice and true for a one-element slice.

### Connection to `Purchasable(items)`

The current implementation sorts the supplied slice by `index`, then checks
that adjacent `Price.Base` values strictly increase. For a finite set whose
indices are pairwise distinct, this is equivalent to the pairwise predicate
above. The `GraphWalk` index bound is intended to prevent `int32` overflow,
but a proof of this equivalence must also establish that:

- indices are unique in the chosen vertex universe, including any items
  already present before this call; and
- sorting uses a valid strict ordering for those indices.

The function sorts its input slice in place. A proof/model that must preserve
path or subset order should evaluate it on a copy or use the mathematical
predicate above directly.

## Goal 1: every path subset is purchasable

For every normal-return final state of `GraphWalk`, the first goal is:

> For every directed path `P` in `G`, and every subset `S ⊆ V(P)` with
> `|S| ≥ 2`, `Purchasable(S)` holds.

In quantifier form:

> For every normal-return final state, for every path `P`, for every
> `S ⊆ V(P)`, if `|S| ≥ 2`, then for all `a, b ∈ S`,
> `a.index < b.index` implies `a.Price.Base < b.Price.Base`.

Because every pair `{a, b}` from a path is itself a subset of that path, this
is equivalent to checking every pair of vertices that lie together on some
path:

> For every path `P` and every `a, b ∈ V(P)`,
> `a.index < b.index` implies `a.Price.Base < b.Price.Base`.

This pairwise form is usually the more practical proof target. It avoids
enumerating all subsets: any violation in a larger subset is already a
violating two-item subset.

The scope is final states after **normal return** only. A panic execution
does not establish this postcondition and is outside this goal.

## Goal 2: non-reachability and non-purchasability

One precise interpretation of the proposed second goal is:

> For every normal-return final state and every `S ⊆ V` with `|S| ≥ 2`, if
> `S` contains two distinct incomparable vertices `a` and `b`—meaning
> neither `a ↝ b` nor `b ↝ a`—then `Purchasable(S)` is false.

This statement is **not implied by non-reachability alone**. For example,
take two disconnected vertices with:

```text
a.index = 1,  a.Price.Base = 10
b.index = 2,  b.Price.Base = 20
```

They are incomparable, but `{a, b}` is purchasable. Graph reachability
constrains which pairs occur together on paths; it does not, by itself,
constrain the index/price order of disconnected pairs.

Thus Goal 2, as stated above, is a separate theorem and is refuted unless
additional program invariants connect disconnectedness to a price-order
violation. If the intended claim is instead that non-reachable pairs are
*not required* to satisfy the path-order property, that is a statement about
proof scope, not a claim that `Purchasable(S)` must be false.

## Proof setup required

Before attempting either goal, make the model-to-program connection
explicit:

1. Define the vertex universe `V` for the final state, including the status
   of preexisting, tree-elided, and otherwise referenced items.
2. Establish that each `Item.index` uniquely identifies an item in `V`, or
   adapt the predicate/checker to handle duplicate indices.
3. Specify how the logical set `S` is represented when evaluating
   `Purchasable`; do not rely on an in-place sort when order must be
   preserved.
4. State the normal-return condition and the `GraphWalk` preconditions under
   which the post-state is considered.

These definitions make the goals precise without requiring a runtime Go
collection representing every ad-hoc subset.

## Intended course of development

The implementation is intended to maintain a vector `roots` such that every
instantiated `Item` is reachable by following `Item.out` from at least one
root. The graph should therefore contain no detached instantiated items,
including items that would otherwise be elided by `GraphStep`. In the proof
model, the universe `V` will be defined as the closure of `roots` under
`Item.out` reachability, and a separate invariant will establish that this
closure covers every instantiated item in scope.

“Fully connected from `roots`” here means coverage from the root set; it does
not mean every pair of vertices is comparable by directed reachability. The
stronger intended relationship is part of the proof goal:

> For every pair of distinct items `a, b ∈ V`, if `{a, b}` is purchasable,
> then there is a directed `Item.out` path containing both `a` and `b`.

Equivalently, every pair that does not occur together on any directed path
must fail the pairwise purchasability condition. Proving this relationship
will require invariants connecting item creation, root maintenance, graph
edges, indices, and base prices; it does not follow from roots coverage
alone.

The planned order is to first update graph construction so every instantiated
item is represented and reachable from `roots`, then state and prove the
roots-coverage invariant, and finally prove the pair-on-a-path /
purchasability relationship over the final graph after normal return from
`GraphWalk`. These are intended design and proof goals, not properties
claimed of the current implementation.
