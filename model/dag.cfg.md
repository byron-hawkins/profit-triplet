# DAG control-flow graphs

These graphs are derived from annotations in `internal/profit-triplet/dag.go` and the callee summaries present in that file. Unannotated Go statements are tacit: they are not nodes, but their control-flow effect is represented by arrows between annotated nodes.

The source annotations are interpreted by intent, not validated as strictly conforming to `symbolic-abstraction.bnf`. Some tags omit `: `, use comma-separated labels, include function signatures, or contain multiple `#` locators.

Graph-generation conventions are defined in [cfg-generation-rules.md](./cfg-generation-rules.md); these diagrams follow those rules. Each function has a separate graph. A call-site node names its target graph; follow that graph to inspect the callee's control flow.

## `PriceSpanSet.GraphWalk`

```mermaid
flowchart TD
    W1["PriceSpanSet.GraphWalk#EachInput: loophead<br/>predicate: for _, i := range input"]
    W2["GraphWalk#AddItem: call → PriceSpanSet.NewItem"]
    W3["PriceSpanSet.GraphWalk#Step: call → PriceSpanSet.GraphStep"]
    W4["GraphWalk#InsertItem"]
    W5["PriceSpanSet.GraphWalk#InsertItemErrorCheck: predicate: if e != nil"]
    WPanic{{"PriceSpanSet.GraphWalk#InsertItemError: panic"}}
    WExit(["exit"])

    W1 -->|next input| W2 --> W3 --> W4 --> W5
    W5 -->|true| WPanic
    W5 -->|false; next input| W1
    W1 -->|input exhausted| WExit
```

## `PriceSpanSet.NewItem`

```mermaid
flowchart TD
    N1["Item#constructor"]
    NExit(["exit"])
    N1 --> NExit
```

## `PriceSpanSet.GraphStep`

```mermaid
flowchart TD
    S1["PriceSpanSet.GraphStep#TreeQuery: call → interval.SearchTree.AllIntersections"]
    S2["PriceSpanSet.GraphStep#TreeQuery#MatchCheck: predicate: if matched"]
    S3["PriceSpanSet.GraphStep#EachMatch: loophead<br/>predicate: for _, match := range matches"]
    S4["PriceSpanSet.GraphStep#MatchBoundCheck: predicate: if item.Price.Base &lt; match.Price.Bound"]
    S5["PriceSpanSet.GraphStep#MatchReplace"]
    S6["PriceSpanSet.GraphStep#MatchTruncate"]
    S7["PriceSpanSet.GraphStep#MatchReplace"]
    S8["PriceSpanSet.GraphStep#MatchReplaceErrorCheck: predicate: if e != nil"]
    SPanic{{"PriceSpanSet.GraphStep#MatchReplaceError: panic"}}
    S9["PriceSpanSet.GraphStep#MatchLink: call → Item.Link"]
    S10["GraphForkSamePrice: loophead<br/>predicate: for _, predecessor := range match.AllIn"]
    S11["PriceSpanSet.GraphBranchSamePrice: call → Item.Link"]
    SExit(["exit"])

    S1 --> S2
    S2 -->|matched| S3
    S2 -->|not matched| SExit
    S3 -->|next match| S4
    S3 -->|matches exhausted| SExit
    S4 -->|true| S5 --> S6 --> S7 --> S8
    S8 -->|true| SPanic
    S8 -->|false; return from Item.Link| S9 --> S3
    S4 -->|false| S10
    S10 -->|next predecessor; call Item.Link| S11 --> S10
    S10 -->|predecessors exhausted; next match| S3
```

## `interval.SearchTree.AllIntersections` (summary)

```mermaid
flowchart TD
    A1["alias func (tree *interval.SearchTree) AllIntersections(a int, b int) { return make([]Item, %count), %matched }"]
    A2["%values[0]: int count &gt; 0, bool matched = true"]
    A3["%values[1]: int count = 0, bool matched = false"]
    AExit(["exit"])

    A1 --> A2 --> AExit
    A1 --> A3 --> AExit
```

## `Item.Link`

```mermaid
flowchart TD
    L1["Item.Link(): (?)invariant: to.out.Empty()"]
    L2["Item.Link#out"]
    L3["Item.Link#in"]
    LExit(["exit"])

    L1 --> L2 --> L3 --> LExit
```
