package dag

import (
  "fmt"
 "strings"
 "slices"
 "math"
 "sync/atomic"
 "github.com/rishabhofficial/golang-standard-template-library/containers/vector"
 "github.com/rdleal/intervalst/interval"
 "gopkg.in/fifo.v0"

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
 * alias func (tree *interval.SearchTree) Find(a, b int) (*Item, bool) { for _, i := range * tree.EverInserted { if i.Price.Base == a && i.Price.Bound == b { return i, true } } return nil, false }
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

	in  *vector.Vector[*Item]
	out *vector.Vector[*Item]
	twin *Item
}

/* Note: unsafe for concurrent execution */
type PriceSpanSet struct {
  ItemIndex atomic.Int32
  roots *vector.Vector[*Item]   /* for proof purposes only, can be omitted for release build */
  tree *interval.SearchTree[*Item,int]
}

func compare (a, b int) int { return a - b }
func compare_items (a, b *Item) int { if a.Price.Base == b.Price.Base { return a.Price.Bound - b.Price.Bound } else { return a.Price.Base - b.Price.Base; } }

/* @-NewPriceSpanSet(): alias func NewPriceSpanSet() (PriceSpanSet) { return PriceSpanSet{ tree { make([]*Item, 0) } } } */
func NewPriceSpanSet() (PriceSpanSet) {
  return PriceSpanSet { roots: vector.NewVector[*Item](), tree: interval.NewSearchTree[*Item,int](compare) }
}

/* @-Item#Constructor */
func (set *PriceSpanSet) NewItem(margin int, price int) (Item) {
  /* @-Item#ConstructorWatermark: call */
  prior := set.ItemIndex.Load()

  /* @-Item#Construct:
   * (?)post-condition: i.index > prior && i.Margin == margin && i.Price == PriceSpan{ price, math.MaxInt32 } && i.best == [3]int{ margin, 0, 0 } && i.in.Empty() && i.out.Empty() && i.twin == nil
   */
  i := Item {
    /* @-Item#ConstructIndex: call */
		set.ItemIndex.Add(1),
		margin,
		PriceSpan { price, math.MaxInt32 },
		[3]int{ margin, 0, 0 },
		vector.NewVector[*Item](),
		vector.NewVector[*Item](),
		nil,
	}
  return i
}

/* @^JoinHomogeneousLarger(existing *Item, joining *Item):
       ispan := (existing.Price.Bound - existing.Price.Base)
       tspan := (joining.Price.Bound - joining.Price.Base)

       if joining.twin == nil && ispan == tspan { return true } // append unmatched fresh singleton

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
       if existing.Price.Bound == joining.Price.Bound {
         if (joining.twin != nil) { return false } // allow only an unmatched fresh singleton
         for x := existing.twin; x != nil; x = x.twin {
           if (x.index >= joining.index) { return false }
         }
       } else {
         for left := existing; left != nil; left = left.twin {
           for right := joining; right != nil; right = right.twin {
             if left.index >= right.index { return false }
           }
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

/* @-Item#Join()
 * (?)pre-condition: JoinHomogeneousLarger(item, twin)
 * (?)pre-condition: MutuallyExclusiveJoin(item, twin)
 * (?)post-condition: UniqueTwins(item)
 */
func (item *Item) Join(twin *Item) {
	/* @-Item#Join#CheckTwin: predicate */
	if item.twin != nil {
    /* @-Item#Join#SpliceTwin */
		twin.twin = item.twin
	}
  /* @-Item#Join#JoinTwin */
	item.twin = twin
}

/* @-AllIn() */
func (item *Item) AllIn(yield func(index int, item *Item) bool) {
	/* @-AllIn#range: loophead, predicate, call */
	for i := range item.in.Size() {
	  /* @-AllIn#yield: predicate, call */
		if !yield(i, item.in.At(i)) {
			return
		}
	}
}

func (item *Item) AllOut(yield func(index int, item *Item) bool) {
	for i := range item.out.Size() {
		if !yield(i, item.out.At(i)) {
			return
		}
	}
}

func (item *Item) String() string {
	edges := func (v *vector.Vector[*Item]) string {
		if v.Empty() {
			return ""
		}
		list := make([]string, v.Size())
		for i := range v.Size() {
			list[i] = fmt.Sprintf("%v", v.At(i).index)
		}
		return strings.Join(list, ",")
  }
	return fmt.Sprintf("%v: %v @ %v in[%v] x out [%v]", item.index, item.Price, item.Margin, edges(item.in), edges(item.out))
}

/* @^DirectedMonotonicGraph(roots []*Item):
       visited := make(map[int]*Item)
       q := fifo.New[*Item](1024)
       for _, root := range roots {
         q.Enqueue(root)
       }
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

/* @-Item.Link():
 * (?)pre-condition: to.out.Empty() && from != to && from.index < to.index && from.Price.Base < to.Price.Base
 * (?)pre-condition: DirectedMonotonicSubGraphs(from)
 * (?)post-condition: to.out.Empty() && slices.Contains(from.out.elements, to) && slices.Contains(to.in.elements, from)
 * (?)post-condition: DirectedMonotonicSubGraphs(from)
 */
func (from *Item) Link(to *Item) {
	/* @-Item.Link#out: call */
	from.out.PushBack(to)
	/* @-Item.Link#in: call */
	to.in.PushBack(from)
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

// TODO: possible to allow non-empty PSS, but requires that only this function ever changes a PSS

/* @-PriceSpanSet.GraphWalk()
 * (?)pre-condition: PriceBounds(input)
 * (?)pre-condition: InputSizeBounds(set, input)
 * (?)pre-condition: len(set.tree.EverInserted) == 0
 */
func (set *PriceSpanSet) GraphWalk(input []util.Input) {
	/* @-PriceSpanSet.GraphWalk#EachInput: loophead, predicate */
	for _, i := range input {
		/* @-GraphWalk#AddItem: call */
		item := set.NewItem(i.Margin, i.Price)

		/* @-PriceSpanSet.GraphWalk#Step: predicate, call */
		if set.GraphStep(&item) {
			/* @-PriceSpanSet.GraphWalk#InsertItem: call */
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

/* @^IdenticalTwins(item *Item):
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

/* @-PriceSpanSet.GraphStep()
 * (?)pre-condition: func () bool { for _, e := range set.tree.EverInserted { for t := e; t != nil; t = t.twin { if t.index >= item.index { return false } } } return true }()
 * (?)pre-condition: DirectedMonotonicGraph(set.roots)
 * (?)post-condition: DirectedMonotonicGraph(set.roots)
 */
func (set *PriceSpanSet) GraphStep(item *Item) (bool) {
	insert := true
	/* @-PriceSpanSet.GraphStep#TreeQuery: call
   * (?)post-condition: OrderedIntersection(matches)
   */
	matches, matched := set.tree.AllIntersections(item.Price.Base, item.Price.Base)
	/* @-PriceSpanSet.GraphStep#TreeQueryMatchCheck: predicate */
	if matched {
	  /* @-PriceSpanSet.GraphStep#EachMatch: loophead, predicate */
		for _, match := range matches {
	    /* @-PriceSpanSet.GraphStep#FauxMatch: predicate */
      if match.Price.Bound == item.Price.Base {
        continue
      }
			/* @-PriceSpanSet.GraphStep#MatchBoundCheck: predicate */
			if item.Price.Base > match.Price.Base {
				e := error(nil)
				/* @-PriceSpanSet.GraphStep#MatchTruncate: call */
				insert, e = set.MatchTruncate(match, item.Price.Base)
				/* @-PriceSpanSet.GraphStep#MatchReplaceErrorCheck: predicate */
				if e != nil {
					/* @-PriceSpanSet.GraphStep#MatchReplaceError: panic */
					panic(e)
				}
        /* @-PriceSpanSet.GraphStep#MatchLinkTwins: loophead, predicate */
				for ego := match; ego != nil; ego = ego.twin {
					/* @-PriceSpanSet.GraphStep#MatchLink: call
           * (?)pre-condition: ego.index < item.index
           */
					ego.Link(item)
				}
			} else {
				/* @-PriceSpanSet.GraphStep#CheckSameInterval: predicate
         * (?)pre-condition: item.Price.Base == match.Price.Base
         */
        if item.Price.Bound == match.Price.Bound {
					/* @-PriceSpanSet.GraphStep#JoinSameInterval: call
           * (?)pre-condition: func () bool { for m := match; m != nil; m = m.twin { if (m.index >= item.index) { return false } } }()
           * (?)post-condition: !slices.Contains(set.tree.EverInserted, item)
           */
					match.Join(item)
					/* @-PriceSpanSet.GraphStep#ElideSameInterval */
					insert = false
				}
        /* @-PriceSpanSet.GraphForkSamePriceCheck: predicate, call */
        if match.in.Empty() {
          /* @-PriceSpanSet.GraphForkSamePriceRoot: call
           * (?)pre-condition: slices.Contains(set.roots.elements, match)
           */
          set.roots.PushBack(item)
        } else {
          /* @-PriceSpanSet.GraphForkSamePrice: loophead, predicate, call */
          for _, predecessor := range match.AllIn {
            /* @-PriceSpanSet.GraphForkSamePriceSuperseded: predicate
             * alias predecessor.out.elements[len(predecessor.out.elements)-1] == match
             */
            if predecessor.out.Back() == match {
              /* @-PriceSpanSet.GraphBranchSamePrice: call */
              predecessor.Link(item)
            } /* else a path exists over an intermediate item */
          }
        }
			}
		}
	} else {
	  /* @-PriceSpanSet.GraphStep#Root */
    set.roots.PushBack(item)
  }
	/* @-PriceSpanSet.GraphStep#return */
	return insert;
}

/* @-PriceSpanSet.MatchTruncate()
 * (?)pre-condition: IdenticalTwins(match)
 */
func (set *PriceSpanSet) MatchTruncate(match *Item, at int) (bool, error) {
	/* @-PriceSpanSet.MatchTruncate#Delete call */
	set.tree.Delete(match.Price.Base, match.Price.Bound)
  /* @-PriceSpanSet.MatchTruncate#FindSameInterval: call */
	twin, twinned := set.tree.Find(match.Price.Base, at)
	/* @-PriceSpanSet.MatchTruncate#CheckSameInterval: predicate */
	if twinned {
    /* @-PriceSpanSet.MatchTruncate#JoinSameInterval: call
     * (?)post-condition: func (tree *interval.SearchTree[*Item, int]) (bool) { for ego := match.twin; ego != nil; ego = ego.twin { if slices.Contains(tree.EverInserted, ego) { return false } } return true }(set.tree)
     */
	  twin.Join(match)
  }
  /* @-PriceSpanSet.MatchTruncate#TruncateTwins: loophead, predicate
   * (?)post-condition: IdenticalTwins(twin)
   */
  for ego := match; ego != nil; ego = ego.twin {
    /* @-PriceSpanSet.MatchTruncate#Truncate */
    ego.Price.Bound = at
  }
	/* @-PriceSpanSet.MatchTruncate#CheckSameInterval: predicate */
	if twinned {
	  /* @-PriceSpanSet.MatchTruncate#ElideSameInterval */
    return false, nil
  } else {
		/* @-PriceSpanSet.MatchTruncate#Insert: call */
		return true, set.tree.Insert(match.Price.Base, match.Price.Bound, match)
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
