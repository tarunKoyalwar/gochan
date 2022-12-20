package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

var wg *sync.WaitGroup = &sync.WaitGroup{}

var worker = func(ch <-chan struct{}, sleeptime time.Duration, chanId string) {
	defer wg.Done()
	if ch == nil {
		fmt.Println("worker chan is nil")
	}
	for {
		val, ok := <-ch
		if !ok {
			log.Printf("closing sink channel %v\n", chanId)
			break
		}
		// time.Sleep(sleeptime) // to simulate blocking network i/o
		log.Printf("completed work %v at chan %v\n", val, chanId)
	}
}

var manager = func(work chan<- struct{}) {
	// send work
	for i := 100; i < 110; i++ {
		log.Println("sending work")
		work <- struct{}{}
	}
	log.Println("Sent total of 10 tasks")
	close(work)
}

func main() {
	// Split single channel stringo multiple channels

	source := make(chan struct{}, 5)
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	ch3 := make(chan struct{})

	splitchan(source, ch1, ch2, ch3) // split 1 channel into 3

	// start working
	wg.Add(3)
	go worker(ch1, time.Duration(1)*time.Millisecond, "1")
	go worker(ch2, time.Duration(3)*time.Millisecond, "2")
	go worker(ch3, time.Duration(5)*time.Millisecond, "3")

	// assing work
	manager(source)

	wg.Wait()

}

func splitchan(source chan struct{}, ch1 chan struct{}, ch2 chan struct{}, ch3 chan struct{}) {
	// currently fixed size 3

	wg.Add(1)
	go func() {
		defer wg.Done()
		// tasks buff
		ch2buff := []struct{}{}
		ch1buff := []struct{}{}
		ch3buff := []struct{}{}

		for {
			switch {
			case len(ch1buff) > 2:
				for _, v := range ch1buff {
					ch1 <- v
				}
				ch1buff = []struct{}{}
				//decrease ch1buff by only sending to ch1

			case len(ch2buff) > 2:
				for _, v := range ch2buff {
					ch2 <- v
				}
				ch2buff = []struct{}{}
				//decrease ch2buff by only sending to ch2

			case len(ch3buff) > 2:
				for _, v := range ch3buff {
					ch3 <- v
				}
				ch3buff = []struct{}{}
				// decrease ch3buff by only sending to ch3

			default:
				sourceval, ok := <-source
				log.Println("got work from source")
				if !ok {
					fmt.Println("notok")
					// implement so that `default` case is not used anymore
					// for now just close all channels and return
					close(ch1)
					close(ch2)
					close(ch3)
					return
				}
				select {

				case ch1 <- sourceval:
					log.Println("only sent to ch1")
					ch2buff = append(ch2buff, sourceval)
					ch3buff = append(ch3buff, sourceval)
				case ch2 <- sourceval:
					log.Println("only sent to ch2")
					ch1buff = append(ch1buff, sourceval)
					ch3buff = append(ch3buff, sourceval)
				case ch3 <- sourceval:
					log.Println("only sent to ch3")
					ch2buff = append(ch2buff, sourceval)
					ch1buff = append(ch1buff, sourceval)

				}
			}
		}
	}()
}