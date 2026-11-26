# Induction proof strategy

This document records the intended composition of proof obligations from
annotations. It supplements
[`symbolic-abstraction.bnf`](./symbolic-abstraction.bnf),
[`symbolic-abstraction-rules.md`](./symbolic-abstraction-rules.md), and
[`proof-generation.md`](./proof-generation.md). It is a strategy under
development, not a claim that a checker or proof assistant implements it.
Open semantic choices are identified below rather than silently resolved.

## Proof environment

Unless a proof explicitly states otherwise, assume:

- One execution thread is designated for the lifetime of the subject
  `PriceSpanSet`; only that thread may invoke its functions. An invocation
  by another thread violates the environment even if the calls do not
  overlap.
- Exactly one `PriceSpanSet` instance is in the proof scope. Introducing a
  second instance into that scope violates the environment.
- All mutations of the subject's tree and of relevant `Item.twin` pointers
  occur through the operations included in the proof analysis. No other
  code or thread mutates those state components, including through aliases.

A violation is outside the proof environment, not a counterexample to a
property proved under these assumptions. The first two assumptions alone do
not rule out direct mutation through a pointer; the third supplies that
restriction. These assumptions do not establish initial state, item
freshness, pointer ownership, or a particular invariant: those remain
separate premises or proof obligations. Specify the relevant item universe
and check that the analyzed operations cover its possible effects.

## Proof components

An abstract model with an induction goal tag declares a property to prove.
Conditions carrying the same tag are intended as induction steps for that
property. For example, `TreeUniquePrices[TUP]` is the goal associated with
conditions tagged `[TUP]` in `dag.go`.

An abstract model can also declare an equivalence set, as in
`TwinPricesUniquelyChained[TPUC](...) <=> {TUP,TCSP,TCU}`. Its proof is
intended to be composed strictly from the indicated sub-goals, not replaced
by an unrelated direct proof. A model not listed in another model's
equivalence set is a top-level goal in this composition. Model references
and equivalence-set dependencies are distinct relationships.

An invariant annotation contributes a condition at its pre- or post-state:

- `(?)` asks the analysis to establish the condition from the possible
  paths reaching that point. A tagged `(?)` may use relevant facts with the
  same tag, another tag, or no tag, provided their validity and applicability
  are established.
- `(!)pre-condition` refers to an explicitly established, logically
  corresponding earlier `(?)` pre- or postcondition. It must still hold at
  the annotated point. When tagged, its supporting `(?)` fact or facts must
  have the same tag.
- `(!)post-condition` refers to a logically corresponding `(?)` precondition
  and requires the fact still to hold after the annotated statement. When
  tagged, the supporting precondition must have the same tag. Whether it
  must be attached to the same statement remains open.

An annotation is an obligation, not a fact merely because it is written.
Tag equality identifies intended proof steps; it does not by itself prove
logical correspondence, availability on a path, or preservation.

## Constructing and checking a proof plan

1. Parse the annotations, resolve names and tags, check expression scope and
   types, and check the static abstract-model call graph as required by the
   symbolic-abstraction rules. Identify the goal, its tagged steps, and any
   equivalence-set dependencies.
2. Build control-flow and state-transition obligations, including relevant
   call sites, operation summaries, branches, and loop boundaries. Record
   assumptions and normal versus exceptional exits. External models are
   trusted only within the abstraction boundary described in
   `proof-generation.md`.
3. Establish each applicable `(?)` condition by analyzing incoming paths
   and modeled state changes. Verify any facts, call preconditions, and
   other dependencies used; retain unresolved dependencies explicitly.
4. For each tagged `(!)`, identify a minimal logically corresponding set
   of matching-tag `(?)` facts that covers the relevant paths. Report an
   ambiguous match rather than arbitrarily selecting among alternatives.
   Then prove that the matched facts remain valid at the `(!)` location,
   accounting for intervening mutations, aliases, calls, and loops.
5. Prove the tagged model using its tagged induction steps: establish the
   base case, preservation across every relevant transition, and the
   conclusion at the required program points. A different proof that bypasses
   the indicated steps does not satisfy this strategy, though it may be
   reported as a possible correction. For a model with an equivalence set,
   establish the required composition from its listed sub-goals.

Do not treat mutually dependent obligations as established solely by
assuming one another. A conditional local result is not a discharged
obligation until its dependencies are proved or explicitly accepted as
assumptions. In particular, proving a `(?)` with help from a `(!)` requires
the latter's provenance and preservation to be discharged without an
unjustified circular argument; ordinary induction still requires its base
case and preservation proof.

The `TreeUniquePrices` experiment applies these checks in gates: first all
`(?)` conditions tagged `TUP`, then all `(!)` correspondences, then their
continued validity, and finally the `GraphWalk` entry, per-iteration, and
exit conditions. Do not claim a later gate succeeds while an earlier one is
unresolved.

## Loops and outcomes

A `loophead` precondition applies once before loop entry and its
postcondition once at each loop exit, as specified in
[`symbolic-abstraction-rules.md`](./symbolic-abstraction-rules.md#loop-head-conditions).
Neither is automatically a per-iteration invariant. Place an intended
per-iteration condition inside the loop, and prove its initialization and
preservation. Distinguish normal postconditions from exceptional exits; the
existing loop-head rule also calls for evaluation on panic.

Report each obligation as `PROVED`, `REFUTED`, or `UNKNOWN` under the
outcome rules in `proof-generation.md`, and separately identify specification
errors. An ambiguous correspondence or a missing required step fails the
proposed proof plan; it is not by itself a counterexample to the property.
An unresolved obligation, unsupported operation, or unavailable assumption
must not be reported as a proof. Record the source locations, matching
facts, path coverage, dependencies, and reason for each failure or unknown.

## Open design questions

- Define logical correspondence and minimality, including branch-specific
  supporting sets, equivalently expressed facts, and what counts as an
  ambiguous match. Define how untagged `(!)` conditions are matched.
- Specify when a `(?)` fact is explicitly established and available across
  calls, loops, and exceptional control flow, including how provisional
  dependencies are discharged without circular reasoning.
- Define tag declaration and uniqueness rules, equivalence-set parameter
  mapping and dependency cycles, and the exact proof obligations implied by
  `<=>` beyond the stated sub-goal-based proof requirement.
- Clarify loop-exit condition evaluation on panic and with variables whose
  scopes end at the loop boundary. A possible `(*)` mode for loop internals
  was discussed but is not part of the current BNF.

A future source-linked intermediate representation could record these
obligations and their provenance for a proof assistant, but its format,
translation semantics, and trust boundary have not yet been specified.
