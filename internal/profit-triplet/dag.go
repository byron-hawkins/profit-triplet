package dag

import (
  "fmt"
 "strings"
 "slices"
 "math"
 "sync/atomic"
 "github.com/rishabhofficial/golang-standard-template-library/containers/vector"
 "github.com/rdleal/intervalst/interval"

	"demo/profit-triplet/pkg/util"
)

/* @+atomic.Int32: data-model n int32 */
/* @+atomic.Int32.Add(n int): alias func (i *atomic.Int32) Add(n int) { i.n += n } */

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

/* @-NewPriceSpanSet(): alias func NewPriceSpanSet() (PriceSpanSet) { return PriceSpanSet{ tree { make([]*Item, 0) } } } */
func NewPriceSpanSet() (PriceSpanSet) {
  return PriceSpanSet { roots: vector.NewVector[*Item](), tree: interval.NewSearchTree[*Item,int](compare) }
}

/* @-Item#constructor:
 * (?)post-condition: return Item{ Price: PriceSpan{ Base: price, Bound: math.MaxInt32 }, out: vector.NewVector[*Item]() }
 */
func (set *PriceSpanSet) NewItem(margin int, price int) (Item) {
	return Item {
		set.ItemIndex.Add(1),
		margin,
		PriceSpan { price, math.MaxInt32 },
		[3]int{ margin, 0, 0 },
		vector.NewVector[*Item](),
		vector.NewVector[*Item](),
		nil,
	}
}

/* @-Item#Join()
 * (?)pre-condition: func () bool { if twin.twin != nil { return false } for ego := item.twin; ego != nil; ego = ego.twin { if twin == ego { return false } } return true }()
 * (?)post-condition: func () bool { for ego := item.twin; ego != nil; ego = ego.twin { if twin == ego { return false } } return true }()
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

/* @-Item.Link():
 * (?)pre-condition: to.out.Empty() && from != to
 * (?)post-condition: to.out.Empty() && slices.Contains(from.out.elements, to) && slices.Contains(to.in.elements, from)
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

/* @^PriceBounds:
   func() bool {
     for _, i := range input {
       if i.Price <= 0 || i.Price >= 10000 {
         return false
       }
     }; return true
   }()
*/

/* @^InputSizeBounds:
   func() bool {
     (len(input) < math.MaxInt32) &&
     (set.ItemIndex.Load() < (math.MaxInt32 - int32(len(input)))) &&
     (set.ItemIndex.Load() >= int32(0))
   }()
*/

/* @-PriceSpanSet.GraphWalk()
 * (?)pre-condition: PriceBounds
 * (?)pre-condition: InputSizeBounds
 */
func (set *PriceSpanSet) GraphWalk(input []util.Input) {
	/* @-PriceSpanSet.GraphWalk#EachInput: loophead, predicate */
	for _, i := range input {
		/* @-GraphWalk#AddItem: call */
		item := set.NewItem(i.Margin, i.Price)

		/* @-PriceSpanSet.GraphWalk#Step: predicate, call */
		if set.GraphStep(&item) {
			/* @-PriceSpanSet.GraphWalk#InsertItem: alias set.tree.EverInserted = append(set.tree.EverInserted, &item)
			 * (?)pre-condition: item.Price.Base < item.Price.Bound
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

/* @-PriceSpanSet.GraphStep()
 * (?)pre-condition: !slices.Contains(set.tree.EverInserted, item)
 */
func (set *PriceSpanSet) GraphStep(item *Item) (bool) {
	insert := true
	/* @-PriceSpanSet.GraphStep#TreeQuery: call */
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
					/* @-PriceSpanSet.GraphStep#MatchLink: call */
					ego.Link(item)
				}
			} else {
				/* @-PriceSpanSet.GraphStep#CheckSameInterval: predicate */
        if item.Price.Bound == match.Price.Bound {
					/* @-PriceSpanSet.GraphStep#JoinSameInterval: call */
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

/* @-PriceSpanSet.MatchTruncate() */
func (set *PriceSpanSet) MatchTruncate(match *Item, at int) (bool, error) {
	/* @-PriceSpanSet.MatchTruncate#Delete */
	set.tree.Delete(match.Price.Base, match.Price.Bound)
  /* @-PriceSpanSet.MatchTruncate#FindSameInterval: call */
	twin, twinned := set.tree.Find(match.Price.Base, at)
  /* @-PriceSpanSet.MatchTruncate#TruncateTwins: loophead, predicate */
  for ego := match; ego != nil; ego = ego.twin {
    /* @-PriceSpanSet.MatchTruncate#Truncate */
    ego.Price.Bound = at
  }
	/* @-PriceSpanSet.MatchTruncate#CheckSameInterval: predicate */
	if twinned {
    /* @-PriceSpanSet.MatchTruncate#JoinSameInterval: call */
	  twin.Join(match)
	  /* @-PriceSpanSet.MatchTruncate#ElideSameInterval */
    return false, nil
  } else {
		/* @-PriceSpanSet.MatchTruncate#Insert: alias set.tree.EverInserted = append(set.tree.EverInserted, match)
		 * (?)pre-condition: match.Price.Bound > match.Price.Base
		 */
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
