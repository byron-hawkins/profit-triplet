# DAG control-flow graphs

These graphs are regenerated from the current annotations in
`internal/profit-triplet/dag.go`. Unannotated Go statements are omitted as
nodes; their control-flow effect is represented by arrows between annotated
nodes. Each function or summary has a separate graph. A call-site node names
the graph for its callee; follow that graph to inspect its behavior.

Graph-generation conventions are defined in
[`cfg-generation-rules.md`](../meta/cfg-generation-rules.md). The labels below
use the annotation tags and labels. For annotations marked `predicate`, the
predicate label is replaced by the corresponding source condition or loop
header. Calls without a `call` label are not shown as call edges.

## `atomic.Int32.Add`

```mermaid
flowchart TD
    A1["alias: i.n += n"]
    AExit(["exit"])
    A1 --> AExit
```

## `vector.Vector.PushBack`

```mermaid
flowchart TD
    P1["alias: v.elements = append(v.elements, item)"]
    PExit(["exit"])
    P1 --> PExit
```

## `vector.Vector.At`

```mermaid
flowchart TD
    At1["alias: return v.elements[i]"]
    AtExit(["exit"])
    At1 --> AtExit
```

## `vector.Vector.Size`

```mermaid
flowchart TD
    S1["alias: return len(v.elements)"]
    SExit(["exit"])
    S1 --> SExit
```

## `vector.Vector.Empty`

```mermaid
flowchart TD
    E1["alias: return len(v.elements) == 0"]
    EExit(["exit"])
    E1 --> EExit
```

## `vector.NewVector`

```mermaid
flowchart TD
    N1["alias: return &amp;vector.Vector[*Item]{ make([]*Item, 0) }"]
    NExit(["exit"])
    N1 --> NExit
```

## `interval.SearchTree.AllIntersections`

```mermaid
flowchart TD
    Q1["alias: return tree.EverInserted[%i, %i + %count], %matched"]
    Q2["%values[0]: %i &gt;= 0 &amp;&amp; %i &lt; len(tree.EverInserted) - %count<br/>%count &gt; 0 &amp;&amp; %count &lt;= len(tree.EverInserted)<br/>%matched = true<br/>pre-condition: len(tree.EverInserted) &gt; 0"]
    Q3["%values[1]: (len(tree.EverInserted) == 0 &amp;&amp; %i == 0) || (%i &gt;= 0 &amp;&amp; %i &lt; len(tree.EverInserted))<br/>%count = 0<br/>%matched = false"]
    QExit(["exit"])

    Q1 --> Q2 --> QExit
    Q1 --> Q3 --> QExit
```

## `NewPriceSpanSet`

```mermaid
flowchart TD
    C1["alias: return PriceSpanSet{ tree { make([]*Item, 0) } }"]
    CExit(["exit"])
    C1 --> CExit
```

## `PriceSpanSet.NewItem`

```mermaid
flowchart TD
    I1["Item#constructor: (?)post-condition<br/>return Item{ Price: PriceSpan{ Base: price, Bound: math.MaxInt32 }, out: vector.NewVector[*Item]() }"]
    IExit(["exit"])
    I1 --> IExit
```

## `Item.AllIn`

```mermaid
flowchart TD
    In1["AllIn#range: loophead<br/>predicate: for i := range item.in.Size()<br/>call → vector.Vector.Size"]
    In2["AllIn#yield<br/>predicate: if !yield(i, item.in.At(i))<br/>call → vector.Vector.At, yield callback"]
    InExit(["exit"])

    In1 -->|next index| In2
    In2 -->|yield returns false| InExit
    In2 -->|yield returns true; next index| In1
    In1 -->|range exhausted| InExit
```

## `Item.Link`

```mermaid
flowchart TD
    L1["Item.Link(): (?)pre-condition<br/>to.out.Empty() &amp;&amp; from != to"]
    L2["Item.Link#out: call → vector.Vector.PushBack"]
    L3["Item.Link#in: call → vector.Vector.PushBack"]
    L4["Item.Link(): (?)post-condition<br/>to.out.Empty() &amp;&amp; slices.Contains(from.out.elements, to) &amp;&amp; slices.Contains(to.in.elements, from)"]
    LExit(["exit"])

    L1 --> L2 --> L3 --> L4 --> LExit
```

## `PriceSpanSet.GraphWalk`

```mermaid
flowchart TD
    W1["PriceSpanSet.GraphWalk#EachInput: loophead<br/>predicate: for _, i := range input"]
    W2["GraphWalk#AddItem: call → PriceSpanSet.NewItem"]
    W3["PriceSpanSet.GraphWalk#Step: call → PriceSpanSet.GraphStep"]
    W4["PriceSpanSet.GraphWalk#InsertItem<br/>alias: set.tree.EverInserted = append(set.tree.EverInserted, &amp;item)<br/>pre-condition: item.Price.Base &lt; item.Price.Bound"]
    W5["PriceSpanSet.GraphWalk#InsertItemErrorCheck<br/>predicate: if e != nil"]
    WPanic{{"PriceSpanSet.GraphWalk#InsertItemError: panic"}}
    WExit(["exit"])

    W1 -->|next input| W2 --> W3 --> W4 --> W5
    W5 -->|true| WPanic
    W5 -->|false; next input| W1
    W1 -->|input exhausted| WExit
```

## `PriceSpanSet.GraphStep`

```mermaid
flowchart TD
    G1["PriceSpanSet.GraphStep(): (?)pre-condition<br/>!slices.Contains(set.tree.EverInserted, item)"]
    G2["PriceSpanSet.GraphStep#TreeQuery: call → interval.SearchTree.AllIntersections"]
    G3["PriceSpanSet.GraphStep#TreeQueryMatchCheck<br/>predicate: if matched"]
    G4["PriceSpanSet.GraphStep#EachMatch: loophead<br/>predicate: for _, match := range matches"]
    G5["PriceSpanSet.GraphStep#MatchBoundCheck<br/>predicate: if item.Price.Base &lt; match.Price.Bound"]
    G6["PriceSpanSet.GraphStep#MatchTruncate"]
    G7["PriceSpanSet.GraphStep#MatchReplaceErrorCheck<br/>predicate: if e != nil"]
    GPanic{{"PriceSpanSet.GraphStep#MatchReplaceError: panic"}}
    G8["PriceSpanSet.GraphStep#MatchLink: call → Item.Link"]
    G9["GraphForkSamePrice: loophead<br/>predicate: for _, predecessor := range match.AllIn<br/>call → Item.AllIn"]
    G10["PriceSpanSet.GraphBranchSamePrice: call → Item.Link"]
    GExit(["exit"])

    G1 --> G2 --> G3
    G3 -->|matched| G4
    G3 -->|not matched| GExit
    G4 -->|next match| G5
    G4 -->|matches exhausted| GExit
    G5 -->|true| G6 --> G7
    G7 -->|true| GPanic
    G7 -->|false| G8 --> G4
    G5 -->|false; next predecessor| G9
    G9 -->|next predecessor| G10 --> G9
    G9 -->|predecessors exhausted; next match| G4
```
