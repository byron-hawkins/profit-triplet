# Proof: `GraphStep` links only to a fresh item

## Claim

At each `Item.Link(from, item)` call reached during `GraphStep`:

1. `item` is distinct from `from`;
2. `item.out` is empty.

Thus the `Item.Link` precondition `to.out.Empty() && from != to` holds at
`GraphStep#MatchLink` (including each iteration of the twin-chain loop) and
`GraphStep#GraphBranchSamePrice`.

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
needed for the `Item.Link` call-site obligations established here.

## Lemma 1: every tree-history entry is an earlier item object

Initially the abstract `EverInserted` history is empty.

Maintain the invariant:

> Every pointer in `EverInserted` points to an item object allocated before
> the current `GraphWalk` iteration.

There are two annotated operations that append to this history:

- `GraphWalk#InsertItem` appends the current `&item`, but only after
  `GraphStep` has returned. It therefore becomes an earlier item on the next
  iteration.
- `MatchTruncate#Insert` appends its `match` parameter. This call is reached
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

## Lemma 4: each twin chain has a uniform base price

Initially each item has no twin, so each twin chain contains only its root and
has a uniform `Price.Base`.

The only `Item.Join` call sites are `GraphStep#JoinSameInterval` and
`MatchTruncate#JoinSameInterval`:

- At `GraphStep#JoinSameInterval`, the query alias returns `match` only when
  `match.Price.Base <= item.Price.Base`. On this `else` branch,
  `item.Price.Base <= match.Price.Base`; therefore both bases are equal.
- At `MatchTruncate#JoinSameInterval`, `Find` returns `twin` only when
  `twin.Price.Base == match.Price.Base`, as specified by its alias. The call
  is `twin.Join(match)`, so the joined bases are equal. `MatchTruncate`
  changes only bounds before this call, not bases.

By induction, every twin link joins items with the same base. The `Join`
splice can extend or share a chain, but cannot create a path between items
with different bases; consequently every item reached by following `twin`
from a match has that match's base.

At `GraphStep#MatchLinkTwins`, the loop is entered only on the branch
`item.Price.Base > match.Price.Base`. The uniform-base invariant therefore
gives `ego.Price.Base == match.Price.Base < item.Price.Base` for every
yielded `ego`, so `ego != item`. This also accounts for a prior same-price
match in the same `GraphStep`: it can join the current item only to a chain
whose base is the current item's base, not to the lower-base chain traversed
by this loop.

## Lemma 5: the current item's `out` remains empty in `GraphStep`

`GraphWalk#AddItem` uses the constructor postcondition and `NewVector` alias
to establish an empty `out` vector for the current item.

All `Item.Link` call sites use the current item as `to`. The annotated
`Item.Link` effects append to `from.out` and `to.in`; neither changes
`to.out`. `Join` changes only twin references. `MatchTruncate`, including its
tree delete, `Find`, price-bound updates, optional join, and tree insert,
does not mutate any `Item.out` vector. `GraphStep` itself adds no other
`out`-vector mutation. Therefore the current item's `out` remains empty
throughout its `GraphStep`.

## Call-site obligations and conclusion

- At `GraphStep#MatchLink`, Lemma 4 gives `ego != item`; Lemma 5 gives
  `item.out.Empty()`.
- At `GraphStep#GraphBranchSamePrice`, Lemma 3 gives
  `predecessor != item`; Lemma 5 gives `item.out.Empty()`.

Therefore all `Item.Link` call sites satisfy the annotated precondition.

## Scope and dependencies

This proof uses the current `AllIntersections` and `Find` aliases, the
`AllIn` vector iteration, the `Item.Link` vector summaries, and the
fresh-object premise. It accounts for conditional `GraphWalk#InsertItem`,
tree-history append from `MatchTruncate#Insert`, `Item.Join`, and the
twin-chain `Item.Link` loop. It does not depend on either `dag-acyclic.md` or
`fresh-item-valid.md`; it is the prerequisite proof for those results.
