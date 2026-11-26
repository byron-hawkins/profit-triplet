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

### PO.1 Proof scope and base case

Define the item universe, initial state, and execution points covered by the proof. Establish an applicable empty-tree
base case and account for items or pointers already present at entry.

**Strategy:**

- item universe: define an *instance* `PSS` of `PriceSpanSet` and type <PSS.I>, a set of `Item` strictly bound to `PSS`
  - an `Item` instance `z` can only be an element of <PSS.I> under the following temporal constraint:
    - define `GW.k` as the `k`th iteration of `@-PriceSpanSet.GraphWalk#EachInput`
      - counter `k` does not reset at a subsequent invocation of `GraphWalk()`
    - `z ∈ PSS.I` at `GW.k` iff it has been constructed in `PSS.NewItem()` during `GraphWalk` iteration `GW.i: i <= k`
  - each `Item` in `T` is defined by `{ B | ∃ A ∈ PSS.I: A.twin == B, B ∈ PSS.I }'
  - consider field `Item.twin` to have a derived type `*PSS.I` and therefore be strictly bound to instance `PSS`
    - type `*PSS.I`, similar to a standard golang pointer, can be either `nil` or point to an element of <PSS.I>
  - `Item` instances have sequential indicies from 0 to no more than (math.MaxInt32 - 1)
  - no two `Item` instances should have the same index
  - no index should be assigned to an `Item` until all prior indicies have been assigned (i.e., no gaps)
  - an `Item` index is assigned exclusively in `PriceSpanSet.NewItem()`
    - only this method may modify the `PriceSpanSet.ItemIndex`
    - field `Item.index` may not be modified after construction
- operations covered by the proof are strictly those specified within `dag.go`, allowing interpretation based on:
  - proof structure and process, e.g., documents in meta/
  - code and annotation syntax and semantics, e.g., from the golang specification, from the annotation BNF, etc.
- lifecycle of the proof is limited to a single invocation of `GraphWalk` on a freshly instantiated `PriceSpanSet`
  - in practice, multiple calls are allowed, so the proof should remain valid over an equivalence of pregressive state:
    - a single call with an array A of N input elements
    - multiple calls on a single instance of `PriceSpanSet`, where:
      - concatenation of the successive input arrays exactly matches that same A
      - freshness is only required for the first call
      - no other `PriceSpanSet` methods are ever called externally

### PO.2 Effect and control-flow coverage

Identify every operation that can change tree membership or a relevant `twin` target, including effects through aliases,
calls, and exceptional paths. Establish non-interference for the remaining operations.

**Strategy:** To be developed.

### PO.3 Tree deletion

Show that `MatchProxy#Delete` preserves exclusivity and establish the membership facts needed by subsequent operations.

**Strategy:** To be developed.

### PO.4 Fresh-item insertion

Show that the item inserted at `GraphWalk#InsertItem` is not a `twin` target, including across all relevant `GraphStep`
iterations and branches.

**Strategy:** To be developed.

### PO.5 Match reinsertion

Show that the item inserted at `MatchProxy#Insert` is not a `twin` target. Account for its status before deletion and
any intervening changes.

**Strategy:** To be developed.

### PO.6 Fresh-item splice

Show that both assignments in `GraphStep#CheckSameInterval` preserve the property, including any relevant intermediate
state.

**Strategy:** To be developed.

### PO.7 Match-chain splice

Show that both assignments in `MatchProxy#CheckSameInterval` preserve the property, including the status of the deleted
match and the existing chain's successor.

**Strategy:** To be developed.

### PO.8 Match snapshot and iteration

Establish the facts needed about each `match` from the query snapshot when it is processed, despite changes made during
earlier iterations.

**Strategy:** To be developed.

### PO.9 Relation to tree-rooted twin chains

Determine what exclusivity implies for `TreeExcludesTwins`, and identify any separate head-membership or
traversal-termination obligations. The exclusivity property alone does not establish acyclicity.

**Strategy:** To be developed.
