package dag

import (
  "fmt"
 "strings"
 "slices"
 "math"
 "sync/atomic"
 "github.com/rdleal/intervalst/interval"

	"demo/profit-triplet/pkg/util"
)

const MaxPrice = 10000

/* @+atomic.Int32: data-model n int32 */
/* @+atomic.Int32.Load(): alias func (i *atomic.Int32) Load() int32 { return i.n } */
/* @+atomic.Int32.Add(n int32): alias func (i *atomic.Int32) Add(n int32) int32 { i.n += n; return i.n } */

/* @+interval.SearchTree[V,R]: data-model ItemSet []V */
/* @+interval.SearchTree[V,R].AllIntersections(a R, b R):
 *   alias func (tree *interval.SearchTree[V,R]) AllIntersections(a, b R) ([]V, bool) { m := make([]V, 0); for _, i := range tree.ItemSet { if i.Price.Base <= b && i.Price.Bound >= a { m = append(m, i) } }; return m, len(m) > 0 }
 */
/* @+interval.SearchTree[V,R].Find(a R, b R):
 *   alias func (tree *interval.SearchTree[V,R]) Find(a, b R) (V, bool) { for _, i := range tree.ItemSet { if i.Price.Base == a && i.Price.Bound == b { return i, true } }; return nil, false }
 */
/* @+interval.SearchTree[V,R].Insert(a R, b R, i V):
 *   alias func (tree *interval.SearchTree[V,R]) Insert(a, b R, i V) (error) { tree.ItemSet = append(tree.ItemSet, i); slices.SortFunc(tree.ItemSet, compare_items); return nil }
 *   (?)pre-condition: i.Price.Base < i.Price.Bound && a == i.Price.Base && b == i.Price.Bound
 *   (?)pre-condition: !slices.Contains(tree.ItemSet, i)
 */
/* @+interval.SearchTree[V,R].Delete(a R, b R):
 *   alias func (tree *interval.SearchTree[V,R]) Delete(a, b R) { tree.ItemSet = slices.Collect( func(yield func(V) bool) { for _, i := range tree.ItemSet { if i.Price.Base != a || i.Price.Bound != b { if !yield(i) { return } } } } ) }
 */

/******** TwinPricesUniquelyChained [TPUC] */

/* @^TwinPricesUniquelyChained[TPUC](set *PriceSpanSet) <=> {TUP,TCSP,TCU}:
       all := make([]*Item, set.ItemIndex.Load())
       for _, node := range set.tree.ItemSet {
         for ego := node; ego != nil; ego = ego.twin {
           if ego.Price != node.Price { return false }
           if ego.index < node.index { return false }
           if all[ego.index] != nil { return false }
           all[ego.index] = ego
         }
       }

       for _, i := range all {
         if i == nil { return false }
       }
       return true
 */

/******** TreeUniquePrices [TUP] */

/* @^TreeUniquePrices[TUP](set *PriceSpanSet):
       for i, left := range set.tree.ItemSet {
         for _, right := range set.tree.ItemSet[i+1:] {
           if left.Price == right.Price { return false }
         }
       }
       return true
 */

/* @^TreeExcludesTwins(set *PriceSpanSet, node *Item):
       if !slices.Contains(set.tree.ItemSet, node) { return false }
       for ego := node.twin; ego != nil; ego = ego.twin {
         if slices.Contains(set.tree.ItemSet, ego) { return false }
       }
       return true
 */

/******** TwinChainsSamePrice [TCSP] */

/* @^TwinChainsSamePrice[TCSP](set *PriceSpanSet):
       for _, i := range set.tree.ItemSet {
         if !@TwinChainSamePrice(i) { return false }
       }
       return true
 */

/* @^TwinChainSamePrice(node *Item):
       if node.twin == nil { return true }
       for ego := node.twin; ego != nil; ego = ego.twin {
         if node.Price != ego.Price { return false }
       }
       return true
 */

/******** TwinChainsUnique [TCU] */

/* @^TwinChainsUnique[TCU](set *PriceSpanSet):
       for _, i := range set.tree.ItemSet {
         if !@TwinChainUnique(i) { return false }
       }
       return true
 */

/* @^TwinChainUnique(node *Item):
        unique := make(map[int32]*Item)
        for ego := node; ego != nil; ego = ego.twin {
          if unique[ego.index] != nil { return false }
          unique[ego.index] = ego
        }
        return true
 */

/* @^TwinChainIndexBounded(set *PriceSpanSet, index int32, node *Item):
       if index < 0 { index = set.ItemIndex.Load() }
       if !slices.Contains(set.tree.ItemSet, node) { return false }
       for ego := node; ego != nil; ego = ego.twin {
         if (ego.index >= index) { return false }
       }
       return true
 */

/* @^MutuallyExclusivePriceChains(existing *Item, joining *Item):
       for x := existing; x != nil; x = x.twin {
         if x.Price == joining.Price { return false }
       }
       for j := joining; j != nil; j = j.twin {
         if j.Price == existing.Price { return false }
       }
       return true
 */

/* @^FreshTwinExtendsChain(node *Item, fresh *Item):
       for ego := node; ego != nil; ego = ego.twin {
         if (ego.index >= fresh.index) { return false }
       }
       return true
 */

/******** [LRGR] */

/* @^LesserRootsGraphReachable[LRGR](set *PriceSpanSet, i *Item):
       for _, root := range set.roots {
         if root.index < i.index && root.Price.Base < i.Price.Base {
           if !@PurchaseStepwiseReachable(root, i) { return false }
         }
       }
       return true
 */

/* @^TwinsProxied(node *Item, frontier *Item):
       if node == nil { return true }
       for ego := node; ego != nil; ego = ego.twin {
         if !slices.Contains(ego.out, frontier) { return false }
         if ego.Price.Base >= frontier.Price.Base || ego.Price.Bound != frontier.Price.Base { return false }
       }
       return true
 */

/* @^PurchaseStepwiseReachable(a *Item, b *Item):
       stack := NewStack()
       for _, o := range a.out { stack.push(o) }
       visited := make(map[int32]*Item)
       for stack.size() > 0 {
         i := stack.pop()
         if visited[i.index] != nil { continue }
         if i == b { return true }
         visited[i.index] = i
         if i.index < b.index && i.Price.Base < b.Price.Base {
           for _, o := range i.out { stack.push(o) }
         }
       }
       return false
*/

/* @^ExcludedReachProxy(set *PriceSpanSet, proxies []*Item, item *Item):
       for _, i := range set.tree.ItemSet {
         if i.Price.Base < item.Price.Base {
           if !@ProxyPurchaseStepwiseReachable(i, proxies) { return false }
         }
       }
       return true
 */

/* @^HasAlterEgo(node *Item, ego *Item):
       for i := node; i != nil; i = i.twin {
         if i == ego { return true }
       }
       return false
 */

/* @^HasAlterEgoAt(set *PriceSpanSet, ego *Item):
       node, found := set.tree.Find(ego.Price.Base, ego.Price.Bound)
       if !found { return false }
       return @HasAlterEgo(node, ego)
 */

/******** GreaterRootsGraphUnreachable [GRGU] */

/* @^GreaterRootsGraphUnreachable[GRGU](set *PriceSpanSet, i *Item):
       for _, root := range set.roots {
         if root.index >= i.index || root.Price.Base >= i.Price.Base {
           if @PurchaseStepwiseReachable(root, i) { return false }
         }
       }
       return true
 */

/******** MinimalPriceCompatibleIntersection [MPCI] */

/* @^MinimalPriceCompatibleIntersection[MPCI](set *PriceSpanSet, matches []*Item, item *Item):
       for _, i := range set.tree.ItemSet {
         if slices.Contains(matches, i) {
           if !@PriceCompatibleItems(i, item) { return false }
           for _, o := range i.out {
             if @PriceCompatibleItems(o, item) { return false }
           }
         } else if !@PriceCompatibleItems(i, item) {
           if @ProxyPurchaseStepwiseReachable(i, matches) { return false }
         } else if !@ProxyPurchaseStepwiseReachable(i, matches) {
           return false
         }
       }
       return true
 */

/******** DirectedMonotonicGraph [DMG]  */

/* @^DirectedMonotonicGraph[DMG](set *PriceSpanSet) <=> {DMSG}:
       visited := make(map[int32]*Item)
       stack := NewStack()
       for _, root := range set.roots { stack.push(root) }
       for stack.size() > 0 {
         i := stack.pop()
         for _, o := range i.out {
           if (i.index >= o.index) { return false }
           if (i.Price.Base >= o.Price.Base) { return false }
           if visited[o.index] != nil { continue }
           visited[o.index] = o
           stack.push(o)
         }
       }
       return true
 */

/* @^DirectedMonotonicSubGraphs[DMSG](pivot *Item):
       visited := make(map[int32]*Item)
       stack := NewStack()
       stack.push(pivot)
       for stack.size() > 0 {
         i := stack.pop()
         for _, o := range i.out {
           if (i.index >= o.index) { return false }
           if (i.Price.Base >= o.Price.Base) { return false }
           if visited[o.index] != nil { continue }
           visited[o.index] = o
           stack.push(o)
         }
       }

       visited = make(map[int32]*Item)
       stack.push(pivot)
       for stack.size() > 0 {
         o := stack.pop()
         for _, i := range o.in {
           if (o.index <= i.index) { return false }
           if (o.Price.Base <= i.Price.Base) { return false }
           if visited[i.index] != nil { continue }
           visited[i.index] = i
           stack.push(i)
         }
       }

       return true
 */

/******** ItemIndexProgress [IIP] */

/* @^ItemIndexProgress[IIP](set *PriceSpanSet):
       max := set.ItemIndex.Load()
       for _, i := range set.tree.ItemSet {
         if i.index >= max { return false }
       }
       return true
 */

/******** [untagged]  */

/* @^PriceBounds(input []util.Input):
       for _, i := range input {
         if i.Price <= 0 || i.Price >= MaxPrice { return false }
       }
       return true
*/

/* @^InputSizeBounds(set *PriceSpanSet, input []util.Input):
       (len(input) < math.MaxInt32) &&
       (set.ItemIndex.Load() < (math.MaxInt32 - int32(len(input)))) &&
       (set.ItemIndex.Load() >= int32(0))
*/

/* @^PurchasableTrio(a *Item, b *Item, c *Item):
       return a.index < b.index && b.index < c.index && a.Price.Base < b.Price.Base && b.Price.Base < c.Price.Base
 */

/* @^PurchasableTriosOnPaths(set *PriceSpanSet):
	     indexer := func(a, b *Item) int {
	     	return int(a.index - b.index)
	     }

       if len(set.all) < 3 { return true }

       for i, item1 := range set.all[ :len(set.all)-2] {
         for j, item2 := range set.all[i+1 : len(set.all)-1] {
           for _, item3 := range set.all[i+j+2 : len(set.all)] {
             trio := [3]*Item { item1, item2, item3 }
             slices.SortFunc(trio[:], indexer)
             c := @PurchaseStepwiseReachable(trio[0], trio[1]) && @PurchaseStepwiseReachable(trio[1], trio[2])
             if c != @PurchasableTrio(trio[0], trio[1], trio[2]) { return false }
           }
         }
       }
       return true
 */

/* @^AllPathTriosPurchasable(set *PriceSpanSet):
       for _, item := range set.all {
         for _, i := range item.in {
           if i.index >= item.index || i.Price.Base >= item.Price.Base { return false }
         }
         for _, o := range item.out {
           if o.index <= item.index || o.Price.Base <= item.Price.Base { return false }
         }
       }
       return true
 */

/* @^OrderedIntersection(matches []*Item):
       c := len(matches)
       if c < 2 { return true }
       for i := range (c - 1) {
         if matches[i].Price.Base > matches[i+1].Price.Base { return false }
         if matches[i].Price.Base == matches[i+1].Price.Base {
           if matches[i].Price.Bound >= matches[i+1].Price.Bound { return false }
         }
       }
       return true
*/

/* @^GraphAnyReachable(a *Item, b []*Item):
       stack := NewStack()
       for _, o := range a.out { stack.push(o) }
       visited := make(map[int32]*Item)
       for stack.size() > 0 {
         i := stack.pop()
         if slices.Contains(b, i) { return true }
         visited[i.index] = i
         for _, o := range i.out {
           if visited[o.index] == nil { stack.push(o) }
         }
       }
       return false
*/

 // TODO: heavy if executed, not really necessary
/* ExcludedIncompatible(PriceSpanSet *set, proxies []*Item, threshold int):
       for _, i := range set.tree.ItemSet {
         if slices.Contains(proxies, i) { continue }
         if i.Price.Base > threshold {
           if @GraphAnyReachable(i, proxies) { return false }
         }
       }
 */

/* JoinHomogeneousLarger(existing *Item, joining *Item):


       if (xspan >= jspan) { return false } // append proxied matched chain
       for i := existing.twin; i != nil; i = i.twin {
         if (xspan != (i.Price.Bound - i.Price.Base)) { return false }
       }
       for i := joining.twin; i != nil; i = i.twin {
         if (jspan != (i.Price.Bound - i.Price.Base)) { return false }
       }
       return true
 */

/* @^UniqueTwins(item *Item):
       for ego := item; ego.twin != nil; ego = ego.twin {
         for successor := ego.twin; successor != nil; successor = successor.twin {
           if ego == successor { return false }
         }
       }
       return true
 */

/******** [generic] */

/* @^ProxyPurchaseStepwiseReachable(item *Item, proxies []*Item):
       if slices.Contains(proxies, item) { return true }

       threshold := 0
       for _, i := range proxies { threshold = max(i.Price.Base, threshold) }

       step:
       for _, o := range item.out {
         if slices.Contains(proxies, o) { return true }
         if o.Price.Base < threshold {
           item = o
           goto step
         }
       }
       return false
 */

/* @^PriceCompatibleItems(a *Item, b *Item):
       return a.index < b.index && a.Price.Base < b.Price.Base
 */

/* @^PriceCompatible(a *Item, b *HypotheticalItem):
       return a.index < b.index && a.Price.Base < b.Price.Base
 */

/* @^PriceCompatibleRange(set *PriceSpanSet, a *Item, plow int, phigh int):
       if phigh > MaxPrice { phigh = MaxPrice }
       return a.index < set.ItemIndex.Load() && a.Price.Base < plow && a.Price.Bound > phigh
 */

/* @^MatchIncludes(set *PriceSpanSet, item *Item):
       low := item.Price.Base
       x, _ := set.tree.AllIntersections(low, low)
       if !slices.Contains(x, item) { return false }

       high := min(MaxPrice, item.Price.Bound)
       x, _ = set.tree.AllIntersections(high, high)
       return slices.Contains(x, item)
 */

type Stack struct {
  a []*Item
}

func NewStack() (Stack) {
  return Stack { a: make([]*Item, 0) }
}

func (s *Stack) size() int {
  return len(s.a)
}

func (s *Stack) push(i *Item) {
  s.a = append(s.a, i)
}

func (s *Stack) pop() *Item {
  if len(s.a) == 0 { return nil }
  i := s.a[len(s.a)-1]
  s.a = s.a[:len(s.a)-1]
  return i
}

type PriceSpan struct {
  // Base: current selling price
	Base  int   /* @== immutable */
  // Bound: end of the span covered by the enclosing item
	Bound int   /* @== under ProxyPrice */
}

type ProxyPrice interface {
  Link(*Item) (*Item)
}

func (price PriceSpan) String() (string) {
	if price.Bounded() {
		return fmt.Sprintf("[%v - %v)", price.Base, price.Bound)
	} else {
		return fmt.Sprintf("[%v -)", price.Base)
	}
}

func (price *PriceSpan) Bounded() (bool) { return price.Bound < math.MaxInt32 }

type Item struct {
	index  int32
	Margin int
	Price  PriceSpan

	best   [3]int

	in  []*Item
	out []*Item
	twin *Item   /* @== under Entwine */
}

type Entwine interface {
  Splice(*Item, *Item)
}

// TODO: state[T] needs ∪ {head..tail}
// TODO: no `set` or `proxy` here for the annotations

/* @-Item.Splice() */
func (item *Item) Splice(head *Item, tail *Item) {
  /* @-Item.Splice#SpliceTwinsRight:
   *   (!)pre-condition[TCU]: TwinChainUnique(item)
   *   (!)pre-condition[TCU]: TwinChainUnique(head)
   *   (!)pre-condition[TCU]: TwinChainIndexBounded(set, head.index, item)
   *   (!)pre-condition[TCSP]: TwinChainSamePrice(item)
   *   (!)pre-condition[TCSP]: item.Price == head.Price
   */
  tail.twin = item.twin
  /* @-Item.Splice#SpliceTwinsLeft:
   *   (?)post-condition[TUP]: TreeExcludesTwins(set, item)
   *   (?)post-condition[TCU]: TwinChainUnique(item)
   *   (?)post-condition[TCSP]: TwinChainSamePrice(item)
   *   (?)post-condition[LRGR]: TwinsProxied(item, proxy)
   */
  item.twin = head
}

/* @-Item.Link() */
func (ego *Item) Link(proxy *Item) (*Item) {
  /* @-Item.Link#Walk: loophead, predicate
   */
  for {
    /* @-Item.Link#Step:
     *   (?)pre-condition: ego.Price.Bound > proxy.Price.Base
     */
    ego.Price.Bound = proxy.Price.Base
		/* @-Item.Link#Enable: call
     *   (?)pre-condition[LRGR]: ego.index < proxy.index
     *   (?)pre-condition: ego.Price == ego.Price
     *   (?)post-condition[LRGR]: PurchaseStepwiseReachable(ego, proxy)
     */
    ego.Enable(proxy)
		/* @-Item.Link#LastTwin: predicate */
    if ego.twin == nil { break }
		/* @-Item.Link#NextTwin */
    ego = ego.twin
  }
		/* @-Item.Link#ReturnTail */
  return ego
}

type PriceScope interface {
  Fresh(*Item) (error)
  MatchProxy(*Item, *Item) (error)
}

// Note: unsafe for concurrent execution

/* @={I *Item}
 * @={N *Item}
 * @={T *Item}
 */
type PriceSpanSet struct {
  ItemIndex atomic.Int32
  roots []*Item
  all []*Item     /* @== as I */
  tree *interval.SearchTree[*Item,int]   /* @== ItemSet as N under PriceScope */
}

func (set *PriceSpanSet) Contains(item *Item) bool {
  _, found := set.tree.Find(item.Price.Base, item.Price.Bound)
  return found
}

func compare (a, b int) int { return a - b }
func compare_items (a, b *Item) int { if a.Price.Base == b.Price.Base { return a.Price.Bound - b.Price.Bound } else { return a.Price.Base - b.Price.Base; } }

/* @-NewPriceSpanSet(): alias func NewPriceSpanSet() (PriceSpanSet) { return PriceSpanSet{ tree: { make([]*Item, 0) } } } */
func NewPriceSpanSet() (PriceSpanSet) {
  return PriceSpanSet { roots: make([]*Item, 0), tree: interval.NewSearchTree[*Item,int](compare) }
}

/* @-Item#Constructor:
 *   (?)pre-condition[IIP]: (set.ItemIndex.Load() < math.MaxInt32) && (set.ItemIndex.Load() >= int32(0))
 */
func (set *PriceSpanSet) NewItem(margin int, price int) (*Item) {
  /* @-NewItem#ConstructorWatermark: call */
  prior := set.ItemIndex.Load()
  if prior == math.MaxInt32 { panic("Error: PriceSpanSet.ItemIndex overflow") }

  /* @-NewItem#Construct:
   *   (?)post-condition[IIP]: i.index <= set.ItemIndex.Load() && i.index >= prior
   *   (?)post-condition: i.Margin == margin && i.Price == PriceSpan{ price, math.MaxInt32 } && i.best == [3]int{ margin, 0, 0 } && len(i.in) == 0 && len(i.out) == 0 && i.twin == nil
   */
  i := &Item {
    /* @-NewItem#ConstructIndex: call */
		set.ItemIndex.Add(1) - 1,
		margin,
		PriceSpan { price, math.MaxInt32 },
		[3]int{ margin, 0, 0 },
		make([]*Item, 0),
		make([]*Item, 0),
		nil,
	}
  /* @-NewItem#UniverseAppend: call
   *   (+)state[I]: ∪ i
   */
  set.all = append(set.all, i)
  return i
}

func (item *Item) String() string {
	edges := func (v []*Item) string {
		if len(v) == 0 {
			return ""
		}
		list := make([]string, len(v))
		for i := range len(v) {
			list[i] = fmt.Sprintf("%v", v[i].index)
		}
		return strings.Join(list, ",")
  }
	return fmt.Sprintf("%v: %v @ %v in[%v] x out [%v]", item.index, item.Price, item.Margin, edges(item.in), edges(item.out))
}

type HypotheticalItem struct {
  index int32
  Price PriceSpan
}

func (set *PriceSpanSet) NewHypotheticalItem(price int) HypotheticalItem {
  return HypotheticalItem{ index: set.ItemIndex.Load(), Price: PriceSpan { price, math.MaxInt32 } }
}

// (?)post-condition: len(to.out) == 0 && slices.Contains(from.out, to) && slices.Contains(to.in, from)

/* @-Item.Enable():
 *   (?)pre-condition[DMSG]: len(to.out) == 0 && from != to && from.index < to.index && from.Price.Base < to.Price.Base
 *   (?)pre-condition[DMG]: DirectedMonotonicSubGraphs(from)
 *   (?)post-condition[DMG]: DirectedMonotonicSubGraphs(from)
 */
func (from *Item) Enable(to *Item) {
	/* @-Item.Enable#out: call */
	from.out = append(from.out, to)
	/* @-Item.Enable#in: call */
	to.in = append(to.in, from)
}

func (set *PriceSpanSet) String() string {
	lines := []string { fmt.Sprintf("• PriceSpanSet:") }
	set.tree.InOrderTraverse(func (base int, bound int, item *Item) error {
		lines = append(lines, fmt.Sprintf("• %v", item))
  	return nil
  })
	return strings.Join(lines, "\n")
}

/* @-PriceSpanSet.GraphWalk():
 *   (?)pre-condition: PriceBounds(input)
 *   (?)pre-condition: InputSizeBounds(set, input)
 *   (?)pre-condition: set.ItemIndex.Load() > 0 || len(set.tree.ItemSet) == 0
 *   (?)post-condition: PurchasableTriosOnPaths(set)
 *   (?)post-condition: AllPathTriosPurchasable(set)
 */
func (set *PriceSpanSet) GraphWalk(input []util.Input) {
	/* @-PriceSpanSet.GraphWalk#EachInput: loophead, predicate
   *   (?)pre-condition: TreeUniquePrices(set)
   *   (!)post-condition: TreeUniquePrices(set)
   */
	for _, i := range input {
		/* @-GraphWalk#AddItem: call */
    item := set.NewItem(i.Margin, i.Price)

		/* @-PriceSpanSet.GraphWalk#Step: call
     *   (!)pre-condition: TreeUniquePrices(set)
     */
    node := set.GraphStep(item)
		if node == nil {
			e := PriceScope(set).Fresh(item)
			/* @-PriceSpanSet.GraphWalk#InsertItemErrorCheck: predicate */
			if e != nil {
				/* @-PriceSpanSet.GraphWalk#InsertItemError: panic */
				panic(e)
      }
		} else {
      /* @-PriceSpanSet.GraphWalk#elide:
       *   (!)pre-condition[TUP]: node.twin == item
       *   (?)post-condition[TUP]: !slices.Contains(set.tree.ItemSet, item)
       */
       _ = item
    }
	}

	fmt.Println(set)
}

/* @-PriceSpanSet.Fresh() */
func (set *PriceSpanSet) Fresh(item *Item) (error) {
	/* @-PriceSpanSet.Fresh#Insert: call
   *   (!)pre-condition[TUP]: !set.Contains(item)
   *   (?)post-condition[TUP]: slices.Contains(set.tree.ItemSet, item)
   *   (?)post-condition[MPCI]: MatchIncludes(set, item)
   *   (+)state[N]: ∪ item
   */
	return set.tree.Insert(item.Price.Base, item.Price.Bound, item)
}

// • price bound can't increase
// • every existing item is either ignored (unmatched), base-matched at item.Price.Base, or proxied at item.Price.Base
// •

/* induction goals:
   • MPCI
   • MPCG
   • GRGU
 */

/* @-PriceSpanSet.GraphStep():
 *   (?)pre-condition: func () bool { for _, e := range set.tree.ItemSet { for t := e; t != nil; t = t.twin { if t.index >= item.index { return false } } }; return true }()
 *   (?)pre-condition: DirectedMonotonicGraph(set)
 *   (?)post-condition: LesserRootsGraphReachable(set, item)
 *   (?)post-condition: GreaterRootsGraphUnreachable(set, item)
 */
func (set *PriceSpanSet) GraphStep(item *Item) (*Item) {
	var node *Item
	/* @-PriceSpanSet.GraphStep#TreeQuery: call
   *   (?)post-condition[TUP]: !slices.Contains(matches, item)
   *   (?)post-condition: MinimalPriceCompatibleIntersection(set, matches, item)
   */
	matches, matched := set.tree.AllIntersections(item.Price.Base, item.Price.Base)
	/* @-PriceSpanSet.GraphStep#TreeQueryMatchCheck: predicate */
	if matched {
	  /* @-PriceSpanSet.GraphStep#EachMatch: loophead, predicate
     *   (?)pre-condition[LRGR]: ExcludedReachProxy(set, matches, item)
     */
		for _, match := range matches {
	    /* @-PriceSpanSet.GraphStep#FauxMatch: predicate */
      if match.Price.Bound == item.Price.Base {
	      /* @-PriceSpanSet.GraphStep#SkipFauxMatch:
         *   (?)pre-condition[GRGU]: func () bool { if len(match.out) == 0 { return false }; for _, o := range match.out { if o.Price.Base == item.Price.Base { return true } }; return false }()
         */
        continue
      }
			/* @-PriceSpanSet.GraphStep#MatchBoundCheck: predicate
       *   (?)pre-condition[GRGU]: !slices.Contains(set.roots, match) || match.Price.Base < item.Price.Base
       */
			if item.Price.Base > match.Price.Base {
				e := error(nil)
				/* @-PriceSpanSet.GraphStep#MatchProxy: call
         *   (?)post-condition: match.Price.Bound == item.Price.Base
         *   (?)post-condition: HasAlterEgoAt(set, match)
         */
				e = PriceScope(set).MatchProxy(match, item)
				/* @-PriceSpanSet.GraphStep#MatchReplaceErrorCheck: predicate */
				if e != nil {
					/* @-PriceSpanSet.GraphStep#MatchReplaceError: panic */
					panic(e)
				}
			} else {
				/* @-PriceSpanSet.GraphStep#CheckSameInterval: predicate
         *   (?)pre-condition: item.Price.Base == match.Price.Base
         *   (!)pre-condition[TCSP]: TwinChainSamePrice(match)
         *   (!)pre-condition[TCSP]: item.twin == nil
         */
        if item.Price.Bound == match.Price.Bound {
					/* @-PriceSpanSet.GraphStep#Splice: call
           *   (!)pre-condition[TCU]: FreshTwinExtendsChain(match, item)
           *   (?)post-condition[TUP]: HasAlterEgo(match, item)
           *   (?)post-condition[TCU]: TwinChainIndexBounded(set, -1, match)
           *   (?)post-condition[TCU]: TwinChainUnique(match)
           *   (?)post-condition[TCSP]: TwinChainSamePrice(match)
           *   (?)post-condition[TUP]: TreeExcludesTwins(set, match)
           *   (+)state[T]: ∪ item
           */
          match.Splice(item, item)
					/* @-PriceSpanSet.GraphStep#ElideSameInterval */
					node = match
				}
        /* @-PriceSpanSet.GraphStep#ForkSamePriceCheck: predicate, call */
        if len(match.in) == 0 {
          /* @-PriceSpanSet.GraphStep#ForkSamePriceRoot: call
           *   (?)pre-condition: slices.Contains(set.roots, match)
           *   (?)pre-condition[MPCI]: len(set.roots) > 0 && set.roots[len(set.roots)-1] == match
           *   (?)post-condition[MPCI]: PriceCompatibleRange(set, item, item.Price.Base + 1, item.Price.Bound)
           *   (?)post-condition[MPCG]: func() bool { for _, i := range set.roots { if @PriceCompatibleItems(i, item) { return false } }; return true }()
           */
          set.roots = append(set.roots, item)
        } else {
          /* @-PriceSpanSet.GraphStep#ForkSamePrice: loophead, predicate, call */
				  for ego := match; ego != nil; ego = ego.twin {
            for _, predecessor := range ego.in {
              /* @-PriceSpanSet.GraphStep#ForkSamePriceSuperseded */
              ptail := predecessor.out[len(predecessor.out)-1]
              if ptail.Price.Bound < item.Price.Base {
                /* @-PriceSpanSet.GraphStep#SkipSamePriceProxied:
                 *   (?)pre-condition[LRGR]: predecessor.Price.Bound < item.Price.Base && ptail.Price.Base < item.Price.Base
                 *   (?)pre-condition[LRGR]: PurchaseStepwiseReachable(ptail, item)
                 */
                continue /* already proxied by a subsequent lesser price */
              }
              if ptail == ego {
                /* @-PriceSpanSet.GraphStep#BranchSamePrice: call
                 *   (!)pre-condition[LRGR]: predecessor.Price.Base < item.Price.Base && predecessor.Price.Bound == item.Price.Base
                 */
                predecessor.Enable(item)
              } else { /* `ptail` must be some other ego */
                /* @-PriceSpanSet.GraphStep#SkipSamePriceDuplicate:
                 *   (?)pre-condition[LRGR]: HasAlterEgo(match, ptail)
                 */
                _ = ptail
              }
            }
          }
        }
			}
		}
	} else {
	  /* @-PriceSpanSet.GraphStep#Root:
     *   (?)pre-condition[MPCI]: len(set.roots) == 0 || set.roots[len(set.roots)-1].Price.Base > item.Price.Base
     *   (?)post-condition[MPCI]: PriceCompatibleRange(set, item, item.Price.Base + 1, item.Price.Bound)
     *   (?)post-condition[MPCG]: func() bool { for _, i := range set.roots { if @PriceCompatibleItems(i, item) { return false } }; return true }()
     *   (?)pre-condition[GRGU]: func () bool { for _, root := range set.roots { if root.Price.Base <= item.Price.Base { return false } }; return true }()
     */
    set.roots = append(set.roots, item)
  }

	/* @-PriceSpanSet.GraphStep#return:
   *   (?)pre-condition[TUP]: node != nil || !set.Contains(item)
   */
	return node;
}

// possible induction goal: MinimalPriceCompatibleGraph (MPCG)
// - no forward edges
// - all paths price compatible
// - always price compatible on paths

/* @-PriceSpanSet.MatchProxy():
 *   (!)pre-condition[TCSP]: TwinChainSamePrice(match)
 */
func (set *PriceSpanSet) MatchProxy(match *Item, proxy *Item) (error) {
	/* @-PriceSpanSet.MatchProxy#Delete: call
   *   (!)pre-condition[TCU]: TwinChainIndexBounded(set, -1, match)
   *   (!)pre-condition[TUP]: TreeExcludesTwins(set, match)
   *   (+)state[N]: \ match
   */
	set.tree.Delete(match.Price.Base, match.Price.Bound)
  /* @-PriceSpanSet.MatchProxy#FindSameInterval: call */
	existing, twinned := set.tree.Find(match.Price.Base, proxy.Price.Base)
	/* @-PriceSpanSet.MatchProxy#Link: call
   *   (?)pre-condition: !twinned || (existing.Price.Bound - existing.Price.Base) < (match.Price.Bound - match.Price.Base)
   *   (?)pre-condition[TCU]: !twinned || @MutuallyExclusivePriceChains(existing, match)
   *   (!)pre-condition[LRGR]: TwinsProxied(existing, proxy)
   *   (?)post-condition[TCSP]: TwinChainSamePrice(match)
   *   (?)post-condition[LRGR]: TwinsProxied(match, proxy)
   */
  matchtail := ProxyPrice(match).Link(proxy)
	/* @-PriceSpanSet.MatchProxy#CheckSameInterval: predicate
   *   (!)pre-condition[TCSP]: TwinChainSamePrice(match)
   */
	if twinned {
		/* @-PriceSpanSet.MatchProxy#Splice: call
     *   (+)state[T]: ∪ match
     */
    Entwine(existing).Splice(match, matchtail)
	  /* @-PriceSpanSet.MatchProxy#ElideSameInterval */
    return nil
  } else {
		/* @-PriceSpanSet.MatchProxy#Insert: call
     *   (?)pre-condition[TUP]: !set.Contains(match)
     *   (?)pre-condition[TUP]: match != proxy
     *   (+)state[N]: ∪ match
     */
		return set.tree.Insert(match.Price.Base, match.Price.Bound, match)
	}
}

func Purchasable(items []*Item) (bool) {
	switch len(items) {
	case 0:
		return false
	case 1:
		return true
	}

	indexer := func(a, b *Item) int {
		return int(a.index - b.index)
	}
	slices.SortFunc(items, indexer)
	for i := range (len(items)-1) {
		if items[i].Price.Base >= items[i+1].Price.Base {
			return false
		}
	}
  return true
}
