# Bounded symbolic-proof generation

These rules describe how an agent can use BNF annotations, generated control-flow graphs, and a bounded amount of Go source analysis to attempt a proof. The goal is to make useful progress while guaranteeing that difficult analysis terminates with an explicit inconclusive result rather than running indefinitely or claiming an unsupported proof.

## Project-wide abstraction boundary

For every proof in this project, annotations that model external components define their abstract semantics. They are not intended to describe or mirror the external component's concrete implementation. Prove properties up to and within the declared abstraction; do not inspect, execute, or prove refinement against dependency source unless a separate refinement proof is explicitly requested.

Treat each external data model and its operation summaries as the proof interface. A proof may rely on these contracts, while still checking that they are internally consistent, type-compatible with their uses, and sufficient for the theorem. If a model or summary is missing, inconsistent, ill-typed under its own abstraction, or insufficient, report `UNKNOWN` or a specification bug rather than silently strengthening it or consulting the concrete dependency.

This boundary is project-wide. A proof under these rules concerns the program composed with its declared abstractions, not whether those abstractions faithfully implement the real libraries.

## 1. Proof outcomes

Every proof attempt must finish with exactly one of these outcomes:

- **PROVED** — all proof obligations were discharged under the stated assumptions and supported Go semantics.
- **REFUTED** — a counterexample was derived. Call it a concrete counterexample only if all its values and operations are valid in the modeled Go subset; otherwise report it as an abstract counterexample requiring validation.
- **UNKNOWN** — a resource limit was reached, an operation was unsupported, an assumption was unavailable, or an obligation could not be established.

Never convert a timeout, depth-limit hit, unsupported operation, missing summary, or solver failure into `PROVED`. State the exact obligation, assumptions, and reason for `UNKNOWN`. A bounded search that finds no counterexample is not, by itself, a proof.

## 2. Inputs and trust boundary

Use the following evidence in order:

1. The source annotation and its explicitly defined BNF meaning.
2. The annotated function's Go control-flow graph, including branch predicates and call-site references.
3. Go statements in a deliberately supported subset, interpreted using the language semantics for that subset.
4. A callee's body, if available; otherwise its summary alias and substitutions.
5. Explicit proof assumptions and loop invariants.

Do not silently infer contracts from function names, comments outside the annotation syntax, library behavior, or unexamined code. Record any such fact as an assumption, and do not report an unconditional proof if it is required but unproved.

For external or generic library calls, use the project's annotated abstract model and operation summaries. Do not validate them against concrete dependency code under the project-wide abstraction boundary. If the abstract model or summary is missing or does not establish a needed fact, stop that branch with `UNKNOWN`.

## 3. Bounded Go analysis

### Supported constructs

An implementation may symbolically interpret a small, explicit Go subset, such as:

- local variable declarations and assignments;
- integer, boolean, nil, and pointer-identity expressions;
- field reads and writes;
- `if` branches with supported predicates;
- `return` and `panic`;
- calls with an available project body or project-declared abstract summary;
- modeled collection operations such as `append`, `len`, and `Empty`.

The subset must be listed by the implementation. Unsupported constructs (for example reflection, `unsafe`, goroutines, channels, arbitrary interface dispatch, unknown external calls, or unmodeled pointer aliasing) yield `UNKNOWN` whenever they affect an obligation. Do not execute arbitrary project code merely to evaluate an annotation. Evaluate only parsed, side-effect-free expressions composed of supported operations and known symbolic values.

### Resource budgets

Apply finite, explicit budgets. Suggested conservative defaults, configurable per proof:

- call expansion depth: **8**;
- symbolic execution steps per path: **10,000**;
- loop concrete unrolling before requiring an invariant: **3 iterations**;
- expression AST depth: **64**;
- live symbolic paths: **1,000**;
- total symbolic steps per proof attempt: **100,000**;
- wall-clock time per proof attempt: **60 seconds**.

The exact values are engineering limits, not logical bounds on program behavior. On reaching any budget, stop expanding that work item and report `UNKNOWN` for affected obligations. Do not truncate paths and then treat the remaining graph as complete.

Any SAT/SMT solver or other potentially blocking reasoning component must also have a finite timeout. Memoize function summaries by function, precondition, and relevant abstract state. Merge equivalent states when their symbolic stores and path conditions are identical. These optimizations must preserve proof soundness.

## 4. Symbolic execution rules

Represent an execution state as:

```text
(program point, symbolic store, heap/alias model, path condition, call stack)
```

Apply these rules:

1. **Entry:** Create symbolic values for inputs and record the function precondition.
2. **Assignment:** Evaluate supported right-hand sides and update the symbolic store according to Go evaluation order and assignment semantics.
3. **Branch:** For `if p`, fork into `pathCondition ∧ p` and `pathCondition ∧ ¬p`. Discard a branch only if its condition is proved unsatisfiable by the selected reasoning procedure; otherwise retain it.
4. **Call:** Check the callee precondition at the call site. If proven, interpret its body within the call-depth budget or apply its trusted summary. Check the callee postcondition on every normal return. A panic is a separate terminal outcome and does not imply a normal postcondition.
5. **Loop:** Prefer an inductive invariant. Check initialization, preservation for one arbitrary iteration, and that the invariant plus loop-exit condition establishes the desired postcondition. If no invariant is supplied or derived, unroll only up to the configured bound. A remaining loop after that bound yields `UNKNOWN`, never proof by finite sampling.
6. **Return:** Check the required postcondition for that path.
7. **Panic:** Record a panic terminal path. Prove properties of successful termination separately from properties that must hold on all paths. Do not silently ignore panic paths unless the theorem explicitly assumes normal completion.

Where exact Go semantics are not modeled, do not approximate them in a way that strengthens the result. For example, do not assume two pointers differ merely because their variable names or source expressions differ.

## 5. Proof obligations and annotations

Translate annotations into explicit obligations before execution:

- `predicate` supplies the source branch or loop predicate at the annotated control-flow point.
- `call` identifies a call edge; use the callee's graph/body or summary.
- A function call whose annotation does not include `call` is a `nop` for symbolic execution: do not expand its callee or apply any state effects from its body or summary. If the call's result is used, treat the result as unconstrained unless another annotation specifies it; do not infer a value from the ignored call.
- `panic` identifies a panic terminal.
- `loophead` marks a loop boundary; it is not itself an invariant.
- An invariant annotation must be classified as a precondition, loop invariant, or postcondition from its BNF meaning. If the grammar does not distinguish these roles, record the interpretation explicitly and do not assume one.
- An alias/substitution provides symbolic semantics only for the cases it states. Each substitution case is a separate guarded path; its guard is not a global fact.
- A `<maybe-invariant-list>` attached to a `<substitution>` constrains that substitution's applicability. The substitution may be selected only when all its value constraints and all attached preconditions hold. These conditions do not constrain sibling substitutions or become global preconditions for the enclosing function unless separately stated there. Attached postconditions describe the selected case's result and must be checked after evaluating it.

For a Hoare-style obligation `{P} f {Q}`, prove that every normally returning execution starting in a state satisfying `P` ends in a state satisfying `Q`. Handle panic behavior according to the theorem's stated scope.

## 6. Example strategy: prove the graph represented by `Item.in` and `Item.out` is acyclic

Use an abstract graph state:

```text
V = the Item objects currently in the graph
E = {(u, v) | v occurs in u.out}
```

Assume or prove that incoming and outgoing vectors agree: `v ∈ u.out` iff `u ∈ v.in`. This consistency is useful for understanding the representation but is not by itself enough to establish acyclicity.

### Lemma A: establish the `Item.Link` precondition

Interpret `Item.Link`'s annotation as the call precondition:

```text
to.out is empty  ∧  from is not to
```

At every call site, prove both conjuncts:

1. `GraphWalk` obtains `item` from `NewItem`; the constructor initializes `item.out` as empty.
2. `GraphWalk` calls `GraphStep(&item)` before inserting `item` into the abstract search-tree model. The model's `AllIntersections` contract returns only previously inserted item identities; `AllIn` yields existing incoming neighbors. Since `item` has not yet been inserted, these sources differ from it. Thus `from != to` at both annotated `Link` call sites.
3. `GraphStep` adds edges to `item` only through those calls, where `item` is the `to` argument. Each call appends to `from.out` and `item.in`; it does not append to `item.out`. Therefore the `to.out` emptiness condition remains true throughout this procedure.

“Fresh” must be established in the heap/alias model, not inferred from different variable names. Model each constructor call as creating a new abstract object identity, and prove that identity is not already in `V`. This is a property of the modeled program and constructor semantics, not of an external collection implementation.

The constructor calls `vector.NewVector`, whose project-declared abstract model establishes that it returns an empty vector. The abstract `PushBack` semantics establish that it appends exactly the supplied reference. These facts are used at the annotation boundary; this proof does not inspect or verify the external library implementation.

### Lemma B: `Item.Link` preserves acyclicity

Precondition: the existing graph `(V, E)` is acyclic, `from != to`, and `to.out` is empty.

The method adds exactly the edge `from → to` and its matching incoming reference. Since `to` has no outgoing edges at the time of insertion, no directed path can leave `to` to complete a cycle containing the new edge. The resulting graph is acyclic. The non-aliasing conjunct excludes the self-loop case. The argument does not require `to` to be fresh as a graph vertex, only that its outgoing vector is empty and `from` is distinct.

### Theorem: the graph remains acyclic throughout `GraphWalk`

Use induction over the input loop:

- **Base:** Before processing any input, there are no `Item` vertices or edges, so the graph is acyclic.
- **Inductive hypothesis:** The graph of all previously inserted items is acyclic.
- **Step:** The next item is fresh and has an empty `out` vector. `GraphStep` only adds edges to that item using `Item.Link`. Lemma A discharges each call precondition; Lemma B shows each added edge preserves acyclicity. Inserting the item into the search tree changes the lookup structure, not the graph edges, so acyclicity is preserved.
- **Conclusion:** After every successfully processed input, the graph represented by `Item.in`/`Item.out` is acyclic.

This theorem is about normal progress through the loop. If the interval-tree operations can panic or return an error that leads to panic, state whether the desired claim concerns all reachable states, only normally completed runs, or the graph immediately before panic. Do not treat panic as a normal loop step.

## 7. Required proof report

For every run, report:

- theorem and exact scope (for example, successful `GraphWalk` prefixes);
- source and annotation versions used;
- assumptions and trusted summaries;
- discharged obligations and the reasoning used;
- any explored/unrolled bounds;
- final outcome: `PROVED`, `REFUTED`, or `UNKNOWN`;
- for `UNKNOWN`, the first blocked obligation and whether the cause was a limit, unsupported semantics, missing contract, or unproved assumption.

Never describe a bounded test run, random mutation campaign, or finite loop unrolling as a proof of an unbounded property.
