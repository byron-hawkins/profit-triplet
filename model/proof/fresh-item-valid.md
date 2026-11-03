# Proof: `GraphWalk#InsertItem` precondition holds

## Goal and result

For every execution of `GraphWalk` that reaches
`PriceSpanSet.GraphWalk#InsertItem`, prove:

```go
item.Price.Base < item.Price.Bound
```

**Outcome:** PROVED under the `GraphWalk` input precondition, the current
constructor and call annotations, and the fresh-object fact established in
[`linked-item-fresh.md`](./linked-item-fresh.md).

The source now calls `GraphStep` as a predicate and executes the insertion
only when it returns true. This proof covers every such insertion. It also
checks the precondition of the reachable `ItemTruncate#Insert` operation.

## Input range

The `GraphWalk` precondition is:

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

Induct over the function literal's loop. The invariant is that each element
in the processed prefix satisfies `0 < i.Price < 10000`. It holds for the
empty prefix; an out-of-range element returns false, and an in-range element
preserves the invariant. Thus when the precondition is true, each input price
is in `[1, 9999]`. An empty input does not reach an insertion.

## `GraphStep` query facts

The current `AllIntersections` alias builds a result slice by iterating over
`tree.EverInserted` and appending only items satisfying:

```go
i.Price.Base <= b && i.Price.Bound >= a
```

For `GraphStep`, both query arguments are the current item's base, so
`a == b == item.Price.Base`. An invariant over the summary's loop is that
every element appended to `m` came from the history and satisfies both
filters. Therefore for each returned `match`:

```text
match.Price.Base <= item.Price.Base
match.Price.Bound >= item.Price.Base
```

The freshness result establishes `match != item`.

## Reachable truncation call

At `GraphStep#MatchBoundCheck`, truncation is called only on the branch:

```go
item.Price.Base > match.Price.Base
```

The call is `ItemTruncate(match, item.Price.Base)`. In the callee, rename its
`item` parameter to `match` to avoid confusing it with the current
`GraphStep` item. At entry:

```text
truncateItem.Price.Base = match.Price.Base
at = callerItem.Price.Base
at > truncateItem.Price.Base
```

The `ItemTruncate#Delete` call is not labelled `call` and is a no-op under
the proof rules. `ItemTruncate#Truncate` then assigns:

```go
truncateItem.Price.Bound = at
```

Thus `truncateItem.Price.Bound > truncateItem.Price.Base` follows from the
`GraphStep#MatchBoundCheck` branch guard. At `ItemTruncate#FindSameInterval`,
the `Find` result may be either `twinned` or not:

- If `twinned` is true, `ItemTruncate#JoinSameInterval` changes only twin
  links and the function returns `(false, nil)`; the annotated insert and its
  precondition are not reached.
- If `twinned` is false, execution reaches `ItemTruncate#Insert`, where the
  required strict inequality has just been established.

Therefore every path that reaches `ItemTruncate#Insert` satisfies its
precondition. If `ItemTruncate` returns an error, `GraphStep` may panic and
does not return normally; the panic path does not reach
`GraphWalk#InsertItem`.

## Current item is unchanged by `GraphStep`

For a current iteration, `Item#constructor` gives:

```text
current.Price.Base = input.Price
current.Price.Bound = math.MaxInt32
```

The input range therefore establishes:

```text
1 <= current.Price.Base <= 9999 < math.MaxInt32
```

`GraphStep#MatchTruncate` passes a query match to `ItemTruncate`; freshness
gives `match != current`. The truncate assignment therefore changes the
matched object's bound, not the current object's price. `Item.Join` changes
only twin fields. `Item.Link` changes only `in` and `out` vectors. None of
these operations changes `current.Price`.

`GraphStep` returns either true or false. Its false path skips
`GraphWalk#InsertItem`; on its true path, the current item's price still has
the constructor values above. Therefore at every reachable
`GraphWalk#InsertItem`:

```text
current.Price.Base <= 9999 < math.MaxInt32 = current.Price.Bound
```

which proves the required strict inequality.

## Annotated-line trace

| Annotated line | Established values or condition |
|---|---|
| `GraphWalk()` precondition | Every input price is in `[1,9999]`. |
| `GraphWalk#EachInput` | For the selected input, `1 <= input.Price <= 9999`. |
| `GraphWalk#AddItem` / `Item#constructor` | `current.Base = input.Price`; `current.Bound = math.MaxInt32`. |
| `GraphWalk#Step` / `GraphStep` | Call precondition holds by freshness; `current.Price` has not changed. |
| `GraphStep#TreeQuery` | Each returned match is a history object satisfying `match.Base <= current.Base <= match.Bound`. |
| `GraphStep#MatchBoundCheck` true branch | `current.Base > match.Base`; hence the truncate call has `at > truncateItem.Base`. |
| `ItemTruncate#Truncate` | Assign `truncateItem.Bound = at`; thus `truncateItem.Bound > truncateItem.Base`. |
| `ItemTruncate#FindSameInterval` / `#CheckSameInterval` | On `twinned`, return without insertion; otherwise reach `#Insert` with the strict inequality established. |
| `GraphStep#return` | Boolean return controls whether the current item is inserted; it does not change `current.Price`. |
| `GraphWalk#InsertItem` | Reached only if `GraphStep` returns true; `current.Base <= 9999 < math.MaxInt32 = current.Bound`. |

## Dependencies and limits

- Depends on [`linked-item-fresh.md`](./linked-item-fresh.md) for match
  provenance and `match != current`.
- Uses the current `AllIntersections` alias filter and the current
  `ItemTruncate` call/assignment/branch annotations.
- Treats unmarked calls as no-ops, as required by
  [`proof-generation.md`](../meta/proof-generation.md).
- Does not prove callers satisfy `GraphWalk`'s precondition, panic-freedom, or
  refinement of the external tree implementation.
