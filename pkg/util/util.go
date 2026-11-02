package util

import (
	"golang.org/x/exp/constraints"
)

func  Min[I constraints.Integer](a I, b I) (I) {
	if (a < b) {
		return a
	} else {
		return b
	}
}

func  Max[I constraints.Integer](a I, b I) (I) {
	if (a < b) {
		return b
	} else {
		return a
	}
}
