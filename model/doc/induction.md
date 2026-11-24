
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
