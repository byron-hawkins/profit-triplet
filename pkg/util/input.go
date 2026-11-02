package util

import (
 "time"
	"math/rand"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Input struct {
	Margin int
	Price  int
}

func Generate(count int) ([]Input) {
	input := make([]Input, count)
	for i := range len(input) {
		input[i] = Input { rand.Intn(1000), rand.Intn(1000) }
	}
	return input
}
