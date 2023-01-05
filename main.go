package main

import (
	"fmt"
	"sync"

	"github.com/tarunKoyalwar/gochan/joiner"
)

func main() {
}

// Simulates joining a channel
func SimulateJoin() {

	dispatcher := func(wg *sync.WaitGroup, ch chan struct{}) {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			ch <- struct{}{}
		}
	}

	wg := &sync.WaitGroup{}

	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	ch3 := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		rec := joiner.Joiner(map[string]chan struct{}{"1": ch1, "2": ch2, "3": ch3})
		counter := 0
		for {
			_, ok := <-rec
			if !ok {
				fmt.Printf("stopped with count %v\n", counter)
				break
			}
			counter++
		}
	}()

	wg.Add(3)
	go dispatcher(wg, ch1)
	go dispatcher(wg, ch2)
	go dispatcher(wg, ch3)

	wg.Wait()

}
