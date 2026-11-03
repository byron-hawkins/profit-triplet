# Proof: `GraphStep` links only to a fresh item

## Claim

At each `Item.Link(from, item)` call reached during `GraphStep`:

1. `item` is distinct from `from`;
2. `item.out` is empty.

Thus the `Item.Link` precondition `to.out.Empty() && from != to` holds at
`GraphStep#MatchLink` and `GraphStep#GraphBranchSamePrice`.

The proof is for the current annotated abstraction. It does not establish
refinement of the external vector or interval-tree implementations.

## Premises

- `GraphWalk` starts with a set whose modeled tree history is empty.
- Each `GraphWalk#AddItem` produces a new object identity, distinct from all
  item objects retained from earlier iterations. This is the Go pointer
  identity/fresh-allocation premise for the address of the escaping local
  `item`.
- `GraphStep` is called before the current item is appended by
  `GraphWalk#InsertItem`.
- Summary aliases and element annotations have the meanings defined by
  `proof-generation.md`. In particular, unmarked calls are no-ops, while
  calls marked `call` use their annotated bodies or summaries.

The input-price range and `GraphWalk#InsertItem` price precondition are not
needed for the two `Item.Link` call-site obligations established here.

## Lemma 1: every tree-history entry is an earlier item object

Initially the abstract `EverInserted` history is empty.

Maintain the invariant:

> Every pointer in `EverInserted` points to an item object allocated before
> the current `GraphWalk` iteration.

There are two annotated operations that append to this history:

- `GraphWalk#InsertItem` appends the current `&item`, but only after
  `GraphStep` has returned. It therefore becomes an earlier item on the next
  iteration.
- `ItemTruncate#Insert` appends its `item` parameter. This call is reached
  with `match` as that parameter; `match` came from the `AllIntersections`
  result and is already an entry of `EverInserted`. Appending this same older
  object, including a duplicate reference, preserves the invariant.

No other annotated operation appends the current item to the history before
the current `GraphStep` completes. `GraphStep`'s new boolean result only
controls whether the later `GraphWalk#InsertItem` is reached. If it returns
false, the current object is not appended there at all.

## Lemma 2: every match is distinct from the current item

`GraphStep#TreeQuery` calls the annotated `AllIntersections` summary. Its
alias initializes `m` as empty and appends only `i` values obtained by
iterating over `tree.EverInserted`. Therefore every returned match is a
pointer already present in the history. By Lemma 1, all such pointers refer
to objects from earlier iterations. By the fresh-object premise, none is the
current item.

This reasoning applies whether the query returns an empty slice or one or
more matches. The overlap filter in the alias narrows the possible matches
but does not alter their provenance.

## Lemma 3: every yielded predecessor is distinct from the current item

`AllIn` iterates over `match.in` and yields the element at each valid vector
index. Every such reference was added by an earlier `Item.Link` call: that
call's `from` item was already an existing graph object, and its `to` item
was the then-current `GraphStep` item. Thus every predecessor is an item
object allocated before the current iteration.

The new `Item.Join` calls only change `twin` fields and do not add `in` or
`out` references. The `insert = false` path may leave an item out of tree
history, but it does not make that object's identity equal to a subsequently
allocated current item. The fresh-object premise therefore gives
`predecessor != item`, even if that predecessor was itself not inserted into
the tree.

## Lemma 4: the current item's `out` remains empty in `GraphStep`

`GraphWalk#AddItem` uses the constructor postcondition and `NewVector` alias
to establish an empty `out` vector for the current item.

Both `Item.Link` call sites use the current item as `to`. The annotated
`Item.Link` effects append to `from.out` and `to.in`; neither changes
`to.out`. The newly added `Join` call effects set `twin` references only.
`ItemTruncate`, including its tree delete, price-bound update, `Find`,
optional join, and tree insert, does not mutate any `Item.out` vector.
`GraphStep` itself adds no other `out`-vector mutation. Therefore the
current item's `out` remains empty throughout its `GraphStep`.

## Call-site obligations and conclusion

- At `GraphStep#MatchLink`, Lemma 2 gives `match != item`; Lemma 4 gives
  `item.out.Empty()`.
- At `GraphStep#GraphBranchSamePrice`, Lemma 3 gives
  `predecessor != item`; Lemma 4 gives `item.out.Empty()`.

Therefore both `Item.Link` call sites satisfy the annotated precondition.

## Scope and dependencies

This proof uses the current `AllIntersections` alias's provenance property
(results are selected only from `EverInserted`), the `AllIn` vector
iteration, the `Item.Link` vector summaries, and the fresh-object premise.
It accounts for conditional `GraphWalk#InsertItem`, tree-history append from
`ItemTruncate#Insert`, and `Item.Join`. It does not depend on either
`dag-acyclic.md` or `fresh-item-valid.md`; it is the prerequisite proof for
those results.
