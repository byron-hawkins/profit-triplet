# DAG control-flow graphs

These graphs are regenerated from the current annotations in
`internal/profit-triplet/dag.go`. Unannotated Go statements are omitted as
nodes; their control-flow effect is represented by arrows between annotated
nodes. Each function or summary has a separate graph. A call-site node names
the graph for its callee; follow that graph to inspect its behavior.

Graph-generation conventions are defined in
[`cfg-generation-rules.md`](../meta/cfg-generation-rules.md). For annotated
predicates, the label is replaced by the corresponding Go condition or loop
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
    Q1["alias: m := make([]*Item, 0)"]
    Q2["for _, i := range tree.EverInserted"]
    Q3["if i.Price.Base &lt;= b &amp;&amp; i.Price.Bound &gt;= a"]
    Q4["m = append(m, i)"]
    Q5["return m, len(m) &gt; 0"]
    QExit(["exit"])

    Q1 --> Q2
    Q2 -->|next history item| Q3
    Q3 -->|true| Q4 --> Q2
    Q3 -->|false| Q2
    Q2 -->|history exhausted| Q5 --> QExit
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

## `Item.Join`

```mermaid
flowchart TD
    J1["Item#Join#CheckTwin<br/>predicate: if item.twin != nil"]
    J2["Item#Join#SpliceTwin"]
    J3["Item#Join#JoinTwin"]
    JExit(["exit"])

    J1 -->|true| J2 --> J3 --> JExit
    J1 -->|false| J3 --> JExit
```

## `Item.AllIn`

```mermaid
flowchart TD
    In1["AllIn#range: loophead<br/>predicate: for i := range item.in.Size()<br/>call → vector.Vector.Size"]
    In2["AllIn#yield<br/>predicate: if !yield(i, item.in.At(i))<br/>call → vector.Vector.At and yield callback"]
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
    W3["PriceSpanSet.GraphWalk#Step<br/>predicate: if set.GraphStep(&amp;item)<br/>call → PriceSpanSet.GraphStep"]
    W4["PriceSpanSet.GraphWalk#InsertItem<br/>alias: set.tree.EverInserted = append(set.tree.EverInserted, &amp;item)<br/>pre-condition: item.Price.Base &lt; item.Price.Bound"]
    W5["PriceSpanSet.GraphWalk#InsertItemErrorCheck<br/>predicate: if e != nil"]
    WPanic{{"PriceSpanSet.GraphWalk#InsertItemError: panic"}}
    WExit(["exit"])

    W1 -->|next input| W2 --> W3
    W3 -->|true| W4 --> W5
    W3 -->|false| W1
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
    G5["PriceSpanSet.GraphStep#MatchBoundCheck<br/>predicate: if item.Price.Base &gt; match.Price.Base"]
    G6["PriceSpanSet.GraphStep#MatchTruncate: call → PriceSpanSet.ItemTruncate"]
    G7["PriceSpanSet.GraphStep#MatchReplaceErrorCheck<br/>predicate: if e != nil"]
    GPanic{{"PriceSpanSet.GraphStep#MatchReplaceError: panic"}}
    G8["PriceSpanSet.GraphStep#MatchLink: call → Item.Link"]
    G9["PriceSpanSet.GraphStep#CheckSameInterval<br/>predicate: if item.Price.Bound == match.Price.Bound"]
    G10["PriceSpanSet.GraphStep#JoinSameInterval: call → Item.Join"]
    G11["PriceSpanSet.GraphStep#ElideSameInterval"]
    G12["GraphForkSamePrice: loophead<br/>predicate: for _, predecessor := range match.AllIn<br/>call → Item.AllIn"]
    G13["PriceSpanSet.GraphForkSamePriceSuperseded<br/>predicate: if predecessor.out.Back() == match"]
    G14["PriceSpanSet.GraphBranchSamePrice: call → Item.Link"]
    G15["PriceSpanSet.GraphStep#return"]
    GExit(["exit"])

    G1 --> G2 --> G3
    G3 -->|matched| G4
    G3 -->|not matched| G15
    G4 -->|next match| G5
    G4 -->|matches exhausted| G15
    G5 -->|true| G6 --> G7
    G7 -->|true| GPanic
    G7 -->|false| G8 --> G4
    G5 -->|false| G9
    G9 -->|true| G10 --> G11 --> G12
    G9 -->|false| G12
    G12 -->|next predecessor| G13
    G13 -->|true| G14 --> G12
    G13 -->|false| G12
    G12 -->|predecessors exhausted; next match| G4
    G15 --> GExit
```

## `PriceSpanSet.ItemTruncate`

```mermaid
flowchart TD
    T1["PriceSpanSet.ItemTruncate#Delete"]
    T2["PriceSpanSet.ItemTruncate#Truncate"]
    T3["PriceSpanSet.ItemTruncate#FindSameInterval"]
    T4["PriceSpanSet.ItemTruncate#CheckSameInterval<br/>predicate: if twinned"]
    T5["PriceSpanSet.ItemTruncate#JoinSameInterval"]
    T6["PriceSpanSet.ItemTruncate#ElideSameInterval"]
    T7["PriceSpanSet.ItemTruncate#Insert<br/>alias: set.tree.EverInserted = append(set.tree.EverInserted, item)<br/>pre-condition: item.Price.Bound &gt; item.Price.Base"]
    TExit(["exit"])

    T1 --> T2 --> T3 --> T4
    T4 -->|true| T5 --> T6 --> TExit
    T4 -->|false| T7 --> TExit
```
