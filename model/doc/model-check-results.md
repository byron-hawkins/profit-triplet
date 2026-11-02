# Model-check results

Summary of the proof reports in `model/proof/` and `model/bug/`.

## [`model/proof/linked-item-fresh.md`](../proof/linked-item-fresh.md)

- **Objective:** Prove `GraphStep` links only to fresh items.
- **Outcome:** Established under stated assumptions.
- **Result:** At each `GraphStep` link, the current item is absent from the
  inserted history, each link source is distinct from it, and its `out`
  vector is empty. The proof uses the annotated constructor, history-only
  query result, `AllIn` vector semantics, and `GraphWalk` ordering. It assumes
  input prices are below `math.MaxInt32` and applies within the annotated
  abstraction, not as a refinement proof of the concrete external libraries.
  This is the supporting proof used by the acyclicity result below.

## [`model/proof/dag-acyclic.md`](../proof/dag-acyclic.md)

- **Objective:** Prove the `Item.out` graph remains acyclic.
- **Outcome:** Established conditionally on link freshness.
- **Result:** Inducts over edge additions from an initially empty graph. Each
  link adds an edge to a target with no outgoing edges, from a distinct
  source, so it cannot create a cycle; the corresponding `in` update leaves
  the `out` graph unchanged. This proof depends on
  [`linked-item-fresh.md`](../proof/linked-item-fresh.md) for the `Item.Link`
  call-site obligations. It is scoped to `GraphWalk` in the annotated
  abstraction, assuming no concurrent or external `out` mutations.

## [`model/bug/invalid-item-insert-bug.md`](../bug/invalid-item-insert-bug.md)

- **Objective:** Find inputs violating the `InsertItem` precondition.
- **Outcome:** Counterexample found; precondition refuted for unrestricted inputs.
- **Result:** A one-element input with `Price == math.MaxInt32` causes the
  constructor to set `Base == Bound == math.MaxInt32`. The empty-history query
  returns no matches, leaving the span unchanged until `InsertItem`, where
  the required strict inequality fails. The report includes an annotated-line
  value-range trace. This identifies a missing input constraint; it is
  independent of the acyclicity induction, while the freshness proof assumes
  prices below `math.MaxInt32`.

## Supporting test output: [`model/bug/dag-mutation-test.log`](../bug/dag-mutation-test.log)

- **Objective:** Stress DAG construction with generated and mutated inputs.
- **Outcome:** A mutated input triggered an equal-endpoint interval panic.
- **Result:** The test generates 10 inputs, creates 10 mutated copies, and
  tests the original plus those copies (11 input sets of 10 items each).
  Mutations either change a randomly selected price to another value that
  changes its relative ordering against at least one other item, or swap two
  distinct input entries. These targeted changes exercise unusual ordering
  and interval-boundary cases without scanning the entire price domain. In
  this run, `input-8` fails; since `input-0` is the original set, this is the
  eighth mutated copy. The log reports an invalid interval with start and end
  both `162`, reached in `GraphStep`. It does not record whether this copy was
  produced by a price change or an input swap, so the exact mutation cannot
  be identified from the output alone. This is runtime test evidence, not a
  proof or a counterexample derived from the annotation-only analysis.
