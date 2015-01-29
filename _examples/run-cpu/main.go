package main

import (
	"fmt"
	"math"
	"os"
	"time"
)

func main() {
	fmt.Printf("starting...\n")
	i := 0
	sum := 0.0
	stop := time.After(10 * time.Second)
	for {
		select {
		default:
			i++
			sum += float64(i+1) * math.Exp(float64(i+1))
		case <-stop:
			fmt.Printf("boom!\n")
			os.Exit(0)
		}
	}
}
