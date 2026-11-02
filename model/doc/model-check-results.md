# Model-check results

## Current proof and bug findings

### [`model/proof/fresh-item-valid.md`](../proof/fresh-item-valid.md)

- **Objective:** Prove the `GraphWalk#InsertItem` precondition.
- **Outcome:** Established under the input-range and freshness assumptions.
- **Result:** The `GraphWalk` precondition bounds each input price to
  `1..9999`; the constructor sets the current item to
  `(Base: i.Price, Bound: math.MaxInt32)`. The truncate alias changes the
  matched item's bound, not the current item. The proof depends on
  [`linked-item-fresh.md`](../proof/linked-item-fresh.md) to establish that
  matches are distinct from the current item. Its scope is the insertion
  inequality on paths where `GraphStep` returns normally; it does not prove
  that all calls inside `GraphStep` are valid. In particular,
  [`invalid-item-rescope.md`](../bug/invalid-item-rescope.md) refutes
  `ItemTruncate#Insert`'s separate precondition. Nor does this prove that
  callers meet `GraphWalk`'s precondition.

### [`model/bug/invalid-item-rescope.md`](../bug/invalid-item-rescope.md)

- **Objective:** Check `ItemTruncate`'s insertion precondition.
- **Outcome:** Refuted by an equal-price counterexample.
- **Result:** The corrected `AllIntersections` postcondition permits a match
  with `Price.Base == a`. With inputs priced `20, 30, 20`, the abstract query
  may return the first item for the third query at `a = 20`. The caller guard
  `a < match.Bound` holds, but the truncate assignment sets `match.Bound = 20`,
  equal to `match.Base`. This violates `ItemTruncate#Insert`'s strict
  `Bound > Base` precondition. The report gives an annotated-line value-range
  trace and explains why initial insertion validity does not prevent a later
  in-place mutation from invalidating an item already referenced by history.

## Existing results

> The following entries describe results associated with git branch
> `iteration/bug-hunt-1`. They no longer correspond to the top-of-trunk
> repository content.

### [`model/proof/linked-item-fresh.md`](../proof/linked-item-fresh.md)

- **Objective:** Prove `GraphStep` links only to fresh items.
- **Outcome:** Established under stated assumptions.
- **Result:** At each `GraphStep` link, the current item is absent from the
  inserted history, each link source is distinct from it, and its `out`
  vector is empty. The proof uses the annotated constructor, history-only
  query result, `AllIn` vector semantics, and `GraphWalk` ordering. It assumes
  input prices are below `math.MaxInt32` and applies within the annotated
  abstraction, not as a refinement proof of the concrete external libraries.

### [`model/proof/dag-acyclic.md`](../proof/dag-acyclic.md)

- **Objective:** Prove the `Item.out` graph remains acyclic.
- **Outcome:** Established conditionally on link freshness.
- **Result:** Inducts over edge additions from an initially empty graph. Each
  link adds an edge to a target with no outgoing edges, from a distinct
  source, so it cannot create a cycle; the corresponding `in` update leaves
  the `out` graph unchanged. This proof depends on
  [`linked-item-fresh.md`](../proof/linked-item-fresh.md) for the `Item.Link`
  call-site obligations. It is scoped to `GraphWalk` in the annotated
  abstraction, assuming no concurrent or external `out` mutations.

### [`model/bug/invalid-item-insert-bug.md`](../bug/invalid-item-insert-bug.md)

- **Objective:** Find inputs violating the `InsertItem` precondition.
- **Outcome:** Counterexample found for unrestricted inputs.
- **Result:** A one-element input with `Price == math.MaxInt32` makes the
  constructor set `Base == Bound == math.MaxInt32`. With no input-range
  precondition, the empty-history query leaves the price unchanged and the
  strict insertion inequality fails. The report contains the annotated-line
  value-range trace.

### Supporting test output: [`model/bug/dag-mutation-test.log`](../bug/dag-mutation-test.log)

- **Objective:** Stress DAG construction with generated and mutated inputs.
- **Outcome:** A mutated input triggered an equal-endpoint interval panic.
- **Result:** The test generates 10 inputs, creates 10 mutated copies, and
  tests the original plus those copies (11 input sets of 10 items each).
  Mutations either change a randomly selected price to another value that
  changes its relative ordering against at least one other item, or swap two
  distinct input entries. These targeted changes exercise unusual ordering
  and interval-boundary cases without scanning the entire price domain. In
  the captured run, `input-8` fails; since `input-0` is the original set,
  this is the eighth mutated copy. The log reports an invalid interval with
  start and end both `162`, reached in `GraphStep`. It does not record which
  mutation was applied. This is runtime test evidence, not a proof.
