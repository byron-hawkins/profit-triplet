
### Item.out trios are exclusively purchasable

@^PurchasableTriosOnPaths(set *PriceSpanSet)
@^AllPathTriosPurchasable(set *PriceSpanSet)

@^PurchasableTrio(a *Item, b *Item, c *Item) {


### ItemIndex ascends without overflow

@-Item#Constructor (1)
@-Item#Construct (1)
@-PriceSpanSet.GraphStep#SpliceTwinsLeft (1)
@-PriceSpanSet.GraphStep() (1)


### Transitive purchase compatibility

@^PriceCompatible(a *Item, b *Item)
@^LesserRootsPriceCompatible(set *PriceSpanSet, i *Item)

### Minimal complete purchase compatibility

@-PriceSpanSet.GraphStep#MatchLink: call (3)


### Twin chain integrity

@^IdenticalTwins(item *Item)
@-PriceSpanSet.GraphStep#MatchLink: call (1) (2)
@-PriceSpanSet.MatchTruncate#SpliceTwinsLeft (1)
@-PriceSpanSet.MatchTruncate#SpliceTwinsRight (3)
@-PriceSpanSet.MatchTruncate#Truncate (1)
@-PriceSpanSet.GraphForkSamePrice (1)


### Annotation listing

@-Item#Constructor:
(1)pre-condition: (set.ItemIndex.Load() < math.MaxInt32) && (set.ItemIndex.Load() >= int32(0))

@-Item#Construct:
(1)post-condition: i.index <= set.ItemIndex.Load() && i.index > prior && i.Margin == margin && i.Price == PriceSpan{ price, math.MaxInt32 } && i.best == [3]int{ margin, 0, 0 } && len(i.in) == 0 && len(i.out) == 0 && i.twin == nil

@^DirectedMonotonicGraph(roots []*Item)

@^DirectedMonotonicSubGraphs(pivot *Item)

@-Item.Link():
(1)pre-condition: len(to.out) == 0 && from != to && from.index < to.index && from.Price.Base < to.Price.Base
(2)pre-condition: DirectedMonotonicSubGraphs(from)
(3)post-condition: DirectedMonotonicSubGraphs(from)

@^PriceBounds(input []util.Input)

@^InputSizeBounds(set *PriceSpanSet, input []util.Input)



@-PriceSpanSet.GraphWalk()
(1)pre-condition: PriceBounds(input)
(2)pre-condition: InputSizeBounds(set, input)
(3)pre-condition: len(set.tree.EverInserted) == 0
(4)post-condition: PurchasableTriosOnPaths(set)
(5)post-condition: AllPathTriosPurchasable(set)

@-PriceSpanSet.GraphWalk#InsertItem: call
(1)post-condition: slices.Contains(set.tree.EverInserted, item)
(2)post-condition: func () bool { for i := item.Price.Base; i < 10000; i++ { x, _ := set.tree.AllIntersections(i, i); if !slices.Contains(x, item) { return false } }; return true }()

@-PriceSpanSet.GraphWalk#elide:
(1)post-condition: !slices.Contains(set.tree.EverInserted, item)

@^IdenticalTwins(item *Item)

@^OrderedIntersection(matches []*Item)

@^PriceCompatible(a *Item, b *Item)

@^LesserRootsPriceCompatible(set *PriceSpanSet, i *Item)

@^GreaterRootsPriceIncompatible(set *PriceSpanSet, i *Item)

@-PriceSpanSet.GraphStep()
(1)pre-condition: func () bool { for _, e := range set.tree.EverInserted { for t := e; t != nil; t = t.twin { if t.index >= item.index { return false } } } return true }()
(2)pre-condition: DirectedMonotonicGraph(set.roots)
(3)post-condition: LesserRootsPriceCompatible(set, item)
(4)post-condition: GreaterRootsPriceIncompatible(set, item)

@-PriceSpanSet.GraphStep#TreeQuery: call
(1)post-condition: OrderedIntersection(matches)

@-PriceSpanSet.GraphStep#MatchTruncate: call
(1)post-condition: func () bool { for i := match.Price.Base; i < match.Price.Bound; i++ { x, _ := set.tree.AllIntersections(i, i); if !slices.Contains(x, match) { return false } }; return true }()
(2)post-condition: func () bool { for i := match.Price.Base; i < item.Price.Base; i++ { x, _ := set.tree.AllIntersections(i, i); if !slices.Contains(x, match) { return false } }; return true }()

@-PriceSpanSet.GraphStep#MatchLink: call
(1)pre-condition: ego.index < item.index
(2)pre-condition: ego.Price == match.Price
(3)post-condition: PriceCompatible(ego, item)

@-PriceSpanSet.GraphStep#CheckSameInterval: predicate
(1)pre-condition: item.Price.Base == match.Price.Base

@-PriceSpanSet.GraphStep#SpliceTwinsRight:
(1)pre-condition: func () bool { for m := match; m != nil; m = m.twin { if (m.index >= item.index) { return false } } return true }()

@-PriceSpanSet.GraphStep#SpliceTwinsLeft
(1)post-condition: func () bool { for m := match; m != nil; m = m.twin { if (m.index >= set.ItemIndex.Load()) { return false } } return true }()
(2)post-condition: UniqueTwins(match)

@-PriceSpanSet.GraphForkSamePriceRoot: call
(1)pre-condition: slices.Contains(set.roots, match)

@-PriceSpanSet.GraphForkSamePrice: loophead, predicate, call
(1)post-condition: func () bool { for ego := match; ego != nil; ego = ego.twin { for _, c := range ego.in { if !PriceCompatible(c, item) { return false } } }; return true }

@-PriceSpanSet.GraphBranchSamePrice: call
(1)pre-condition: LesserRootsPriceCompatible(set, predecessor)
(2)post-condition: LesserRootsPriceCompatible(set, item)

@^JoinHomogeneousLarger(existing *Item, joining *Item)

@^MutuallyExclusiveJoin(existing *Item, joining *Item)

@^UniqueTwins(item *Item)

@-PriceSpanSet.MatchTruncate()
(1)pre-condition: IdenticalTwins(match)

@-PriceSpanSet.MatchTruncate#TruncateTwins: loophead, predicate
(1)pre-condition: JoinHomogeneousLarger(existing, match)
(2)pre-condition: MutuallyExclusiveJoin(existing, match)
(3)post-condition: IdenticalTwins(twin)

@-PriceSpanSet.MatchTruncate#Truncate
(1)pre-condition: ego.Price.Bound > at

@-PriceSpanSet.MatchTruncate#SpliceTwinsRight
(1)pre-condition: UniqueTwins(existing)
(2)pre-condition: UniqueTwins(match)
(3)pre-condition: func (tree *interval.SearchTree[*Item, int]) (bool) { for ego := match.twin; ego != nil; ego = ego.twin { if slices.Contains(tree.EverInserted, ego) { return false } } return true }(set.tree)

@-PriceSpanSet.MatchTruncate#SpliceTwinsLeft
(1)post-condition: func (tree *interval.SearchTree[*Item, int]) (bool) { for ego := match.twin; ego != nil; ego = ego.twin { if slices.Contains(tree.EverInserted, ego) { return false } } return true }(set.tree)
(2)post-condition: UniqueTwins(existing)
