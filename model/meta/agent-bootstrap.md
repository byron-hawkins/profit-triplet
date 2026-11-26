# Agent session bootstrap

This is a starting map for an agent joining the project without the earlier
conversation. Follow the user's current request and the current source; use
this file to find relevant context rather than treating it as a substitute
for reading the files involved.

## Project and annotations

The project is developing a Go price-span graph and annotations for reasoning
about its behavior. The main working source is
[`internal/profit-triplet/dag.go`](../../internal/profit-triplet/dag.go):
it contains the implementation, external-operation models, abstract models,
and condition annotations. Read the current source and its surrounding
control flow before analyzing a particular goal; annotations change often.

The main annotation and proof documents are:

- [`symbolic-abstraction.bnf`](./symbolic-abstraction.bnf): annotation
  syntax, including induction tags, equivalence sets, condition modes, and
  indentation. Parse against the current grammar, not examples from memory.
- [`symbolic-abstraction-rules.md`](./symbolic-abstraction-rules.md):
  supplementary naming, model-reference, argument-safety, call-graph, and
  loop-head rules.
- [`induction-proof-strategy.md`](./induction-proof-strategy.md): intended
  composition of tagged conditions and abstract-model goals, including the
  `TreeUniquePrices` (`TUP`) experiment and open semantic questions. This is
  a developing strategy, not an implemented or machine-checked proof system.
- [`proof-generation.md`](./proof-generation.md): proof outcomes, symbolic
  execution guidance, and the abstraction boundary for external models.
- [`cfg-generation-rules.md`](./cfg-generation-rules.md): conventions for
  generating control-flow diagrams from annotated Go source.

Other useful context, depending on the request:

- [`model/proof/`](../proof/): earlier proof attempts and goal definitions.
- [`model/doc/`](../doc/): generated or working documents, including a CFG,
  proof dependencies, induction notes, and model-check results.
- [`model/scripts/`](../scripts/): project scripts; inspect a script before
  relying on its output or assuming what it validates.

These supporting documents can describe earlier source revisions, even when
their headings say "current". Recheck their premises and conclusions against
the current source and rules before relying on them.

## Conversation history

[`agent/`](../../agent/) contains dated conversation-history summaries.
Start with the newest relevant summary, currently
[`conversation-history.2026-11-25.md`](../../agent/conversation-history.2026-11-25.md),
then consult older files for context as needed. Elements in older summaries
are more likely to be obsolete: models, names, grammar, and proof results
have changed. A history file records what was discussed, not a guarantee
that a claim remains correct or that all later edits are reflected.

## When continuing a proof

- Read the current annotation, relevant model definitions and operation
  summaries, and all relevant source paths; identify assumptions and
  outstanding obligations explicitly.
- Do not infer proof from a tag, an annotation, a passing Go test, or a
  prior informal conclusion. Follow the requested proof order and stop at
  its gates. Distinguish `PROVED`, `REFUTED`, and `UNKNOWN`, including
  conditional results and specification errors.
- The external models define the project's proof interface; they do not
  automatically prove that a dependency's concrete implementation refines
  them. Inspect concrete dependency behavior only when that is requested.
- For tagged induction, a `(?)` condition may use justified facts with any
  tag or no tag. A tagged `(!)` condition's corresponding established
  `(?)` facts must have the same tag. See the induction strategy for the
  remaining requirements and unresolved choices.
