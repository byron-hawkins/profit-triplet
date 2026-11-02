# Proof: `GraphWalk#InsertItem` precondition holds

## Summary

The precondition on `PriceSpanSet.GraphWalk#InsertItem` is:

```go
item.Price.Base < item.Price.Bound
```

The `GraphWalk` precondition requires every input price to be in `1..9999`.
`NewItem` initializes the current item's price to
`PriceSpan{Base: i.Price, Bound: math.MaxInt32}`. During `GraphStep`, the
`MatchTruncate` alias changes `match.Price.Bound`, not the current item's
price. The established freshness fact `match != item` ensures those are
distinct objects. No other annotated operation on this path changes the
current item's price. Therefore, whenever execution reaches `InsertItem`,
its precondition holds.

**Outcome:** PROVED for the `GraphWalk#InsertItem` price inequality on paths
where `GraphStep` returns normally, assuming the established freshness fact
that every query match differs from the current item and the annotation
semantics in [`proof-generation.md`](../meta/proof-generation.md).

This does not prove that every call made inside `GraphStep` satisfies its own
preconditions. In particular, see
[`invalid-item-rescope.md`](../bug/invalid-item-rescope.md) for a counterexample
to `ItemTruncate#Insert`'s precondition. Thus this result is not a proof that
all of `GraphWalk` is free of invalid calls.

## Input-range invariant

The `GraphWalk` precondition is the boolean expression:

```go
func() bool {
    for _, i := range input {
        if i.Price <= 0 || i.Price >= 10000 {
            return false
        }
    }
    return true
}()
```

At the start of each iteration over `input`, use this loop invariant:

> Every input element in the already-processed prefix has a price `p` with
> `0 < p < 10000`.

The invariant holds for the empty prefix. An iteration returns `false` if
the current element violates the range; otherwise it preserves the invariant
for the longer prefix. Thus, if the precondition evaluates to `true`, every
input element has `1 <= i.Price <= 9999`. For empty input, `GraphWalk` has no
iteration and never reaches `InsertItem`.

## Annotated-line value-range execution

Consider any iteration that reaches `GraphWalk#InsertItem`. The trace follows
the annotated source lines and applies the explicit `MatchTruncate` alias.

| Annotated line | Path condition and value ranges |
|---|---|
| `PriceSpanSet.GraphWalk()` precondition | The boolean expression evaluates to `true`; therefore every `i` in `input` satisfies `1 <= i.Price <= 9999`. |
| `PriceSpanSet.GraphWalk#EachInput` | For the selected element, `1 <= i.Price <= 9999`. |
| `GraphWalk#AddItem` → `Item#constructor` | The constructor postcondition gives `item.Price.Base = i.Price` and `item.Price.Bound = math.MaxInt32`. Hence `1 <= item.Price.Base <= 9999` and `item.Price.Bound = 2147483647`. |
| `PriceSpanSet.GraphWalk#Step` → `GraphStep` precondition | `GraphStep`'s freshness precondition holds for the newly created item. The price range remains unchanged. |
| `PriceSpanSet.GraphStep#TreeQuery` | The query returns the modeled history slice and match flag. It does not change `item.Price`. |
| `PriceSpanSet.GraphStep#TreeQueryMatchCheck` and `#EachMatch` | On a path with a match, the established freshness fact gives `match != item`. If there are no matches, no truncation occurs. |
| `PriceSpanSet.GraphStep#MatchBoundCheck` | The true branch is guarded by `item.Price.Base < match.Price.Bound`; the false branch does not execute `MatchTruncate`. Neither branch itself changes `item.Price`. |
| `PriceSpanSet.GraphStep#MatchTruncate` | Apply the annotation alias exactly: `match.Price.Bound = item.Price.Base`. Its left-hand side is a field of `match`. Since `match != item`, it does not change `item.Price.Base` or `item.Price.Bound`. The current item's values remain `Base = i.Price` and `Bound = math.MaxInt32`. |
| `PriceSpanSet.GraphStep#MatchReplaceErrorCheck` | If the error branch panics, the execution does not reach `InsertItem`. On a continuing path, the current item's price remains unchanged. |
| `PriceSpanSet.GraphStep#MatchLink` / `#GraphBranchSamePrice` | `Item.Link` changes the annotated `in` and `out` vectors, not the current item's price. The current item's price remains `Base in [1,9999]`, `Bound = 2147483647`. |
| `PriceSpanSet.GraphWalk#InsertItem` | `item.Price.Base <= 9999 < 2147483647 = item.Price.Bound`; therefore `item.Price.Base < item.Price.Bound`. |

The same range argument applies whether `GraphStep` has zero or multiple
matches, provided `GraphStep` returns normally. A path that panics inside
`GraphStep` does not reach the insertion obligation. A path where an internal
call precondition fails is not a valid normally returning proof path; that
call-site obligation must be handled separately.

## Dependencies and scope

- The input loop in the `GraphWalk` precondition has the invariant stated
  above; the proof uses the true outcome of that boolean expression.
- `Item#constructor` has the postcondition currently specified in `dag.go`.
- The established freshness fact applies to the current annotations and
  guarantees `match != item` at `MatchTruncate`.
- The `MatchTruncate` alias is applied as a state update to
  `match.Price.Bound`; it is not treated as a no-op.
- This proof does not discharge `ItemTruncate#Insert`'s precondition; that
  separate obligation is refuted in
  [`invalid-item-rescope.md`](../bug/invalid-item-rescope.md).
- This proves the annotated precondition for executions satisfying the
  `GraphWalk` precondition and reaching its insertion point after normal
  return from `GraphStep`. It does not prove that arbitrary callers satisfy
  that precondition, nor does it prove refinement of the external tree
  implementation.
