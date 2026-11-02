package util_test

import (
	"testing"

	"demo/profit-triplet/pkg/util"
)

func TestGenerate(test *testing.T) {
	expected := 4
	input := util.Generate(expected)

	count := 0
	for _, i := range input {
		if i.Margin < 0 || i.Margin > 1000 {
			test.Errorf("Margin out of range: %v", i.Margin)
		}
		if i.Price < 0 || i.Price > 1000 {
			test.Errorf("Price out of range: %v", i.Price)
		}
		count++
	}
	if count != expected {
		test.Errorf("Expected to generate 4 inputs, but got %v", count)
	}
}

func TestGenerateUniformDistribution(test *testing.T) {
	const sampleCount = 10000
	const binCount = 10
	const expectedPerBin = sampleCount / binCount
	const allowedDeviation = sampleCount / 100

	for iteration := 0; iteration < 10; iteration++ {
		input := util.Generate(sampleCount)
		marginCounts := make([]int, binCount)
		priceCounts := make([]int, binCount)

		for _, value := range input {
			marginCounts[value.Margin/100]++
			priceCounts[value.Price/100]++
		}

		iterationPassed := true
		for bin := 0; bin < binCount; bin++ {
			marginDifference := marginCounts[bin] - expectedPerBin
			priceDifference := priceCounts[bin] - expectedPerBin
			if marginDifference < -allowedDeviation || marginDifference > allowedDeviation ||
				priceDifference < -allowedDeviation || priceDifference > allowedDeviation {
				iterationPassed = false
				break
			}
		}
		if iterationPassed {
			return
		}
	}

	test.Error("Uniform distribution check failed in all 10 iterations")
}
