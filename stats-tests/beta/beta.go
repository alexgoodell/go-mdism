package main

import (
	"fmt"
	"github.com/leesper/go_rng" //imported as rng
	"time"
)

func main() {
	betaGen := rng.NewBetaGenerator(time.Now().UnixNano())
	for i := 0; i < 10000; i++ {
		fmt.Println(betaGen.Beta(1, 1)) // alpha, beta
	}

}
