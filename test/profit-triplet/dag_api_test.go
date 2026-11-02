package dag_test

import (
	"testing"

	"experimental/profit-triplet/internal/profit-triplet"
)

func TestFoo(test *testing.T) {
	item := dag.NewItem(1, 1)
	if item.Margin != 1 {
		test.Errorf("Item margin should be 1, but found %v", item.Margin)
	}
	if item.Price.Base != 1 {
		test.Errorf("Item margin should be 1, but found %v", item.Price)
	}
}

