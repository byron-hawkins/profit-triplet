
**Induction term semantics:**

- (!) indicates a property already asserted earlier, but to be confirmed in context of the enclosing annotation
  - pre: e.g., post-condition of a prior induction step (at the same or possible another annotation)
  - post: e.g., pre-condition already astablished at the enclosing annotation (unchanged by the annotated statement)
- (?) indicates a property to be evaluated on the basis of context at the enclosing annotation
  - pre: e.g., operations within the function prior to the enclosing annotation
  - post: e.g., operations within the annotated statement
- [tag] indicates an induction goal composed of the tagged condition and possibly other like-tagged conditions

**Questions:**

- subgoals usually have parameters, so how does a subgoal apply concretely to a containing goal?
  - assume forall semantics?


----

- twin chain traversal terminates
  A) no cycles
    1) twin chain is extended in limited ways:
       a) fresh item insertion
          1) item.twin == nil
          2) item.Price == node.Price
       b) append a truncated match chain to an existing chain
          1) match chain is empty (match.twin == nil && match.Price == existing.Price)
          2) both match and existing chains were assembled from A.1.a, A.1.b.1, A.1.b.2
  B) finite: only math.MaxInt32 instances of Item can be generated without panic
  C) non-comodification: concurrent invocation not allowed

This depends on the chain always having the same price. But how can that be established with a potentially infinite
chain? By induction over chain construction. However, it must also be shown that prices of twins always change in a
uniform atomic operation... which is performed on a chain traversal, requiring chain termination.

Also need to show that no twin is in the tree, and that the node is not a twin of any item.


          1) all prices had higher bound, therefore cannot have been on the chain

