package dag

import (
	"fmt"
	"math/rand"
	"testing"

	"demo/profit-triplet/pkg/util"
)

func TestGraph(test *testing.T) {
	input := util.Generate(10)
	mutatedInputs, err := mutateInputs(input)
	if err != nil {
		test.Fatalf("Failed to generate mutated inputs: %v", err)
	}

	inputSets := append([][]util.Input{input}, mutatedInputs...)
	for i, inputSet := range inputSets {
		test.Run(fmt.Sprintf("input-%d", i), func(test *testing.T) {
			testGraphIsAcyclic(test, inputSet)
		})
	}
}

func mutateInputs(base []util.Input) ([][]util.Input, error) {
	const mutationCount = 10
	const priceDomain = 1000
	const priceAttempts = 100

	mutated := make([][]util.Input, mutationCount)
	for i := range mutated {
		input := append([]util.Input(nil), base...)

		if rand.Intn(2) == 0 && mutatePrice(input, priceDomain, priceAttempts) {
			mutated[i] = input
			continue
		}
		if swapInputs(input) {
			mutated[i] = input
			continue
		}
		if mutatePrice(input, priceDomain, priceAttempts) {
			mutated[i] = input
			continue
		}
		return nil, fmt.Errorf("input set %d cannot be mutated", i)
	}
	return mutated, nil
}

func mutatePrice(input []util.Input, priceDomain, attempts int) bool {
	for attempt := 0; attempt < attempts; attempt++ {
		index := rand.Intn(len(input))
		newPrice := rand.Intn(priceDomain)
		oldPrice := input[index].Price
		if newPrice == oldPrice {
			continue
		}

		relevant := false
		for otherIndex, other := range input {
			if otherIndex != index && relativePrice(oldPrice, other.Price) != relativePrice(newPrice, other.Price) {
				relevant = true
				break
			}
		}
		if relevant {
			input[index].Price = newPrice
			return true
		}
	}
	return false
}

func relativePrice(first, second int) int {
	difference := first - second
	if difference > 1 {
		return 1
	}
	if difference < -1 {
		return -1
	}
	return 0
}

func swapInputs(input []util.Input) bool {
	if len(input) < 2 {
		return false
	}

	for attempt := 0; attempt < len(input)*len(input); attempt++ {
		first := rand.Intn(len(input))
		second := rand.Intn(len(input))
		if first != second && input[first] != input[second] {
			input[first], input[second] = input[second], input[first]
			return true
		}
	}
	return false
}

func testGraphIsAcyclic(test *testing.T, input []util.Input) {
	test.Helper()

	set := NewPriceSpanSet()
	set.GraphWalk(input)
	defer func() {
		set.tree = nil
	}()

	nodes := make([]*Item, 0, set.tree.Size())
	set.tree.InOrderTraverse(func(_, _ int, item *Item) error {
    for ego := item; ego != nil; ego = ego.twin {
		  nodes = append(nodes, ego)
    }
		return nil
	})
	if len(nodes) != len(input) {
		test.Fatalf("Graph contains %d nodes; expected %d", len(nodes), len(input))
	}

	graphNodes := make(map[*Item]struct{}, len(nodes))
	for _, node := range nodes {
		graphNodes[node] = struct{}{}
	}

	visited := make(map[*Item]bool, len(nodes))
	inPath := make(map[*Item]bool, len(nodes))
	var visit func(*Item) bool
	visit = func(node *Item) bool {
		visited[node] = true
		inPath[node] = true
		for i := range node.out.Size() {
			successor := node.out.At(i)
			if _, exists := graphNodes[successor]; !exists {
				test.Fatalf("Node %d references a successor outside the graph", node.index)
			}
			if inPath[successor] {
				return false
			}
			if !visited[successor] && !visit(successor) {
				return false
			}
		}
		delete(inPath, node)
		return true
	}

	for _, node := range nodes {
		if !visited[node] && !visit(node) {
			test.Fatalf("Graph contains a cycle involving node %d", node.index)
		}
	}
}
