package dag_test

import (
	"testing"

	"demo/profit-triplet/internal/profit-triplet"
)

func TestFoo(test *testing.T) {
 	set := dag.NewPriceSpanSet()
	item := set.NewItem(1, 1)
	if item.Margin != 1 {
		test.Errorf("Item margin should be 1, but found %v", item.Margin)
	}
	if item.Price.Base != 1 {
		test.Errorf("Item margin should be 1, but found %v", item.Price)
	}
}

