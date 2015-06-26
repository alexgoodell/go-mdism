package main

import (
	"fmt"
	"github.com/leesper/go_rng" //imported as rng
	"time"
)

func main() {
	gammaGen := rng.NewGammaGenerator(time.Now().UnixNano())
	for i := 0; i < 10000; i++ {
		fmt.Println(gammaGen.Gamma(16, 0.1)) // alpha, beta
	}

}