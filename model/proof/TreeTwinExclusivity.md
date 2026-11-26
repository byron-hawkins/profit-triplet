# Tree/twin exclusivity

## Initial proof outline

For the subject `PriceSpanSet`, let `N` be the set of item pointers in its tree and `T` the set of non-nil targets of
relevant `Item.twin` pointers. The goal is:

```text
N ∩ T = ∅
```

In other words, no item is both nodular (in the tree) and twinned (the target of a `twin` pointer). The relevant item
universe and proof boundary still need to be specified. Use the proof environment in
[induction-proof-strategy.md](../meta/induction-proof-strategy.md).

Proposed compact induction:

- **Base:** An empty tree gives `N ∩ T = ∅`.
- **Deletion:** Removing a tree entry preserves the property.
- **Insertion:** Adding a tree entry preserves the property if the inserted item is not already a `twin` target.
- **Twin assignment:** Writing a `twin` pointer preserves the property if its new non-nil target is not in the tree.
- **Other operations:** Preserve the property if they do not change either set.

This is an initial outline, not a proof. In particular, the insertion and assignment premises and the coverage of all
relevant transitions remain to be established.

## Proof obligations

This section is a working draft. Obligations may be refined, combined, or restructured as the strategies are developed.

### Proof scope and base case

Define the item universe, initial state, and execution points covered by the proof. Establish an applicable empty-tree
base case and account for items or pointers already present at entry.

**Strategy:**

- item universe: a set of `Item` having sequential indicies from 0 to no more than (math.MaxInt32 - 1)
  - no duplicate indices
  - no missing sequence steps
- operations covered by the proof are entirely contained within `dag.go`, allowing external information for:
  - proof structure and process, e.g., documents in meta/
  - code and annotation syntax and sedmantics, e.g., from the golang specification, from the annotation BNF, etc.
- lifecycle of the proof is limited to a single invocation of `GraphWalk` on a freshly instantiated `PriceSpanSet`
  - in practice, multiple calls are allowed, so the proof depends on the absolute equivalence between:
    - a single call with an array A of N input elements
    - multiple calls for which the concatenation of successive input arrays exactly matches that same A

### Effect and control-flow coverage

Identify every operation that can change tree membership or a relevant `twin` target, including effects through aliases,
calls, and exceptional paths. Establish non-interference for the remaining operations.

**Strategy:** To be developed.

### Tree deletion

Show that `MatchProxy#Delete` preserves exclusivity and establish the membership facts needed by subsequent operations.

**Strategy:** To be developed.

### Fresh-item insertion

Show that the item inserted at `GraphWalk#InsertItem` is not a `twin` target, including across all relevant `GraphStep`
iterations and branches.

**Strategy:** To be developed.

### Match reinsertion

Show that the item inserted at `MatchProxy#Insert` is not a `twin` target. Account for its status before deletion and
any intervening changes.

**Strategy:** To be developed.

### Fresh-item splice

Show that both assignments in `GraphStep#CheckSameInterval` preserve the property, including any relevant intermediate
state.

**Strategy:** To be developed.

### Match-chain splice

Show that both assignments in `MatchProxy#CheckSameInterval` preserve the property, including the status of the deleted
match and the existing chain's successor.

**Strategy:** To be developed.

### Match snapshot and iteration

Establish the facts needed about each `match` from the query snapshot when it is processed, despite changes made during
earlier iterations.

**Strategy:** To be developed.

### Relation to tree-rooted twin chains

Determine what exclusivity implies for `TreeExcludesTwins`, and identify any separate head-membership or
traversal-termination obligations. The exclusivity property alone does not establish acyclicity.

**Strategy:** To be developed.
