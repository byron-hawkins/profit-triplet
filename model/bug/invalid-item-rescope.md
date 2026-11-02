# Bug: `AllIntersections` permits a match at the truncate boundary

## Summary

`PriceSpanSet.ItemTruncate#Insert` requires:

```go
item.Price.Bound > item.Price.Base
```

In `GraphStep`, the call is reached when:

```go
callerItem.Price.Base < match.Price.Bound
```

The current `AllIntersections` postcondition permits
`match.Price.Base == callerItem.Price.Base` because it requires only
`match.Price.Base <= a`. When that equality holds, `ItemTruncate` assigns
`match.Price.Bound = callerItem.Price.Base`, making the match's bound equal
its base. The truncate insertion precondition then fails.

The fact that items satisfy the `GraphWalk#InsertItem` precondition when first
inserted does not rule this out: `ItemTruncate` mutates the already-inserted
object through its pointer before reaching its own insertion precondition.

**Outcome:** REFUTED in the annotated abstraction.

## Counterexample

Use three inputs, all satisfying `GraphWalk`'s price constraint:

```text
first.Price  = 20
second.Price = 30
third.Price  = 20
```

The first input is inserted with `(Base, Bound) = (20, math.MaxInt32)`.
Choose the empty-result case for the second input's query, so the second is
inserted with `(30, math.MaxInt32)` and the first remains unchanged.

For the third input, the query point is `a = 20`. The nonempty
`AllIntersections` case can return the first item: its base satisfies the
postcondition `20 <= a`, and its bound is `math.MaxInt32`. The
`GraphStep#MatchBoundCheck` guard `20 < math.MaxInt32` holds. Truncation then
sets the first item's bound to `20`, equal to its base. The required
precondition `Bound > Base` becomes false.

## Annotated-line value-range execution

| Annotated line | Path condition and value ranges |
|---|---|
| `PriceSpanSet.GraphWalk()` precondition | Prices `[20, 30, 20]` are each in `[1, 9999]`. |
| `GraphWalk#AddItem` / `Item#constructor` (first item) | Constructor gives `first.Price = (Base: 20, Bound: math.MaxInt32)`. |
| `GraphWalk#InsertItem` (first item) | `20 < math.MaxInt32`; precondition holds and the history contains `first`. |
| `GraphWalk#AddItem` / `Item#constructor` (second item) | Constructor gives `second.Price = (Base: 30, Bound: math.MaxInt32)`. |
| `GraphStep#TreeQuery` (second item) | Select `%values[1]` with `count = 0`, `matched = false`; the empty slice satisfies the query postcondition vacuously. |
| `GraphWalk#InsertItem` (second item) | `30 < math.MaxInt32`; precondition holds and the history contains `[first, second]`. |
| `GraphWalk#AddItem` / `Item#constructor` (third item) | Constructor gives caller `item.Price = (Base: 20, Bound: math.MaxInt32)`. |
| `GraphStep#TreeQuery` (third item) | Select `%values[0]` with `i = 0`, `count = 1`, `matched = true`. With history length two, `0 >= 0`, `0 < 2 - 1`, and `1 <= 2`. The result is `[first]`; its base `20` satisfies the postcondition `first.Price.Base <= a` for `a = 20`. |
| `GraphStep#MatchBoundCheck` | `callerItem.Price.Base < match.Price.Bound` is `20 < math.MaxInt32`, so the call is reached. |
| `GraphStep#MatchTruncate: call` | Expand `ItemTruncate(first, 20)`. In the callee, `item.Price.Base = 20`, `item.Price.Bound = math.MaxInt32`, and `at = 20`. |
| `ItemTruncate#Truncate` | Apply the annotated assignment `item.Price.Bound = at`. The already-inserted `first` now has `Base = 20`, `Bound = 20`. |
| `ItemTruncate#Insert` precondition | Required: `item.Price.Bound > item.Price.Base`, i.e. `20 > 20`. False; the precondition is refuted. |

## Impact and possible correction

The `AllIntersections` postcondition's non-strict comparison admits a match
whose base is exactly the query point. To prove this truncate precondition
from the current branch guard, the summary must rule out that equality for
results that reach this branch, for example by requiring
`i.Price.Base < a` in the relevant query case. Alternatively, the caller or
truncate contract must explicitly handle the equal-base case without
performing a zero-width truncation.
