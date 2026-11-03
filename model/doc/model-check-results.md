# Model-check results

## Current proofs for the latest `dag.go` annotations

### [`model/proof/linked-item-fresh.md`](../proof/linked-item-fresh.md)

- **Objective:** Establish the `Item.Link` call-site preconditions.
- **Outcome:** Established under the fresh-object and initial-state premises.
- **Result:** The current query alias builds results only from `EverInserted`,
  whose entries are earlier allocated items. `AllIn` yields previously
  linked graph nodes. The current item is fresh and has an empty `out`;
  `Link` changes its `in`, not its `out`. The proof accounts for conditional
  `GraphWalk#InsertItem`, `ItemTruncate#Insert` re-appending an existing
  match, and `Join` changing only twin links. It is the prerequisite for the
  acyclicity and insertion-range proofs.

### [`model/proof/dag-acyclic.md`](../proof/dag-acyclic.md)

- **Objective:** Prove `Item.out` remains acyclic in `GraphWalk`.
- **Outcome:** Established conditionally on link freshness.
- **Result:** Induction over every `Item.out` append. Each append is the
  `from.out` update in `Item.Link`, whose call-site conditions make the
  destination an empty-out current item and distinct from the source.
  `Join`, truncation, conditional insertion, and history updates do not
  mutate `out`. Depends on
  [`linked-item-fresh.md`](../proof/linked-item-fresh.md).

### [`model/proof/fresh-item-valid.md`](../proof/fresh-item-valid.md)

- **Objective:** Prove both reachable interval-insert inequalities.
- **Outcome:** Established under the input-range and freshness premises.
- **Result:** The `GraphWalk` precondition restricts input prices to
  `1..9999`; constructor bounds are `Base = input.Price` and
  `Bound = math.MaxInt32`. The current query alias returns only history
  items satisfying `Base <= query` and `Bound >= query`. On the truncation
  branch, `query > match.Base`, so after the bound assignment
  `Bound > Base`; if the same-interval branch is taken, the insert is skipped.
  The current item's price is unaffected because the matched item is fresh
  relative to it. Depends on
  [`linked-item-fresh.md`](../proof/linked-item-fresh.md).

## Historical results

> The following branch-specific findings do not necessarily correspond to
> the current top-of-trunk annotations.

### Results from `iteration/bug-hunt-1`

### [`model/bug/invalid-item-insert-bug.md`](../bug/invalid-item-insert-bug.md)

- **Objective:** Find inputs violating the original `InsertItem` precondition.
- **Outcome:** Counterexample found for the earlier unrestricted-input model.
- **Result:** A one-element input with `Price == math.MaxInt32` made the
  constructor set `Base == Bound == math.MaxInt32`. This report predates the
  current `GraphWalk` input-range precondition.

### Supporting test output: [`model/bug/dag-mutation-test.log`](../bug/dag-mutation-test.log)

- **Objective:** Stress DAG construction with generated and mutated inputs.
- **Outcome:** A mutated input triggered an equal-endpoint interval panic.
- **Result:** The test generated 10 inputs, created 10 mutated copies, and
  tested the original plus those copies. In the captured run, `input-8`
  failed with an invalid interval whose endpoints were both `162`. The log
  does not record whether the copy came from a relevant price change or an
  input swap. This is runtime test evidence for the earlier implementation,
  not a proof or a counterexample to the current annotations.

### Results from `iteration/bug-hunt-2`

### [`model/bug/invalid-item-rescope.md`](../bug/invalid-item-rescope.md)

- **Objective:** Check `ItemTruncate#Insert`'s interval-bound precondition.
- **Outcome:** Equal-price counterexample in `iteration/bug-hunt-2`.
- **Result:** The branch's query postcondition permits `match.Price.Base == a`.
  With inputs priced `20, 30, 20`, the abstract query may return the first
  item for the third query at `a = 20`. The caller's truncation branch is
  taken, and setting `match.Price.Bound = 20` makes it equal to `Base`,
  violating the strict `Bound > Base` precondition. This finding belongs to
  `iteration/bug-hunt-2`; it is not present in `iteration/bug-hunt-1`.
