package dag

import (
  "fmt"
 "strings"
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
 * alias func (tree *interval.SearchTree) AllIntersections(a int, b int) { slice := tree.EverInserted[%i, %i + %count]; return slice, %matched }
 * %values[0]:
 *   %i int: %i >= 0 && %i < (len(tree.EverInserted) - %count)
 *   %count int: count > 0 && count <= len(tree.EverInserted)
 *   %matched bool: matched == true
 * (?)pre-condition: len(tree.EverInserted) > 0
 * %values[1]:
 *   %i int: (len(tree.EverInserted) == 0 && %i == 0) || (%i >= 0 && %i < len(tree.EverInserted))
 *   %count int: count == 0
 *   %matched bool: matched == false
 * (?)post-condition: func() bool { for _, i := range slice { if i.Price.Base > a { return false } }; return true }()
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
}

type PriceSpanSet struct {
  ItemIndex atomic.Int32
  tree *interval.SearchTree[*Item,int]
}

func compare (a, b int) int { return a - b }

/* @-NewPriceSpanSet(): alias func NewPriceSpanSet() (PriceSpanSet) { return PriceSpanSet{ tree { make([]*Item, 0) } } } */
func NewPriceSpanSet() (PriceSpanSet) {
	return PriceSpanSet { tree: interval.NewSearchTree[*Item,int](compare) }
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
	}
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
	lines := make([]string, set.tree.Size() + 1)
	lines[0] = fmt.Sprintf("• PriceSpanSet:")
	set.tree.InOrderTraverse(func (base int, bound int, item *Item) error {
		lines[item.index] = fmt.Sprintf("• %v", item)
  	return nil
  })
	return strings.Join(lines, "\n")
}

/* @-PriceSpanSet.GraphWalk()
 * (?)pre-condition: func() bool { for _, i := range input { if i.Price <= 0 || i.Price >= 10000 { return false } }; return true }()
 */
func (set *PriceSpanSet) GraphWalk(input []util.Input) {
	/* @-PriceSpanSet.GraphWalk#EachInput: loophead, predicate */
	for _, i := range input {
		/* @-GraphWalk#AddItem: call */
		item := set.NewItem(i.Margin, i.Price)

		/* @-PriceSpanSet.GraphWalk#Step: call */
		set.GraphStep(&item)

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

	fmt.Println(set)
}

/* @-PriceSpanSet.GraphStep()
 * (?)pre-condition: !slices.Contains(set.tree.EverInserted, item)
 */
func (set *PriceSpanSet) GraphStep(item *Item) {
	/* @-PriceSpanSet.GraphStep#TreeQuery: call */
	matches, matched := set.tree.AllIntersections(item.Price.Base, item.Price.Base)
	/* @-PriceSpanSet.GraphStep#TreeQueryMatchCheck: predicate */
	if matched {
	  /* @-PriceSpanSet.GraphStep#EachMatch: loophead, predicate */
		for _, match := range matches {
			/* @-PriceSpanSet.GraphStep#MatchBoundCheck: predicate */
			if item.Price.Base < match.Price.Bound {
				/* @-PriceSpanSet.GraphStep#MatchTruncate: call */
				e := set.ItemTruncate(match, item.Price.Base)
				/* @-PriceSpanSet.GraphStep#MatchReplaceErrorCheck: predicate */
				if e != nil {
					/* @-PriceSpanSet.GraphStep#MatchReplaceError: panic */
					panic(e)
				}
				/* @-PriceSpanSet.GraphStep#MatchLink: call */
				match.Link(item)
			} else {   // better here to just merge the better margin into the match and skip this input?
				/* @-PriceSpanSet.GraphForkSamePrice: loophead, predicate, call */
				for _, predecessor := range match.AllIn {
					/* @-PriceSpanSet.GraphBranchSamePrice: call */
					predecessor.Link(item)
				}
			}
		}
	}
}

/* @-PriceSpanSet.ItemTruncate() */
func (set *PriceSpanSet) ItemTruncate(item *Item, at int) error {
	/* @-PriceSpanSet.ItemTruncate#Delete */
	set.tree.Delete(item.Price.Base, item.Price.Bound)
	/* @-PriceSpanSet.ItemTruncate#Truncate */
	item.Price.Bound = at
	/* @-PriceSpanSet.ItemTruncate#Insert: alias set.tree.EverInserted = append(set.tree.EverInserted, item)
	 * (?)pre-condition: item.Price.Bound > item.Price.Base
	 */
  return set.tree.Insert(item.Price.Base, item.Price.Bound, item)  // bug: collapsed bounds
}
