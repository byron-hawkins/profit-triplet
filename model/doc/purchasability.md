
### Purchasability

**Purchasable Trio**: A sequence of 3 items {A, B, C} is purchasable if two properties hold:
- ( A.index < B.index < C.index }
- { A.Price.Base < B.Price.Base < C.Price.Base }.

**Purchase Compatible**: A sequence of items {A, B, ..., Z} is purchase eligible if two properties hold:
- ( A.index < B.index < ... < Z.index }
- { A.Price.Base < B.Price.Base < ... << Z.Price.Base }.

### Proof Goals

1. Every purchasable trio lies along some path from PriceSpanSet.roots over Item.out edges.
2. Every path of length at least 3 from PriceSpanSet.roots along Item.out edges is purchasable

#### Sub-Goals

0. Every path over Item.out is purchase compatible
1. Every constructed Item has an index superior to any prior constructed Item
2. For every pair of Item { A, B }, where B is linked over A.out, ( A.index < B.index && A.Price.Base < B.Price.Base )
3. For all Item A and every Item T reachable over A.twin, ( T.index > A.index && T.Price == A.Price )
4. After adding Item A to a PriceSpanSet S, for each R in S.roots:
   - A can be reached from R over Item.out if R.Price.Base < A.Price.Base
   - A cannot be reached from R over Item.out if R.Price.Base >= A.Price.Base



----

What's the point of PriceSpanSet.tree.AllIntersection()? On new input Item I, for any existing Item E:
- Where E.Price.Base >= I.Price.Base
  - (c) E will not be price compatible with I
  - for each Item C reachable over E.out
    - (o) C.out remains unchanged
    - (c) C will not be price compatible with I
- Where E.Price.Base == I.Price.Base
  - for each Item C that reaches E over C.out
    - (c) C will be price compatible with I
    - (o) if E in C.out, I will be added to C.out
    - if E not in C.out
      - (o) I will not be added to E.out
      - (o) for at least one Item P reachable over C.out, I will be added to P.out
- Where E.Price.Base < I.Price.Base
  - (c) I will be price compatible with E
  - if E.out contains at least one item P with P.Price.Base < I.Price.Base
    - (o) E.out will not be modified
    - for at least one Item P reachable over E.out
      - (o) I will be added to P.out
  - if E.out contains no item P with P.Price.Base < I.Price.Base
    - (o) I will be added to E.out


Link(from, to)
- pre: for all r in PriceSpanSet.roots where r.Price.Base < from.Price.Base && r.index < from.index: r.out -> ... -> from
- pre: from.index < to.index, from.Price.Base < to.Price.Base
- post: from in PriceSpanSet.tree.AllIntersection(i) for all i.Price.Base < to.Price.Base
- post: to in PriceSpanSet.tree.AllIntersection(i) for all i.Price.Base >= to.Price.Base


Dependency graph
- format
  - arity
- condition addressing
  - A ::= <tag> [pre|post]#

A0:
- A1
- A2
- ...
