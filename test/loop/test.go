package main

import (
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

// A simple function to consume CPU time
func busyWork() {
	for i := 0; i < 10000000; i++ {
		_ = i * i
	}
}

// Simulate a more complex function with variable execution time
func benchmarkFunction(n int) {
	// Create a large slice and perform operations
	data := make([]int, n)
	for i := 0; i < n; i++ {
		data[i] = rand.Intn(1000)
	}

	// Sort the slice using bubble sort (inefficient for large n)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if data[j] > data[j+1] {
				data[j], data[j+1] = data[j+1], data[j]
			}
		}
	}

	// Perform additional CPU work
	busyWork()
}

// Another function to show in the profile
func helper() {
	busyWork()
	time.Sleep(time.Millisecond)
}

func main() {
	// Create CPU profile file
	f, err := os.Create("cpu.pprof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close()

	// Start CPU profiling
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	// Run initial work
	for i := 0; i < 5; i++ {
		busyWork()
		helper()
		time.Sleep(time.Millisecond * 10)
	}

	// Run benchmark with larger dataset and multiple iterations
	log.Println("Starting benchmark with larger dataset")
	for i := 0; i < 100; i++ {
		benchmarkFunction(10000)
	}
	log.Println("Benchmark completed")

	// Run more work
	for i := 0; i < 5; i++ {
		busyWork()
		helper()
		time.Sleep(time.Millisecond * 10)
	}
}
