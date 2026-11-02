# Bug: `GraphWalk` can reach `InsertItem` with an invalid price span

## Summary

`PriceSpanSet.GraphWalk#InsertItem` requires:

```go
item.Price.Base < item.Price.Bound
```

The current annotations do not constrain input prices to be less than
`math.MaxInt32`. A one-item input with `Price == math.MaxInt32` reaches
`InsertItem` with equal `Base` and `Bound`, violating its precondition.

**Outcome:** REFUTED for unrestricted `GraphWalk` inputs.

## Counterexample

Start with a `PriceSpanSet` created by the annotated `NewPriceSpanSet` and
provide a single input item with:

```text
i.Price = math.MaxInt32 = 2147483647
```

Initially the abstract `EverInserted` history is empty. The constructor sets
the new item's `Price.Base` to the input price and `Price.Bound` to
`math.MaxInt32`. The empty-history query case returns no matches, so
`GraphStep` does not change the new item's price. At `InsertItem`, both
endpoints are `2147483647`, and the required strict inequality is false.

## Annotated-line value-range execution

The trace follows annotated lines only. Unannotated statements are not
treated as proof nodes under `proof-generation.md`.

| Annotated line | Path condition and value ranges |
|---|---|
| `PriceSpanSet.GraphWalk#EachInput` | Choose the sole input with `i.Price = 2147483647`. The loop annotation supplies no narrower price constraint. |
| `GraphWalk#AddItem` → `Item#constructor` | Constructor postcondition gives `item.Price.Base = price = 2147483647` and `item.Price.Bound = math.MaxInt32 = 2147483647`. |
| `PriceSpanSet.GraphWalk#Step` → `PriceSpanSet.GraphStep()` precondition | The abstract history is empty, so `!slices.Contains(EverInserted, item)` holds. Price ranges remain `Base = 2147483647`, `Bound = 2147483647`. |
| `PriceSpanSet.GraphStep#TreeQuery` | Select the empty-history substitution: `count = 0`, `matched = false`, and the returned match slice is empty. The item's price is unchanged. |
| `PriceSpanSet.GraphStep#TreeQueryMatchCheck` | Take the `matched == false` branch. `GraphStep` exits without changing the item's price. |
| `PriceSpanSet.GraphWalk#InsertItem` | Required condition is `2147483647 < 2147483647`, which is false. The precondition fails. |

## Impact and possible correction

For this input, the insertion operation is attempted with an invalid
zero-width span under its stated precondition. The proof cannot establish the
precondition for arbitrary inputs without an additional constraint.

One possible correction is to specify and enforce an input constraint such
as `i.Price < math.MaxInt32` before constructing or inserting the item. The
annotation and runtime behavior should agree on how invalid input is handled.
