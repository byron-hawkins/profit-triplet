# Project conversation history - collections, policies, and generic models

Summary of the discussion since
[November 27](./conversation-history.2026-11-27.md), recorded November 28,
2026. The user edits the source and annotations between reviews; findings
below refer to the revision discussed unless explicitly described as current.
Consult the current [BNF](../model/meta/symbolic-abstraction.bnf),
[semantic rules](../model/meta/symbolic-abstraction-rules.md), and
[`dag.go`](../internal/profit-triplet/dag.go) before continuing.

## Tree membership and collection grammar

- The user considered replacing the modeled `SearchTree.ItemSet` slice with
  an index-to-item map to specify `N`. The recommendation was to retain the
  slice and interpret `N` as the set of all pointers in it: a map would
  require index uniqueness and key consistency, lose the slice's ordering,
  and is not a supported abstract-collection operand. Excluding twin targets
  from the definition of `N` would assume the property to be proved.
- The collection declaration grammar was iteratively adjusted so the first
  collection appears on the opening comment line and each additional one
  occupies a separate decorated line. Inline and newline closing forms are
  supported. `<go-type>` was restricted to a single-line reference to an
  existing type; generic type references can otherwise span lines too.
- An earlier PO.2 review found that collection annotations identified
  important effects but did not yet establish complete coverage. In
  particular, the modeled deletion can remove every entry with a matching
  key, both writes in a twin splice need accounting, and non-interference,
  aliases, and exceptional paths remain proof obligations. The source has
  since been refactored; re-evaluate the current operations rather than
  relying on the earlier placement of annotations.

## Data-model references and field policies

- A `<data-model-name>` in an `@==` instantiation resolves by exact spelling
  against the models for the annotated field's type and denotes the modeled
  member of that particular field instance. A missing or non-unique match is
  a specification error. This rule does not prescribe a context-table
  implementation. Omitting the model name refers directly to the field.
- The user proposed policy-based reasoning for exact tree deletion: require
  unique `(Base, Bound)` keys and prevent key changes while an item is in
  the tree. Exact deletion additionally requires the proposed item to be
  present and the deletion arguments to match its current key. Declaring a
  unique-key policy does not independently prove `TreeUniquePrices`; any
  dependence on tree/twin exclusivity must be handled without circularity.
  A known policy violation invalidates policy-dependent analysis; inability
  to establish compliance is `UNKNOWN`, not a pass.
- The BNF now treats `@==` as an `<abstract-collection-model>` with
  qualifiers. The assistant updated its placement rule and added field
  policy semantics. `immutable` establishes an immutable field value after
  initialization. `under Interface` permits mutation during an invocation
  of a named interface method, including synchronous calls within it. Calls
  governed by the policy must select the method through that interface;
  for a different static receiver type, use an explicit conversion such as
  `PriceScope(set).Fresh(item)`.
- The user clarified that policies cover a field and its owned modeled
  contents, not independent objects merely referenced by those contents.
  Thus `ItemSet as N under PriceScope` governs changes to the tree and its
  modeled collection, but membership alone does not impose `PriceScope`
  restrictions on the pointed-to `Item` objects. Their own field policies
  remain applicable. The current grammar allows `immutable` alone, but not
  combined with another qualifier.

## Annotation and refactoring reviews

- The user introduced `ProxyPrice`, `Entwine`, and `PriceScope`, with field
  annotations for `PriceSpan.Base`, `PriceSpan.Bound`, `Item.twin`, and
  `PriceSpanSet.tree`. A review identified direct twin writes outside an
  `Entwine` invocation; the user subsequently refactored that splice.
  The current `GraphStep` calls `match.Splice(item, item)` directly, which
  is still relevant to the interface-selection policy. The user explicitly
  deferred policy compliance and proof justifications in the later
  syntax-and-semantics-only review.
- The later review identified spaces before `*/` in one-line element
  annotations as incompatible with the then-current `<comment-close>`.
  The BNF permits an empty element list, so an annotation without a colon
  is not intrinsically invalid; its closing whitespace is a separate issue.
  The user has not requested a grammar or source fix for this finding.
- Moved `Item.Splice` annotations still referenced `set` and `proxy`, which
  are not in its scope. The user added a source TODO noting this. A previous
  `MatchProxy#Link` expression applied `!` to an integer and dereferenced a
  potentially nil `existing`; the current source now uses a `!twinned ||`
  guard and a comparison. The source also has a TODO about accounting for
  the entire `head`-to-`tail` chain in `state[T]`.
- The last syntax-focused review found no missing abstract-model references
  or cycles in the static model-call graph. A package test passed during an
  earlier refactoring review, but Go tests do not validate annotation
  comments and are not a proof of policy compliance or current annotations.

## Generic model templates

- The user changed the `SearchTree` data model and summaries to use
  `interval.SearchTree[V,R]`. The concrete tree field has type
  `*interval.SearchTree[*Item,int]`. Previously the grammar and rules did
  not define `V` and `R` as model parameters or specify pointer matching.
- At the user's request, the assistant updated both the BNF and semantic
  rules. Generic data-model and summary tags introduce annotation-local,
  distinct type parameters; matching binds them by position to actual type
  arguments. Substitute resolved type-symbol references before checking
  the model or summary in the modeled environment. The actual Go generic
  type's constraints apply. A template is not asserted valid for every
  instantiation: incompatible instantiations are specification errors.
- Data-model resolution now also permits a model tagged with `T` to match
  a field of type `*T`, including instantiated generic types; competing
  matches for the same model name are errors. These changes were just made
  for the user to inspect. The entire current annotation set has not been
  re-reviewed under the new generic rules.

## VS Code working-set panel

The user hid the Keep/Undo working-set panel using its context menu. The
discussed alternative is `"chat.editor.workingSet.enabled": false`, which
shows a smaller pill rather than automatically accepting edits.

No complete tree/twin exclusivity proof or policy-compliance analysis was
established in this discussion. Continue to treat modeled components as the
proof boundary and distinguish specification errors, unresolved obligations,
and proved results.
