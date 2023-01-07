package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tarunKoyalwar/gochan/channels"
)

func main() {
	// drain sink
	for n := 1; n < 44; n++ {
		fmt.Printf("N is %v\n", n)
		dg := &sync.WaitGroup{}
		sink := make(chan struct{})
		dg.Add(1)
		go DrainSinkChan(sink, time.Millisecond, "sink", dg)

		srcs := []chan struct{}{}
		for i := 0; i < n; i++ {
			srcs = append(srcs, make(chan struct{}))
		}

		err := channels.JoinChannel(context.TODO(), sink, srcs...)
		if err != nil {
			panic(err)
		}

		sg := &sync.WaitGroup{}
		// now push to source
		for _, v := range srcs {
			sg.Add(1)
			go FillSource(v, 10, sg)
		}

		sg.Wait()
		dg.Wait()
	}

	// for n := 2; n < 20; n++ {
	// 	fmt.Printf("N is %v\n", n)
	// 	source := make(chan struct{})
	// 	sinks := []chan struct{}{}
	// 	wg := &sync.WaitGroup{}

	// 	for i := 0; i < n; i++ {
	// 		sinks = append(sinks, make(chan struct{}))
	// 		wg.Add(1)
	// 		go split.Worker(sinks[i], time.Millisecond, strconv.Itoa(i), wg)
	// 	}

	// 	channels.SplitChannel(context.TODO(), source, sinks...)

	// 	split.Manager(source, 10)

	// 	wg.Wait()
	// }

	// wg := &sync.WaitGroup{}
	// sinks := map[int]chan struct{}{}
	// for i := 0; i < 3; i++ {
	// 	sinks[i] = make(chan struct{})
	// 	wg.Add(1)
	// 	go split.Worker(sinks[i], time.Millisecond, strconv.Itoa(i), wg)
	// }
	// sinks[4] = nil
	// sinks[3] = nil
	// for _, v := range sinks {
	// 	if v == nil {
	// 		fmt.Println("channel is nil")
	// 	}
	// }

	// source := make(chan struct{})
	// err := channels.SplitChannel(source, sinks)
	// if err != nil {
	// 	panic(err)
	// }

}

// Simulates joining a channel
func SimulateJoin() {

	dispatcher := func(sg *sync.WaitGroup, ch chan<- struct{}) {
		defer sg.Done()
		for i := 0; i < 1000; i++ {
			ch <- struct{}{}
		}
		close(ch)
	}

	sg := &sync.WaitGroup{}
	wg := &sync.WaitGroup{}

	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	ch3 := make(chan struct{})
	rec := make(chan struct{})
	if err := channels.JoinChannel(context.TODO(), rec, ch1, ch2, ch3); err != nil {
		panic(err)
	}
	// joiner.Joiner(map[string]chan struct{}{"1": ch1, "2": ch2, "3": ch3}, rec)

	wg.Add(1)
	go func() {
		defer wg.Done()
		counter := 0
		for {
			_, ok := <-rec
			if !ok {
				fmt.Printf("stopped with count %v\n", counter)
				break
			}
			// fmt.Printf("g")
			counter++
		}
	}()

	fmt.Println("dispatching")
	sg.Add(3)
	go dispatcher(sg, ch1)
	go dispatcher(sg, ch2)
	go dispatcher(sg, ch3)

	// wg.Add(3)
	// go dispatcher(wg, ch1)
	// go dispatcher(wg, ch2)
	// go dispatcher(wg, ch3)

	sg.Wait()
	wg.Wait()
}

var FillSource = func(work chan<- struct{}, count int, sg *sync.WaitGroup) {
	defer sg.Done()
	// send work
	for i := 100; i < 100+count; i++ {
		log.Println("sending work")
		work <- struct{}{}
	}
	log.Println("Sent total of 10 tasks")
	close(work)
}

var DrainSinkChan = func(ch <-chan struct{}, sleeptime time.Duration, chanId string, wg *sync.WaitGroup) {
	defer wg.Done()
	if ch == nil {
		fmt.Println("worker chan is nil")
	}
	for {
		val, ok := <-ch
		if !ok {
			log.Printf("sink channel %v closed\n", chanId)
			break
		}
		// time.Sleep(sleeptime) // to simulate blocking network i/o
		log.Printf("completed work %v at chan %v\n", val, chanId)
	}
}
