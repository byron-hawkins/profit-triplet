# Model-check results

Summary of the three proof and bug-report documents in `model/proof/` and
`model/bug/`.

## [`model/proof/linked-item-fresh.md`](../proof/linked-item-fresh.md)

- **Objective:** Prove `GraphStep` links only to fresh items.
- **Outcome:** Established under stated assumptions.
- **Result:** Shows that each `GraphStep` link source differs from the newly
  created item and that the target item's `out` vector is empty at each link.
  The proof relies on the annotated empty initial history, the tree query's
  history-only result model, the `AllIn` vector semantics, and the ordering of
  `GraphStep` before insertion. It assumes each input price is less than
  `math.MaxInt32` and applies within the project's annotated abstraction, not
  as a refinement proof of the concrete external libraries.

## [`model/proof/dag-acyclic.md`](../proof/dag-acyclic.md)

- **Objective:** Prove the `Item.out` graph remains acyclic.
- **Outcome:** Established conditionally on link freshness.
- **Result:** Inducts over edge additions from an initially empty graph. Each
  link adds an edge to a target with no outgoing edges, from a distinct
  source, so it cannot create a cycle; the corresponding `in` update does not
  change the `out` graph. This proof explicitly depends on
  [`linked-item-fresh.md`](../proof/linked-item-fresh.md) for the `Item.Link`
  call-site obligations. It is scoped to `GraphWalk` under the annotated
  abstraction, with no external or concurrent `out` mutations.

## [`model/bug/invalid-item-insert-bug.md`](../bug/invalid-item-insert-bug.md)

- **Objective:** Find inputs violating the `InsertItem` precondition.
- **Outcome:** Counterexample found; precondition refuted for unrestricted inputs.
- **Result:** A one-element input with `Price == math.MaxInt32` makes the
  constructor set `Base == Bound == math.MaxInt32`. The empty-history query
  returns no matches, leaving that span unchanged until `InsertItem`, where
  the required strict inequality fails. The report includes the annotated-line
  value-range trace. This result identifies a missing input constraint; it is
  independent of the acyclicity induction, while the freshness proof assumes
  prices below `math.MaxInt32`.
