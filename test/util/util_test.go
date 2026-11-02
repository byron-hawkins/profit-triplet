package util_test

import (
	"testing"

	"experimental/profit-triplet/pkg/util"
)

func TestMin(test *testing.T) {
	lesser := util.Min(22, 21)
	if lesser != 21 {
		test.Errorf("Min(22, 21) = %v??", lesser)
	}
}

func TestMax(test *testing.T) {
	greater := util.Max(22, 21)
	if greater != 22 {
		test.Errorf("Max(22, 21) = %v??", greater)
	}
}

