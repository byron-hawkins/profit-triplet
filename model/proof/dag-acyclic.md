# Proof: `Item.out` is acyclic during `GraphWalk`

## Theorem

Starting from a `PriceSpanSet` with no items or edges, every `Item.out` graph
state reached during `GraphWalk` is acyclic, assuming the link-freshness
result in [`linked-item-fresh.md`](./linked-item-fresh.md) and the current
annotation semantics in [`proof-generation.md`](../meta/proof-generation.md).

The graph has an edge `u → v` exactly when `v` occurs in `u.out.elements`.
The claim covers each individual `out` mutation and does not require
`GraphWalk` to complete normally.

## Assumptions and scope

- The initial modeled graph is empty.
- Item creation in each input iteration yields a fresh identity, as in
  [`linked-item-fresh.md`](./linked-item-fresh.md).
- There is no concurrent or external mutation of `Item.out`.
- The current annotations are the complete graph-mutation model. In
  particular, `Item.Link` appends to `from.out`; `Join` mutates only `twin`;
  `ItemTruncate` mutates price/tree state; and `ItemTruncate`'s history alias
  does not alter graph edges.
- A `GraphStep` result of false can skip `GraphWalk#InsertItem`; this may
  leave a graph node absent from tree history, but does not remove or add any
  edge by itself.

This is an annotated-abstraction proof, not a refinement proof of external
collections. It does not prove panic-freedom or that the interval-query model
is complete.

## Edge-addition lemma

Let `G` be an acyclic directed graph. Let `t` have no outgoing edges, and
let `s != t`. Adding the edge `s → t` preserves acyclicity.

If a new cycle existed, it would contain the new edge. Removing that edge
would leave a directed path from `t` back to `s`. No nonempty path can start
at `t` because `t.out` is empty; a zero-length path ends at `t`, not at
`s != t`. Hence there is no such cycle.

## Induction over `out` mutations

Induct on the number of `Item.out` append operations reached.

### Base

The initial graph has no edges and is acyclic.

### Step

Assume the graph is acyclic immediately before the next `out` append.
According to the current annotations, the only such append is
`Item.Link#out`, called from either `GraphStep#MatchLink` or
`GraphStep#GraphBranchSamePrice`.

The refreshed link-freshness proof establishes at each site that:

- the destination is the current item, whose `out` vector is empty;
- the source (`match` or `predecessor`) is a distinct item.

`Item.Link#out` adds exactly the edge from that source to the current item.
The edge-addition lemma preserves acyclicity. The following
`Item.Link#in` update changes no `out` vector and therefore preserves the
graph state.

The item remains a valid empty-out destination for later links in the same
`GraphStep`: each link changes its `in`, not its `out`. `JoinSameInterval`
may be reached either directly in `GraphStep` or from `ItemTruncate`, but
`Item.Join` changes only `twin` fields. `ItemTruncate`'s price updates,
tree operations, and return value do not mutate `out`. Whether `GraphStep`
returns true and causes `GraphWalk#InsertItem`, or returns false and elides
that insertion, does not affect the already-established graph state.

Thus every finite prefix of edge additions remains acyclic. The argument
also covers a prefix that ends in panic.

## Conclusion and limits

The `Item.out` graph remains acyclic throughout the modeled `GraphWalk`,
conditional on [`linked-item-fresh.md`](./linked-item-fresh.md). Conditional
tree insertion, same-interval joining, truncation, and tree-history updates
do not invalidate the edge induction because none changes an `out` vector
except the `Item.Link#out` append analyzed above.
