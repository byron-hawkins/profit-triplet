# Proof: `GraphStep` links only to a fresh item

## Claim

For each iteration of `PriceSpanSet.GraphWalk` that reaches
`PriceSpanSet.GraphWalk#Step`, the call to `GraphStep(&item)` satisfies the
precondition of `PriceSpanSet.GraphStep()`. During that call:

1. `item` is absent from `set.tree.EverInserted`.
2. Every `match` returned by `AllIntersections` differs from `item`.
3. Every `predecessor` yielded by `match.AllIn` differs from `item`.
4. `item.out` is empty at each call to `Item.Link`.

Consequently, both `GraphStep#MatchLink` and
`GraphBranchSamePrice#call` satisfy `Item.Link()`'s precondition:
`to.out.Empty() && from != to`.

This is a proof in the project's annotated abstraction. It does not model the
concrete implementation of `interval.SearchTree` or `vector.Vector`.

## Scope and premises

- The `PriceSpanSet` is created using the annotated `NewPriceSpanSet`, whose
  alias initializes the abstract `EverInserted` history to an empty slice.
- Execution follows `GraphWalk`'s iteration order: `GraphWalk#AddItem`,
  `GraphWalk#Step`, then `GraphWalk#InsertItem`.
- Each input price is less than `math.MaxInt32`.
- Item addresses follow Go pointer identity: the address of the current
  non-zero-sized local `item` is distinct from pointers to items retained from
  previous iterations.
- The abstract summaries in `dag.go` define Vector operations,
  `AllIntersections`, and tree history. A call marked `call` applies its
  annotated callee semantics; calls not so marked are no-ops, as specified by
  `proof-generation.md`.

The argument concerns every graph state reached, including a prefix ending
in panic. It does not require `GraphWalk` to complete normally.

## Lemma 1: the abstract history contains only earlier iteration items

Initially, `EverInserted` is empty by the `NewPriceSpanSet` alias.

`Item#constructor` guarantees `item.Price.Base == i.Price` and
`item.Price.Bound == math.MaxInt32`. `GraphStep` does not mutate the current
item's price: `GraphStep#MatchTruncate` has no `call` label and is therefore a
no-op, and the current item is distinct from each match. Thus, for each input
price less than `math.MaxInt32`, the `GraphWalk#InsertItem` precondition
`item.Price.Base < item.Price.Bound` holds.

In `GraphWalk`, the history update is attached to
`GraphWalk#InsertItem`:

```go
set.tree.EverInserted = append(set.tree.EverInserted, &item)
```

This occurs after `GraphWalk#Step`. Thus, at the beginning of each
`GraphStep(&item)`, the history contains only pointers appended by earlier
iterations and does not yet contain `&item`.

The `GraphStep()` precondition states:

```go
!slices.Contains(set.tree.EverInserted, item)
```

The preceding order and history update establish this condition. The current
item's address is distinct from retained addresses of prior local items, so it
cannot alias a prior history entry.

## Lemma 2: query matches differ from the current item

`GraphStep#TreeQuery` invokes the annotated
`interval.SearchTree.AllIntersections` summary. Its alias returns a slice of
`tree.EverInserted`:

```go
tree.EverInserted[%i, %i + %count]
```

For every applicable substitution, every element of this slice is therefore
an element of `EverInserted`. By Lemma 1, `item` is not in that history.
Consequently, each `match` in the slice is distinct from `item`.

For the empty-history case, `%values[1]` permits `%i == 0`, `count == 0`,
and `matched == false`; the query returns an empty slice, so there is no
`match` to check. For nonempty history, the applicable query result remains a
slice of history entries, which is all this freshness argument requires.

## Lemma 3: predecessors differ from the current item

Maintain this graph/history invariant at the start of each `GraphStep`:

> Every item reference in any existing `Item.in` or `Item.out` vector refers
> to an item in `EverInserted`.

It holds initially because the history and graph are empty. During
`GraphStep`, the only graph updates are the two annotated `Item.Link` calls.
Each adds the current item to a source's `out` and adds that source to the
current item's `in`. `GraphWalk#InsertItem` then appends the current item to
`EverInserted`, restoring the invariant before the next iteration. If an
execution panics before that append, it has no next iteration to invalidate
the argument.

`AllIn`'s annotated body iterates from `item.in.Size()` and passes
`item.in.At(i)` to its callback. The `Size` and `At` aliases operate over
the abstract vector's `elements`; hence each yielded predecessor is an
existing reference in `match.in`. By the invariant, that predecessor is in
`EverInserted`. By Lemma 1, the current item is not in `EverInserted`, so
the predecessor differs from `item`.

## Lemma 4: `item.out` stays empty during `GraphStep`

`GraphWalk#AddItem` obtains the current value from `Item#constructor`. The
constructor postcondition gives:

```go
return Item{
    Price: PriceSpan{ Base: price, Bound: math.MaxInt32 },
    out: vector.NewVector[*Item](),
}
```

The `NewVector` alias constructs the modeled vector with an empty `elements`
slice, and `Vector.Empty` defines emptiness as `len(v.elements) == 0`.
Therefore the current item's `out` is empty before `GraphStep`.

In either `Item.Link` call, the precondition requires `to.out.Empty()`.
Its annotated updates append `to` to `from.out` and `from` to `to.in`;
neither update changes `to.out`. The only `GraphStep` calls pass the current
item as `to`. Thus `item.out` remains empty before every such call, including
multiple calls during the match and predecessor loops.

## Conclusion

At `GraphStep#MatchLink`, Lemma 2 gives `match != item`; at
`GraphBranchSamePrice`, Lemma 3 gives `predecessor != item`. Lemma 4 gives
`item.out.Empty()` at both sites. Therefore the `Item.Link()` precondition
holds at both call sites.

The premises are sufficient for the local acyclicity step: adding an edge
`from → item` cannot create a cycle when `item.out` is empty and `from` is
not `item`. This proof establishes the freshness and empty-target obligations
for `GraphStep`; the complete DAG theorem additionally requires that no
other code mutates graph edges outside the annotated `Item.Link` operations
and that the graph is initially acyclic.

## Revalidation checklist

- `NewPriceSpanSet` still initializes `EverInserted` empty.
- `GraphWalk#Step` still occurs before `GraphWalk#InsertItem`.
- The insertion alias still appends the current item after `GraphStep`.
- `GraphStep()` retains its `!slices.Contains(EverInserted, item)` precondition.
- `AllIntersections` still returns only a slice of `EverInserted`.
- `AllIn` still yields only `in.At(i)` for indices in `in.Size()`.
- `Item#constructor` still guarantees an empty `out`.
- `Item#constructor` still sets `Price.Base` to the input price and `Price.Bound`
  to `math.MaxInt32`.
- `Item.Link` still changes only `from.out` and `to.in`.
- No additional graph mutation path is introduced.
