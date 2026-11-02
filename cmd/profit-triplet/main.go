package main

import (
	"os"
	"fmt"
	"strconv"

	"experimental/profit-triplet/pkg/util"
	"experimental/profit-triplet/internal/profit-triplet"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("• Missing argument: <item-count>")
		os.Exit(1)
	}
	itemcount, fail := strconv.Atoi(os.Args[1])
	if fail != nil {
		fmt.Println("• Error: failed to parse", os.Args[1], "as an integer:", fail)
		os.Exit(1)
	}

	fmt.Println(itemcount)

	input := util.Generate(itemcount)
	// for i := range util.Min(len(input), 3) {
	// 	fmt.Println(i, ":", input[i])
	// }

 	set := dag.NewPriceSpanSet()
 	set.GraphWalk(input)

	// treetest()
}
