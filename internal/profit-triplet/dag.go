package dag

import (
  "fmt"
 "strings"
 "slices"
 "math"
 "sync/atomic"
 "github.com/rdleal/intervalst/interval"
 _ "gopkg.in/fifo.v0"

	"demo/profit-triplet/pkg/util"
)

 // "github.com/go-fifo/fifo"

/* @+atomic.Int32: data-model n int32 */
/* @+atomic.Int32.Load(): alias func (i *atomic.Int32) Load() int32 { return i.n } */
/* @+atomic.Int32.Add(n int32): alias func (i *atomic.Int32) Add(n int32) int32 { i.n += n; return i.n } */

/* @+vector.Vector: data-model elements []*Item */
/* @+vector.Vector.PushBack: alias func (v *vector.Vector[*Item]).PushBack(item *Item) { v.elements = append(v.elements, item) } */
/* @+vector.Vector.At: alias func (v *vector.Vector[*Item]).At(int i) (*Item) { return v.elements[i] } */
/* @+vector.Vector.Size: alias func (v *vector.Vector[*Item]).Size() (int) { return len(v.elements) } */
/* @+vector.Vector.Empty: alias func (v *vector.Vector[*Item]).Empty() (bool) { return len(v.elements) == 0 } */
/* @+vector.NewVector(): alias func NewVector() (*vector.Vector[*Item]) { return &vector.Vector[*Item]{ make([]*Item, 0) } } */

/* @+interval.SearchTree: data-model ItemSet []*Item */
/* @+interval.SearchTree.AllIntersections(a int, b int):
 * alias func (tree *interval.SearchTree) AllIntersections(a, b int) ([]*Item, bool) { m := make([]*Item, 0); for _, i := range tree.ItemSet { if i.Price.Base <= b && i.Price.Bound >= a { m = append(m, i) } } return m, len(m) > 0 }
 * @+interval.SearchTree.Find(a int, b int):
 * alias func (tree *interval.SearchTree) Find(a, b int) (*Item, bool) { for _, i := range tree.ItemSet { if i.Price.Base == a && i.Price.Bound == b { return i, true } } return nil, false }
 * @+interval.SearchTree.Insert(a int, b int, i *Item):
 * (?)pre-condition: i.Price.Base < i.Price.Bound && a == i.Price.Base && b == i.Price.Bound
 * (?)pre-condition: !slices.Contains(tree.ItemSet, i)
 * alias func (tree *interval.SearchTree) Insert(a, b int, i *Item) (error) { tree.ItemSet = append(tree.ItemSet, i); slices.SortFunc(tree.ItemSet, compare_items) return nil }
 * @+interval.SearchTree.Delete(a int, b int):
 * alias func (tree *interval.SearchTree) Delete(a, b int) { tree.ItemSet = slices.Collect( func(yield func(*Item) bool) { for _, i := range tree.ItemSet { if i.Price.Base != a || i.Price.Bound != b { if !yield(i) { return } } } } ) }
 */

/******** TwinPricesUniquelyChained [TPUC] */

/* @^TwinPricesUniquelyChained[TPUC](set *PriceSpanSet) <=> {TUP,TCSP,TCU}:
       all := make([]*Item, set.ItemIndex.Load())
       for _, node := range set.tree.ItemSet {
         for ego := node; ego != nil; ego = ego.twin {
           if ego.Price != node.Price { return false }
           if ego.index < node.index { return false }
           if all[ego.index] { return false }
           all[ego.index] = ego
         }
       }

       for i := range all {
         if !i { return false }
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
       for _, i := set.tree.ItemSet {
         if !TwinChainSamePrice(i) { return false }
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
       for _, i := set.tree.ItemSet {
         if !TwinChainUnique(i) { return false }
       }
       return true
 */

/* @^TwinChainUnique(node *Item):
        unique := make(map[int]*Item)
        for ego := node; ego != nil; ego = ego.twin {
          if unique[ego.index] { return false }
          unique[ego.index] = ego
        }
        return true
 */

/* @^TwinChainIndexBounded(set *PriceSpanSet, index int, node *Item):
       if !slices.Contains(set.tree.ItemSet, node) { return false }
       for ego := node; ego != nil; ego = ego.twin {
         if (ego.index >= index) { return false }
       }
       return true
 */

/* @^MutuallyExclusivePriceChains(existing *Item, joining *Item):
       for x := existing; x != nil; x = x.twin {
         if (x.Price == joining.Price) return false
       }
       for j := joining; j != nil; j = j.twin {
         if (j.Price == existing.Price) return false
       }
 */

/* @^FreshTwinExtendsChain(node *Item, fresh *Item):
       for ego := node; ego != nil; ego = ego.twin {
         if (ego.index >= fresh.index) { return false }
       }
       return true
 */

/******** [LRGR] */

/* @^LesserRootsGraphReachable[LRGR](set *PriceSpanSet, i *Item):
       for _, root := range set.tree.roots {
         if root.index < i.index && root.Price.Base < i.Price.Base {
           if !GraphReachable(root, i) { return false }
         }
       }
       return true
 */

/* @^TwinPrefixProxy(node *Item, sentinel *Item, frontier *Item):
       if node == nil { return true }
       for ego := node; ego != sentinel; ego = ego.twin {
         if !slices.Contains(ego.out, frontier) { return false }
         if ego.Price.Base >= frontier.Price.Base || ego.Price.Bound != frontier.Price.Base { return false }
       }
 */

/* @^TwinChainProxy(joined *Item, proxied *Item, frontier *Item):
       for ego := (joined == nil ? proxied : joined); ego != nil; ego = ego.twin {
         if !slices.Contains(ego.out, item) { return false }
         if ego.Price.Base >= frontier.Price.Base || ego.Price.Bound != frontier.Price.Base { return false }
       }
 */

/* @^GraphReachable(a *Item, b *Item):
       q := fifo.New[*Item](1024)
       for _, o := range a.out { q.Enqueue(o) }
       visited := make(map[int]*Item)
       for q.Len() > 0 {
         i := q.Dequeue()
         if visited[i.index] { continue }
         if i == b { return true }
         visited[i.index] = i
         if i.index < b.index && i.Price.Base < b.Price.Base {
           for _, o := range i.out { q.Enqueue(o) }
         }
       }
       return false
*/

/* @^ExcludedReachProxy(PriceSpanSet *set, proxies []*Item, threshold int):
       for _, i := range set.tree.ItemSet {
         if i.Price.Base < threshold {
           if !ProxyReachable(i, proxies) { return false }
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

/******** GreaterRootsGraphUnreachable [GRGU] */

/* @^GreaterRootsGraphUnreachable[GRGU](set *PriceSpanSet, i *Item):
       for _, root := range set.tree.roots {
         if root.index >= i.index || root.Price.Base >= i.Price.Base {
           if GraphReachable(root, i) { return false }
         }
       }
       return true
 */

/******** MinimalPriceCompatibleIntersection [MPCI] */

/* @^MinimalPriceCompatibleIntersection[MPCI](set *PriceSpanSet, matches []*Item, item *Item):
       for _, i := range set.tree.ItemSet {
         if slices.Contains(matches, i) {
           if !PriceCompatible(i, item) { return false }
           for _, o := range i.out {
             if PriceCompatible(o, item) { return false }
           }
         } else if !PriceCompatible(i, item) {
           if ProxyReachable(i, matches) { return false }
         } else if !ProxyReachable(i matches) {
           return false
         }
       }
 */

/******** DirectedMonotonicGraph [DMG]  */

/* @^DirectedMonotonicGraph[DMG](roots []*Item):
       visited := make(map[int]*Item)
       q := fifo.New[*Item](1024)
       for _, root := range roots { q.Enqueue(root) }
       for q.Len() > 0 {
         i := q.Dequeue()
         for _, o := range i.out {
           if visited[o.index] { continue }
           if (i.index >= o.index) { return false }
           if (i.Price.Base >= o.Price.Base) { return false }
           visited[o.index] = o
           q.Enqueue(o)
         }
       }
       return true
 */

/* @^DirectedMonotonicSubGraphs(pivot *Item):
       visited := make(map[int]*Item)
       q := fifo.New[*Item](1024)
       q.Enqueue(pivot)
       for q.Len() > 0 {
         i := q.Dequeue()
         for _, o := range i.out {
           if visited[o.index] { continue }
           if (i.index >= o.index) { return false }
           if (i.Price.Base >= o.Price.Base) { return false }
           visited[o.index] = o
           q.Enqueue(o)
         }
       }

       visited := make(map[int]*Item)
       q.Enqueue(pivot)
       for q.Len() > 0 {
         o := q.Dequeue()
         for _, i := range o.in {
           if visited[i.index] { continue }
           if (o.index <= i.index) { return false }
           if (o.Price.Base <= i.Price.Base) { return false }
           visited[i.index] = i
           q.Enqueue(i)
         }
       }

       return true
 */

/******** [untagged]  */

/* @^PriceBounds(input []util.Input):
       for _, i := range input {
         if i.Price <= 0 || i.Price >= 10000 {
           return false
         }
       }; return true
*/

/* @^InputSizeBounds(set *PriceSpanSet, input []util.Input):
       (len(input) < math.MaxInt32) &&
       (set.ItemIndex.Load() < (math.MaxInt32 - int32(len(input)))) &&
       (set.ItemIndex.Load() >= int32(0))
*/

/* @^PurchasableTrio(a *Item, b *Item, c *Item) {
       return a.index < b.index && b.index < c.index && a.Price.Base < b.Price.Base && b.Price.Base < c.Price.Base
     }
 */

/* @^PurchasableTriosOnPaths(set *PriceSpanSet):
	     indexer := func(a, b *Item) int {
	     	return int(a.index - b.index)
	     }

       for _, i := range set.all[ :len(set.all)-2] {
         for _, j := range set.all[i+1 : len(set.all)-1] {
           for _, k := range set.all[j+1 : len(set.all)] {
             trio := { i, j, k }
             slices.SortFunc(trio, indexer)
             c := GraphReachable(trio[0], trio[1]) && GraphReachable(trio[1], trio[2])
             if c != PurchasableTrio(trio[0], trio[1], trio[2]) { return false }
           }
         }
       }
 */

/* @^AllPathTriosPurchasable(set *PriceSpanSet):
       for item := range set.all {
         for _, i := range item.in {
           i.index < item.index && i.Price.Base < item.Price.Base
         }
         for _, o := range item.out {
           o.index > item.index && o.Price.Base > item.Price.Base
         }
       }
 */

/* @^OrderedIntersection(matches []*Item):
       c := len(matches)
       if c < 2 { return true }
       for _, i := range (c - 1) {
         if matches[i].Price.Base > matches[i+1].Price.Base { return false }
         if matches[i].Price.Base == matches[i+1].Price.Base {
           if matches[i].Price.Bound >= matches[i+1].Price.Bound { return false }
         }
       }
       return true
*/

/* @^GraphAnyReachable(a *Item, b []*Item):
       q := fifo.New[*Item](1024)
       for _, o := range a.out { q.Enqueue(o) }
       visited := make(map[int]*Item)
       for q.Len() > 0 {
         i := q.Dequeue()
         if slices.Contains(b, i) { return true }
         visited[i.index] = i
         for _, o := range i.out {
           if !visited[o.index] { q.Enqueue(o) }
         }
       }
       return false
*/

 // TODO: heavy if executed, not really necessary
/* ExcludedIncompatible(PriceSpanSet *set, proxies []*Item, threshold int):
       for _, i := range set.tree.ItemSet {
         if slices.Contains(proxies, i) { continue }
         if i.Price.Base > threshold {
           if GraphAnyReachable(i, proxies) { return false }
         }
       }
 */

/* @^JoinHomogeneousLarger(existing *Item, joining *Item):


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

/* @^ProxyReachable(item *Item, proxies []*Item):
       if slices.Contains(proxies, item) { return true }
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

/* @^PriceCompatible(a *Item, b *Item):
       return a.index < b.index && a.Price.Base < b.Price.Base
 */

/* @^PriceCompatible(a *Item, b *HypotheticalItem):
       return a.index < b.index && a.Price.Base < b.Price.Base
 */

/* @^PriceCompatible(set *PriceSpanSet, a *Item, int plow, int phigh) {
       if phigh > 10000 { phigh = 10000 }
       return a.index < set.ItemIndex.Load() && a.Price.Base < plow && a.price.Bound > phigh
 */


type PriceSpan struct {
	Base  int   // current selling price
	Bound int   // end of the span covered by the enclosing item
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
	twin *Item
}

/* Note: unsafe for concurrent execution */
type PriceSpanSet struct {
  ItemIndex atomic.Int32
  roots []*Item   /* for proof purposes only, can be omitted for release build */
  all []*Item     /* for proof purposes only, can be omitted for release build */
  tree *interval.SearchTree[*Item,int]
}

func compare (a, b int) int { return a - b }
func compare_items (a, b *Item) int { if a.Price.Base == b.Price.Base { return a.Price.Bound - b.Price.Bound } else { return a.Price.Base - b.Price.Base; } }

/* @-NewPriceSpanSet(): alias func NewPriceSpanSet() (PriceSpanSet) { return PriceSpanSet{ tree { make([]*Item, 0) } } } */
func NewPriceSpanSet() (PriceSpanSet) {
  return PriceSpanSet { roots: make([]*Item, 0), tree: interval.NewSearchTree[*Item,int](compare) }
}

/* @-Item#Constructor:
 * (?)pre-condition: (set.ItemIndex.Load() < math.MaxInt32) && (set.ItemIndex.Load() >= int32(0))
 */
func (set *PriceSpanSet) NewItem(margin int, price int) (Item) {
  /* @-Item#ConstructorWatermark: call */
  prior := set.ItemIndex.Load()
  if prior == math.MaxInt32 { panic("Error: PriceSpanSet.ItemIndex overflow") }

  /* @-Item#Construct:
   * (?)post-condition: i.index <= set.ItemIndex.Load() && i.index > prior && i.Margin == margin && i.Price == PriceSpan{ price, math.MaxInt32 } && i.best == [3]int{ margin, 0, 0 } && len(i.in) == 0 && len(i.out) == 0 && i.twin == nil
   */
  i := Item {
    /* @-Item#ConstructIndex: call */
		set.ItemIndex.Add(1),
		margin,
		PriceSpan { price, math.MaxInt32 },
		[3]int{ margin, 0, 0 },
		make([]*Item, 0),
		make([]*Item, 0),
		nil,
	}
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
  PriceSpan Price
}

func (set *PriceSpanSet) NewHypotheticalItem(price int) {
  return HypotheticalItem{ index: set.ItemIndex.Load(), Price: { price, math.MaxInt32 } }
}

// (?)post-condition: len(to.out) == 0 && slices.Contains(from.out, to) && slices.Contains(to.in, from)

/* @-Item.Link():
 * (?)pre-condition: len(to.out) == 0 && from != to && from.index < to.index && from.Price.Base < to.Price.Base
 * (?)pre-condition: DirectedMonotonicSubGraphs(from)
 * (?)post-condition: DirectedMonotonicSubGraphs(from)
 */
func (from *Item) Link(to *Item) {
	/* @-Item.Link#out: call */
	from.out = append(from.out, to)
	/* @-Item.Link#in: call */
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

// TODO: possible to allow non-empty PSS, but requires that only this function ever changes a PSS

/* @-PriceSpanSet.GraphWalk()
 * (?)pre-condition: PriceBounds(input)
 * (?)pre-condition: InputSizeBounds(set, input)
 * (?)pre-condition: len(set.tree.ItemSet) == 0
 * (?)post-condition: PurchasableTriosOnPaths(set)
 * (?)post-condition: AllPathTriosPurchasable(set)
 */
func (set *PriceSpanSet) GraphWalk(input []util.Input) {
	/* @-PriceSpanSet.GraphWalk#EachInput: loophead, predicate */
	for _, i := range input {
		/* @-GraphWalk#AddItem: call */
		item := set.NewItem(i.Margin, i.Price)

		/* @-PriceSpanSet.GraphWalk#Step: predicate, call */
    node := set.GraphStep(&item)
		if node {
      /* @-PriceSpanSet.GraphWalk#elide:
       * (!)pre-condition[TPU]: node->twin == item
       * (?)post-condition[TPU]: !slices.Contains(set.tree.ItemSet, item)
       */
       _ = item
		} else {
			/* @-PriceSpanSet.GraphWalk#InsertItem: call
       * (!)pre-condition[TPU]: set.tree.Find(item.Price.Base, item.Price.Bound) == nil
       * (?)post-condition[TPU]: slices.Contains(set.tree.ItemSet, item)
       * (?)post-condition: func () bool { for i := item.Price.Base; i < 10000; i++ { x, _ := set.tree.AllIntersections(i, i); if !slices.Contains(x, item) { return false } }; return true }()
       */
			e := set.tree.Insert(item.Price.Base, item.Price.Bound, &item)
			/* @-PriceSpanSet.GraphWalk#InsertItemErrorCheck: predicate */
			if e != nil {
				/* @-PriceSpanSet.GraphWalk#InsertItemError: panic */
				panic(e)
			}
    }
	}

	fmt.Println(set)
}

// • price bound can't increase
// • every existing item is either ignored (unmatched), base-matched at item.Price.Base, or proxied at item.Price.Base
// •

/* induction goals:
   • MPCI
   • MPCG
   • GRGU
 */

/* @-PriceSpanSet.GraphStep()
 * (?)pre-condition: func () bool { for _, e := range set.tree.ItemSet { for t := e; t != nil; t = t.twin { if t.index >= item.index { return false } } } return true }()
 * (?)pre-condition: DirectedMonotonicGraph(set.roots)
 * (?)post-condition: LesserRootsGraphReachable(set, item)
 * (?)post-condition: GreaterRootsGraphUnreachable(set, item)
 */
func (set *PriceSpanSet) GraphStep(item *Item) (*Item) {
	node := nil
	/* @-PriceSpanSet.GraphStep#TreeQuery: call
   * (?)post-condition: MinimalPriceCompatibleIntersection(set, matches, item)
   */
	matches, matched := set.tree.AllIntersections(item.Price.Base, item.Price.Base)
	/* @-PriceSpanSet.GraphStep#TreeQueryMatchCheck: predicate */
	if matched {
	  /* @-PriceSpanSet.GraphStep#EachMatch: loophead, predicate
     * (?)pre-condition[LRGR]: ExcludedReachProxy(set, matches, item.Price.Base)
     */
		for _, match := range matches {
	    /* @-PriceSpanSet.GraphStep#FauxMatch: predicate */
      if match.Price.Bound == item.Price.Base {
	      /* @-PriceSpanSet.GraphStep#SkipFauxMatch:
         * (?)pre-condition[GRGU]: func () bool { if len(match.out) == 0 { return false }; for _, o := range match.out { if o.Price.Base == item.Price.Base { return true } }; return false }()
         */
        continue
      }
			/* @-PriceSpanSet.GraphStep#MatchBoundCheck: predicate
       * (?)pre-condition[GRGU]: !slices.Contains(set.tree.roots, match) || match.Price.Base < item.Price.Base
       */
			if item.Price.Base > match.Price.Base {
				e := error(nil)
				/* @-PriceSpanSet.GraphStep#MatchProxy: call
         * (?)post-condition: match.Price.Bounds == item.Price.Base
         * (?)post-condition: HasAlterEgo(set.tree.Find(match.Price.Base, match.Price.Bounds), match)
         */
				node, e = set.MatchProxy(match, item.Price.Base)
				/* @-PriceSpanSet.GraphStep#MatchReplaceErrorCheck: predicate */
				if e != nil {
					/* @-PriceSpanSet.GraphStep#MatchReplaceError: panic */
					panic(e)
				}
			} else {
				/* @-PriceSpanSet.GraphStep#CheckSameInterval: predicate
         * (?)pre-condition: item.Price.Base == match.Price.Base
         * (!)pre-condition[TCSP]: TwinChainSamePrice(match)
         * (!)pre-condition[TCSP]: item.twin == nil
         */
        if item.Price.Bound == match.Price.Bound {
					/* @-PriceSpanSet.GraphStep#SpliceTwinsRight:
           * (!)pre-condition[TCU]: FreshTwinExtendsChain
           */
          item.twin = match.twin
					/* @-PriceSpanSet.GraphStep#SpliceTwinsLeft:
           * (?)post-condition[TCU]: TwinChainIndexBounded(set, set.ItemIndex.Load(), match)
           * (?)post-condition[TCU]: TwinChainUnique(match)
           * (?)post-condition[TCSP]: TwinChainSamePrice(match)
           */
          match.twin = item
					/* @-PriceSpanSet.GraphStep#ElideSameInterval */
					node = match
				}
        /* @-PriceSpanSet.GraphForkSamePriceCheck: predicate, call */
        if len(match.in) == 0 {
          /* @-PriceSpanSet.GraphForkSamePriceRoot: call
           * (?)pre-condition: slices.Contains(set.roots, match)
           * (?)pre-condition[MPCI]: len(set.roots) > 0 || set.roots[len(set.roots)-1] == match
           * (?)post-condition[MPCI]: PriceCompatible(item, item.Price.Base + 1, item.Price.Bound)
           * (?)post-condition[MPCG]: func() bool { for _, i := range set.tree.roots { if (PriceCompatible(i, item) { return false } } }()
           */
          set.roots = append(set.roots, item)
        } else {
          /* @-PriceSpanSet.GraphForkSamePrice: loophead, predicate, call */
				  for ego := match; ego != nil; ego = ego.twin {
            for _, predecessor := range ego.in {
              /* @-PriceSpanSet.GraphForkSamePriceSuperseded: predicate */
              ptail := predecessor.out[len(predecessor.out)-1]
              if ptail.Price.Bound < item.Price.Base {
                /* @-PriceSpanSet.GraphSkipSamePriceProxied:
                 * (?)pre-condition[LRGR]: predecessor.Price.Bound < item.Price.Base && ptail.Price.Base < item.Price.Base
                 * (?)pre-condition[LRGR]: GraphReachable(ptail, item)
                 */
                continue /* already proxied by a subsequent lesser price */
              }
              if ptail == ego {
                /* @-PriceSpanSet.GraphBranchSamePrice: call
                 * (!)pre-condition[LRGR]: predecessor.Price.Base < item.Price.Base && predecessor.Price.Bound == item.Price.Base
                 */
                predecessor.Link(item)
              } else { /* `ptail` must be some other ego */
                /* @-PriceSpanSet.GraphSkipSamePriceDuplicate
                 * (?)pre-condition[LRGR]: HasAlterEgo(match, ptail)
                 */
                _ = ptail
              }
            }
          }
        }
			}
		}
	} else {
	  /* @-PriceSpanSet.GraphStep#Root
     * (?)pre-condition[MPCI]: len(set.roots) == 0 || set.roots[len(set.roots)-1].Price.Base > item.Price.Base
     * (?)post-condition[MPCI]: PriceCompatible(item, item.Price.Base + 1, item.Price.Bound)
     * (?)post-condition[MPCG]: func() bool { for _, i := range set.tree.roots { if (PriceCompatible(i, item) { return false } } }()
     * (?)pre-condition[GRGU]: func () bool { for _, root := range set.tree.roots { if root.Price.Base <= item.Price.Base { return false } }; return true }()
     */
    set.roots = append(set.roots, item)
  }

	/* @-PriceSpanSet.GraphStep#return */
	return node;
}

// possible induction goal: MinimalPriceCompatibleGraph (MPCG)
// - no forward edges
// - all paths price compatible
// - always price compatible on paths

/* @-PriceSpanSet.MatchProxy()
 * (!)pre-condition[TCSP]: TwinChainSamePrice(match)
 */
func (set *PriceSpanSet) MatchProxy(match *Item, proxy *Item) (*Item, error) {
	/* @-PriceSpanSet.MatchProxy#Delete call
   * (!)pre-condition[TCU]: TwinChainIndexBounded(set, set.ItemIndex.Load(), match)
   * (!)pre-condition[TUP]: TreeExcludesTwins(set, match)
   */
	set.tree.Delete(match.Price.Base, match.Price.Bound)
	/* @-PriceSpanSet.MatchProxy#TruncateNode: call */
  match.Price.Bound = proxy.Price.Base
  /* @-PriceSpanSet.MatchProxy#FindSameInterval: call */
	existing, twinned := set.tree.Find(match.Price.Base, proxy.Price.Base)
  var ego *Item
  /* @-PriceSpanSet.MatchProxy#TruncateTwins: loophead, predicate
   * (?)pre-condition: (existing.Price.Bound - existing.Price.Base) < (match.Price.Bound - match.Price.Base)
   * (?)pre-condition[TCU]: MutuallyExclusivePriceChains(existing, match)
   * (!)pre-condition[LRGR]: TwinPrefixProxy(existing, match, proxy)
   * (?)post-condition[TCSP]: TwinChainSamePrice(match)
   * (?)post-condition[LRGR]: TwinChainProxy(existing, match, proxy)
   */
  for ego = match; ego.twin != nil; ego = ego.twin {
    /* @-PriceSpanSet.MatchProxy#Truncate
     * (?)pre-condition: ego.Price.Bound > proxy.Price.Base
     */
    ego.twin.Price.Bound = proxy.Price.Base
		/* @-PriceSpanSet.MatchProxy#MatchLink: call
     * (?)pre-condition[LRGR]: ego.index < proxy.index
     * (?)pre-condition: ego.Price == match.Price
     * (?)post-condition[LRGR]: GraphReachable(ego, proxy)
     */
    ego.Link(proxy)
  }
	/* @-PriceSpanSet.MatchProxy#CheckSameInterval: predicate
   * (!)pre-condition[TCSP]: TwinChainSamePrice(match)
   * (!)pre-condition[TCSP]: TwinChainSamePrice(existing)
   * (!)pre-condition[TCSP]: existing.Price == match.Price
   */
	if twinned {
    /* @-PriceSpanSet.MatchProxy#SpliceTwinsRight
     * (!)pre-condition[TCU]: TwinChainUnique(existing)
     * (!)pre-condition[TCU]: TwinChainUnique(match)
     * (!)pre-condition[TCU]: TwinChainIndexBounded(set, match.index, existing)
     */
    ego.twin = existing.twin
    /* @-PriceSpanSet.MatchProxy#SpliceTwinsLeft:
     * (?)post-condition[TUP]: TreeExcludesTwins(set, existing)
     * (?)post-condition[TCU]: TwinChainUnique(existing)
     * (?)post-condition[TCSP]: TwinChainSamePrice(existing)
     */
    existing.twin = match
	  /* @-PriceSpanSet.MatchProxy#ElideSameInterval */
    return existing, nil
  } else {
		/* @-PriceSpanSet.MatchProxy#Insert: call
     * (?)pre-condition[TUP]: set.tree.Find(match.Price.Base, match.Price.Bound) == nil
     */
		return nil, set.tree.Insert(match.Price.Base, match.Price.Bound, match)
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
