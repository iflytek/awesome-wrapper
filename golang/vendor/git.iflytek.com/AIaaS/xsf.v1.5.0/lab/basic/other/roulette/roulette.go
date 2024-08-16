package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	probability := []int{1, 2, 4}
	for {
		fmt.Println(roulette(probability))
		time.Sleep(time.Millisecond * 200)
	}
}
func roulette(probability []int) int {
	randIx := rand.Intn(len(probability))
	var pointer int
	for i := 0; i < len(probability); i++ {
		pointer += probability[i]
		if randIx < probability[i] {
			return i
		}
	}
	return -1
}
