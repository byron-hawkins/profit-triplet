# Proof: `Item.out` is acyclic during `GraphWalk`

## Theorem

Starting from a `PriceSpanSet` created by `NewPriceSpanSet`, every `Item.out`
graph state reached while executing `GraphWalk` is acyclic, assuming the
freshness facts in [linked-item-fresh.md](./linked-item-fresh.md) and the
annotation semantics in [proof-generation.md](./proof-generation.md).

Here, the graph has an edge `u → v` exactly when `v` occurs in
`u.out.elements`. The claim covers each individual `out` update, including
states between the two updates in `Item.Link`, and does not require normal
completion of `GraphWalk`.

## Assumptions and scope

- The set starts with `NewPriceSpanSet`; its modeled history and graph are
  empty.
- `GraphWalk` is the graph-building operation under consideration. There is
  no concurrent mutation and no other code mutates an `Item.out` vector.
- The freshness proof establishes, at every `Item.Link` call reached from
  `GraphStep`, that the current item is distinct from the source and has an
  empty `out` vector. It also establishes the history and predecessor facts
  used to justify those call-site conditions.
- Annotated calls and substitutions have the meanings specified in
  `proof-generation.md`. In particular, the `PushBack` calls in `Item.Link`
  append their arguments, and an unannotated call is a no-op.

This is a property of the annotated abstraction. It does not prove that the
concrete interval-tree or vector implementations refine that abstraction.
Nor does it claim that `GraphWalk` completes without panic.

## Edge-addition lemma

Let `G` be an acyclic directed graph, and let `t` be a vertex whose outgoing
edge set is empty. If `s != t`, adding the edge `s → t` preserves acyclicity.

Any newly created cycle would have to contain the new edge. Removing that
edge from such a cycle would leave a directed path from `t` back to `s`.
There is no such path: since `t` has no outgoing edges, no nonempty path can
start at `t`; and since `s != t`, the zero-length path at `t` does not end at
`s`. Thus no new cycle is created.

Both conditions matter. In particular, if `s == t`, the new edge is itself a
cycle even when `t.out` was empty.

## Induction over graph updates

Use the number of `Item.out` append operations as the induction measure.

### Base case

`NewPriceSpanSet` initializes an empty item history. There have been no
`Item.Link` calls and therefore no `out` edges. The graph is acyclic.

### Inductive step

Assume the graph is acyclic immediately before the next `Item.out` append.
The only graph-edge mutations in `GraphStep` are the `Item.Link` calls at
`GraphStep#MatchLink` and `GraphStep#GraphBranchSamePrice`. The freshness
proof establishes for each such call:

1. `to` is the newly created current item;
2. `to.out` is empty at the call; and
3. `from != to` (`from` is the match or the yielded predecessor).

`Item.Link` first appends `to` to `from.out`. Apply the edge-addition lemma
with `s = from` and `t = to`; acyclicity is preserved after this append.
The second `PushBack` appends `from` to `to.in`. It does not change any
`out` vector, so it leaves the `Item.out` graph unchanged and acyclic.

The current item remains a valid empty-out target for any later link in the
same `GraphStep`: each `Item.Link` changes its `in`, not its `out`. Once
`GraphStep` finishes, the item is inserted into the tree/history before the
next `GraphWalk` iteration. The next iteration creates a distinct new item,
and the same argument applies. If execution panics, there is no subsequent
iteration; any graph prefix reached before the panic is covered by the
induction.

Therefore every finite prefix of the `out` append sequence is acyclic, which
proves the theorem.

## Conclusion and limits

Taking the freshness proof as established, acyclicity of the modeled
`Item.out` graph follows by induction from the empty initial graph. This
proves the graph property for `GraphWalk` under the assumptions above; it
does not establish panic-freedom, correctness of interval queries, or
acyclicity under arbitrary external mutations of `Item.out`.
