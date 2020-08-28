package main

import (
	"flag"
	"fmt"
	"math"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "lifetime of the process")
	flag.Parse()

	fmt.Printf("starting...\n")
	i := 0
	sum := 0.0
	mem := make([]byte, 1024)

	stop := time.After(*timeout)
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

loop:
	for {
		select {
		default:
			i++
			sum += float64(i+1) * math.Exp(float64(i+1))
		case <-tick.C:
			mem = append(mem, make([]byte, 10*1024*1024)...)
			fmt.Printf("tick\n")
		case <-stop:
			fmt.Printf("boom!\n")
			break loop
		}
	}
}
