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

/* @+interval.SearchTree: data-model EverInserted []*Item */
/* @+interval.SearchTree.AllIntersections(a int, b int):
 * alias func (tree *interval.SearchTree) AllIntersections(a, b int) ([]*Item, bool) { m := make([]*Item, 0); for _, i := range tree.EverInserted { if i.Price.Base <= b && i.Price.Bound >= a { m = append(m, i) } } return m, len(m) > 0 }
 * @+interval.SearchTree.Find(a int, b int):
 * alias func (tree *interval.SearchTree) Find(a, b int) (*Item, bool) { for _, i := range tree.EverInserted { if i.Price.Base == a && i.Price.Bound == b { return i, true } } return nil, false }
 * @+interval.SearchTree.Insert(a int, b int, i *Item):
 * alias func (tree *interval.SearchTree) Insert(a, b int, i *Item) (error) { tree.EverInserted = append(tree.EverInserted, i); slices.SortFunc(tree.EverInserted, compare_items) return nil }
 * (?)pre-condition: i.Price.Base < i.Price.Bound && a == i.Price.Base && b == i.Price.Bound
 * (?)pre-condition: !slices.Contains(tree.EverInserted, i)
 * @+interval.SearchTree.Delete(a int, b int):
 * alias func (tree *interval.SearchTree) Delete(a, b int) { tree.EverInserted = slices.Collect( func(yield func(*Item) bool) { for _, i := range tree.EverInserted { if i.Price.Base != a || i.Price.Bound != b { if !yield(i) { return } } } } ) }
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
  all []*Item                   /* for proof purposes only, can be omitted for release build */
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

/* @^DirectedMonotonicGraph(roots []*Item):
       visited := make(map[int]*Item)
       q := fifo.New[*Item](1024)
       for _, root := range roots { q.Enqueue(root) }
       for q.Len() > 0 {
         i := q.Dequeue()
         for o := range i.out {
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
         for o := range i.out {
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
         for i := range o.in {
           if visited[i.index] { continue }
           if (o.index <= i.index) { return false }
           if (o.Price.Base <= i.Price.Base) { return false }
           visited[i.index] = i
           q.Enqueue(i)
         }
       }

       return true
 */

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

       for i := range set.all[ :len(set.all)-2] {
         for j := range set.all[i+1 : len(set.all)-1] {
           for k := range set.all[j+1 : len(set.all)] {
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
         for i := range item.in {
           i.index < item.index && i.Price.Base < item.Price.Base
         }
         for o := range item.out {
           o.index > item.index && o.Price.Base > item.Price.Base
         }
       }
 */

// TODO: possible to allow non-empty PSS, but requires that only this function ever changes a PSS

/* @-PriceSpanSet.GraphWalk()
 * (?)pre-condition: PriceBounds(input)
 * (?)pre-condition: InputSizeBounds(set, input)
 * (?)pre-condition: len(set.tree.EverInserted) == 0
 * (?)post-condition: PurchasableTriosOnPaths(set)
 * (?)post-condition: AllPathTriosPurchasable(set)
 */
func (set *PriceSpanSet) GraphWalk(input []util.Input) {
	/* @-PriceSpanSet.GraphWalk#EachInput: loophead, predicate */
	for _, i := range input {
		/* @-GraphWalk#AddItem: call */
		item := set.NewItem(i.Margin, i.Price)

		/* @-PriceSpanSet.GraphWalk#Step: predicate, call */
		if set.GraphStep(&item) {
			/* @-PriceSpanSet.GraphWalk#InsertItem: call
       * (?)post-condition: slices.Contains(set.tree.EverInserted, item)
       * (?)post-condition: func () bool { for i := item.Price.Base; i < 10000; i++ { x, _ := set.tree.AllIntersections(i, i); if !slices.Contains(x, item) { return false } }; return true }()
       */
			e := set.tree.Insert(item.Price.Base, item.Price.Bound, &item)
			/* @-PriceSpanSet.GraphWalk#InsertItemErrorCheck: predicate */
			if e != nil {
				/* @-PriceSpanSet.GraphWalk#InsertItemError: panic */
				panic(e)
			}
		} else {
      /* @-PriceSpanSet.GraphWalk#elide:
       * (?)post-condition: !slices.Contains(set.tree.EverInserted, item)
       */
       _ = item
    }
	}

	fmt.Println(set)
}

/* @^IdenticalTwins(item *Item):
       if item == nil { return true }
       for ego := item.twin; ego != nil; ego = ego.twin {
         if (item.Price.Base != ego.Price.Base || item.Price.Bound != ego.Price.Bound) { return false }
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

/* @^GraphReachable(a *Item, b *Item):
       q := fifo.New[*Item](1024)
       for o := range a.out { q.Enqueue(o) }
       visited := make(map[int]*Item)
       for q.Len() > 0 {
         i := q.Dequeue()
         if visited[i.index] { continue }
         if i == b { return true }
         visited[i.index] = i
         if i.index < b.index && i.Price.Base < b.Price.Base {
           for o := range i.out { q.Enqueue(o) }
         }
       }
       return false
*/

// induction goal LRGR
/* @^LesserRootsGraphReachable(set *PriceSpanSet, i *Item):
       for root := range set.tree.roots {
         if root.index < i.index && root.Price.Base < i.Price.Base {
           if !GraphReachable(root, i) { return false }
         }
       }
       return true
 */

// induction goal GRGR
/* @^GreaterRootsGraphReachable(set *PriceSpanSet, i *Item):
       for root := range set.tree.roots {
         if root.index >= i.index || root.Price.Base >= i.Price.Base {
           if GraphReachable(root, i) { return false }
         }
       }
       return true
 */

// • price bound can't increase
// • every existing item is either ignored (unmatched), base-matched at item.Price.Base, or truncated at item.Price.Base
// •

/* induction goals:
   • TPUC
   • MPCI
   • MPCG
   • LRGR
   • GRGR
 */

/* @-PriceSpanSet.GraphStep()
 * (?)pre-condition: func () bool { for _, e := range set.tree.EverInserted { for t := e; t != nil; t = t.twin { if t.index >= item.index { return false } } } return true }()
 * (?)pre-condition: DirectedMonotonicGraph(set.roots)
 * (?)post-condition: LesserRootsGraphReachable(set, item)
 * (?)post-condition: GreaterRootsGraphReachable(set, item)
 */
func (set *PriceSpanSet) GraphStep(item *Item) (bool) {
	insert := true
	/* @-PriceSpanSet.GraphStep#TreeQuery: call
   * (?)post-condition: MinimalPriceCompatibleIntersection
   */
	matches, matched := set.tree.AllIntersections(item.Price.Base, item.Price.Base) // GRGR: assert unmatched are irrelevant
	/* @-PriceSpanSet.GraphStep#TreeQueryMatchCheck: predicate */
	if matched {
	  /* @-PriceSpanSet.GraphStep#EachMatch: loophead, predicate */
		for _, match := range matches {
	    /* @-PriceSpanSet.GraphStep#FauxMatch: predicate */
      if match.Price.Bound == item.Price.Base {
        continue  // GRGR: assert not relevant
      }
			/* @-PriceSpanSet.GraphStep#MatchBoundCheck: predicate */
			if item.Price.Base > match.Price.Base {
				e := error(nil)
				/* @-PriceSpanSet.GraphStep#MatchTruncate: call
         * (?)post-condition: func () bool { for i := match.Price.Base; i < match.Price.Bound; i++ { x, _ := set.tree.AllIntersections(i, i); if !slices.Contains(x, match) { return false } }; return true }()
         * (?)post-condition: func () bool { for i := match.Price.Base; i < item.Price.Base; i++ { x, _ := set.tree.AllIntersections(i, i); if !slices.Contains(x, match) { return false } }; return true }()
         */
				insert, e = set.MatchTruncate(match, item.Price.Base) // 2 cases: { true: minimal proxy on truncated (LRGR), false: joined twin chain (LRGR, TPUC }
				/* @-PriceSpanSet.GraphStep#MatchReplaceErrorCheck: predicate */
				if e != nil {
					/* @-PriceSpanSet.GraphStep#MatchReplaceError: panic */
					panic(e)
				}
        /* @-PriceSpanSet.GraphStep#MatchLinkTwins: loophead, predicate */
				for ego := match; ego != nil; ego = ego.twin {
					/* @-PriceSpanSet.GraphStep#MatchLink: call
           * (?)pre-condition: ego.index < item.index
           * (?)pre-condition: ego.Price == match.Price
           * (?)post-condition: GraphReachable(ego, item)
           */
					ego.Link(item)
				}
			} else {
				/* @-PriceSpanSet.GraphStep#CheckSameInterval: predicate
         * (?)pre-condition: item.Price.Base == match.Price.Base
         */
        if item.Price.Bound == match.Price.Bound {
					/* @-PriceSpanSet.GraphStep#SpliceTwinsRight:
           * (?)pre-condition: func () bool { for m := match; m != nil; m = m.twin { if (m.index >= item.index) { return false } } return true }()
           */
          item.twin = match.twin
					/* @-PriceSpanSet.GraphStep#SpliceTwinsLeft:
           * (?)post-condition: func () bool { for m := match; m != nil; m = m.twin { if (m.index >= set.ItemIndex.Load()) { return false } } return true }()
           * (?)post-condition: UniqueTwins(match)
           */
          match.twin = item
					/* @-PriceSpanSet.GraphStep#ElideSameInterval */
					insert = false // join a direct match (LRGR, TPUC)
				} // else implicit `insert = true`: equivalent proxy on match.in, or root if empty (LRBR)
        /* @-PriceSpanSet.GraphForkSamePriceCheck: predicate, call */
        if len(match.in) == 0 {
          /* @-PriceSpanSet.GraphForkSamePriceRoot: call
           * (?)pre-condition: slices.Contains(set.roots, match)
           * (?)pre-condition[MPCI]: len(set.roots) == 0 || set.roots[len(set.roots)-1].Price.Base <= item.Price.Base
           * (?)post-condition[MPCI]: PriceCompatible(item, item.Price.Base + 1, item.Price.Bound)
           * (?)post-condition[MPCG]: func() bool { for i := range set.tree.roots { if (PriceCompatible(i, item) { return false } } }()
           */
          set.roots = append(set.roots, item)
        } else {
          /* @-PriceSpanSet.GraphForkSamePrice: loophead, predicate, call
           * (?)post-condition: func () bool { for ego := match; ego != nil; ego = ego.twin { for _, c := range ego.in { if !GraphReachable(c, item) { return false } } }; return true }
           */
          for _, predecessor := range match.in {  // 2 cases: equivalent proxy as direct twin or (price incompatible) peer (LRBR)
            /* @-PriceSpanSet.GraphForkSamePriceSuperseded: predicate */
            if predecessor.out[len(predecessor.out)-1] == match {
              /* @-PriceSpanSet.GraphBranchSamePrice: call
               * (?)pre-condition: LesserRootsGraphReachable(set, predecessor)
               * (?)post-condition: LesserRootsGraphReachable(set, item)
               */
              predecessor.Link(item)
            } /* else a path exists over an intermediate item */
          }
        }
			}
		}
	} else {
	  /* @-PriceSpanSet.GraphStep#Root
     * (?)pre-condition[MPCI]: len(set.roots) == 0 || set.roots[len(set.roots)-1].Price.Base < item.Price.Base
     * (?)post-condition[MPCI]: PriceCompatible(item, item.Price.Base + 1, item.Price.Bound)
     * (?)post-condition[MPCG]: func() bool { for i := range set.tree.roots { if (PriceCompatible(i, item) { return false } } }()
     */
    set.roots = append(set.roots, item)
  }

	/* @-PriceSpanSet.GraphStep#return */
	return insert;
}

/* @^JoinHomogeneousLarger(existing *Item, joining *Item):
       ispan := (existing.Price.Bound - existing.Price.Base)
       tspan := (joining.Price.Bound - joining.Price.Base)

       if (ispan >= tspan) { return false } // append truncated matched chain
       for i := existing.twin; i != nil; i = i.twin {
         if (ispan != (i.Price.Bound - i.Price.Base)) { return false }
       }
       for i := joining.twin; i != nil; i = i.twin {
         if (tspan != (i.Price.Bound - i.Price.Base)) { return false }
       }
       return true
 */

/* @^MutuallyExclusiveJoin(existing *Item, joining *Item):
       if existing.index >= joining.index { return false }
       if existing.Price.Bound == joining.Price.Bound { return false }

       for left := existing; left != nil; left = left.twin {
         for right := joining; right != nil; right = right.twin {
           if left.index >= right.index { return false }
         }
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

// induction goal: TwinPricesUniquelyChained (TPUC)
// need an induction goal that only the first arriving instance of Price P occurs in set.tree.EverInserted
// - subsequent instances matching P or truncated to P will be appended to P.twin
// - twin chain is acyclic
// - no twin on the chain occurs on any other twin chain
// - no twin appears in set.tree.EverInserted
// - every twin has a unique root present in set.tree.EverInserted
//   - every twin root truncation will be applied to each twin in its chain
/* @^TwinPricesUniquelyChained(set *PriceSpanSet):
       for i, left := range set.tree.EverInserted {
         for _, right := range set.tree.EverInserted[i+1:] {
           if left.Price == right.Price { return false }
         }
         if left.twin != nil {
           for tleft := left; tleft.twin != nil; tleft = tleft.twin {
             if left.Price != tleft.Price { return false }
             for tright := tleft.twin; tright != nil; tright = tright.twin {
               if tleft == tright { return false }
             }
           }
         }
 */

/* MPCI: AllIntersections returns an optimal set S of price compatible items
 * - items outside S that are price compatible with some element of S will be price compatible with i
 * - items outside S not price compatible with any element of S will not be price compatible with i
 * - items in S are price compatible with i
 * - no item in S is price compatible with any other item in S
 * where inclusion in S covers S.twin...
 */
/* @^MinimalPriceCompatibleIntersection(set *PriceSpanSet, matches []*Item, item *Item):
       for i := set.tree.EverInserted {
         if slices.Contains(matches, i) {
           if !PriceCompatible(i, item) { return false }
           for o := range i.out {
             if PriceCompatible(o, item) { return false }
           }
         } else if (!PriceCompatible(i, item) {
           if slices.Contains(matches, i) { return false }
         } else {
           for m := range matches {
             if GraphReachable(i, m) { return true }
           }
           return false
         }
       }
 */


// possible induction goal: MinimalPriceCompatibleGraph (MPCG)
// - no forward edges
// - all paths price compatible
// - all price compatible on paths

/* @-PriceSpanSet.MatchTruncate()
 * (?)pre-condition: IdenticalTwins(match)
 */
func (set *PriceSpanSet) MatchTruncate(match *Item, at int) (bool, error) {
	/* @-PriceSpanSet.MatchTruncate#Delete call */
	set.tree.Delete(match.Price.Base, match.Price.Bound)
  match.Price.Bound = at // LRGR: assert input supersedes match at new bound
  var ego *Item
  /* @-PriceSpanSet.MatchTruncate#TruncateTwins: loophead, predicate
   * (?)pre-condition: JoinHomogeneousLarger(existing, match)
   * (?)pre-condition: MutuallyExclusiveJoin(existing, match)
   * (?)post-condition: IdenticalTwins(twin)
   */
  for ego = match; ego.twin != nil; ego = ego.twin {
    /* @-PriceSpanSet.MatchTruncate#Truncate
     * (?)pre-condition: ego.Price.Bound > at
     */
    ego.twin.Price.Bound = at // LRGR: assert input supersedes match at new bound
  }
  /* @-PriceSpanSet.MatchTruncate#FindSameInterval: call */
	existing, twinned := set.tree.Find(match.Price.Base, at)
	/* @-PriceSpanSet.MatchTruncate#CheckSameInterval: predicate */
	if twinned {
    /* @-PriceSpanSet.MatchTruncate#SpliceTwinsRight
     * (?)pre-condition: UniqueTwins(existing)
     * (?)pre-condition: UniqueTwins(match)
     * (?)pre-condition: func () (bool) { for ego := match.twin; ego != nil; ego = ego.twin { if slices.Contains(set.tree.EverInserted, ego) { return false } } return true }()
     */
    ego.twin = existing.twin // TPUC: none of existing.twin... { were the first at existing.Price, occur on match.twin... }
    /* @-PriceSpanSet.MatchTruncate#SpliceTwinsLeft:
     * (?)post-condition: func () (bool) { for ego := match.twin; ego != nil; ego = ego.twin { if slices.Contains(set.tree.EverInserted, ego) { return false } } return true }()
     * (?)post-condition: UniqueTwins(existing)
     */
    existing.twin = match // TPUC: none of match.twin... { were the first at existing.Price, occur on existing.twin... }
  }
	/* @-PriceSpanSet.MatchTruncate#CheckSameInterval: predicate */
	if twinned {
	  /* @-PriceSpanSet.MatchTruncate#ElideSameInterval */
    return false, nil  // MPCI: AllIntersections on existing.Price is sufficient for match and .twin...
  } else {
		/* @-PriceSpanSet.MatchTruncate#Insert: call */
		return true, set.tree.Insert(match.Price.Base, match.Price.Bound, match)  // MPCI: match should be found at .Price in AllIntersections
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
