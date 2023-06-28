package main

import (
	"github.com/rgierusz/ex1/multithreading"
	"log"
	"time"
)

func main() {
	//log.Println("Proxifier started...")
	//
	//server.InitServer()

	testMv()
}

func testMv() {
	mv := multithreading.NewMovingWindow((3 * time.Second).Milliseconds(), multithreading.AverageProcessor, multithreading.SumProcessor)

	go printMwStatusEverySec(mv)

	time.Sleep(time.Second)
	mv.AddValue(3)

	time.Sleep(time.Millisecond * 1500) // 1.5s
	mv.AddValue(2)

	time.Sleep(time.Second * 2)
	mv.AddValue(1)

	time.Sleep(time.Second * 5)
}

func printMwStatusEverySec(mv *multithreading.MovingWindow) {
	var counter int

	for {
		log.Printf("--- %vs ---", counter)
		mv.ProcessWithProcessors()
		counter++

		time.Sleep(time.Second)
	}
}
