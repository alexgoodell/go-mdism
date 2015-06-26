package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 10000; i++ {
		fmt.Println(rand.NormFloat64())
	}

}
