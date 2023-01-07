package channels

import (
	"context"
	"fmt"
	"strconv"
	"sync"
)

// Internal Sync WaitGroup
var wg *sync.WaitGroup = &sync.WaitGroup{}

// MaxDrain specifes max number of buffers to drain at once
var MaxDrain int = 3
var Threshold int = 4

// // SplitChannel
// func Split[T any](source chan T, sinkss map[int]chan T) error {
// 	if source == nil {
// 		return fmt.Errorf("source is nil")
// 	}
// 	if len(sinkss) == 0 {
// 		return fmt.Errorf("sink is nil")
// 	}

// 	sendqueue := map[int]chan<- T{}
// 	for k, v := range sinkss {
// 		if len(sendqueue) == 5 {
// 			break
// 		}
// 		sendqueue[k] = v
// 	}

// 	worker := func(src chan T, sinks map[int]chan<- T, sg *sync.WaitGroup) {
// 		if len(sinks) != 5 {
// 			panic(len(sinks))
// 		}
// 		fmt.Println(sinks)
// 		defer sg.Done()
// 		// Buffer threshold is 5 for each sink
// 		buffer := map[int][]T{}
// 		backlog := map[int]struct{}{}

// 		// add waiting packet to all sinks except the successful one
// 		addtoBuff := func(id int, data T) {
// 			for k, v := range sinks {
// 				// nil channels should have no buffer
// 				if k != id && v != nil {
// 					if buffer[k] == nil {
// 						buffer[k] = []T{}
// 					}
// 					buffer[k] = append(buffer[k], data)
// 					if len(buffer) == 5 {
// 						backlog[k] = struct{}{}
// 					}
// 				}
// 			}
// 		}

// 		drainChanBuffer := func() {
// 			for chanID := range backlog {
// 				if sinks[chanID] != nil {
// 					for _, item := range buffer[chanID] {
// 						sinks[chanID] <- item
// 					}
// 					buffer[chanID] = []T{}
// 					delete(backlog, chanID)
// 					return
// 				}
// 			}
// 		}

// 	forloop:
// 		for {
// 			switch {
// 			case len(backlog) > 0:
// 				// if buffer of any channel is above 5 drain it
// 				drainChanBuffer()
// 				// if it is true
// 			default:
// 				// send to sinks
// 				w, ok := <-source
// 				if !ok {
// 					break forloop
// 				}
// 				select {
// 				case sinks[0] <- w:
// 					addtoBuff(0, w)
// 				case sinks[1] <- w:
// 					addtoBuff(1, w)
// 				case sinks[2] <- w:
// 					addtoBuff(2, w)
// 				case sinks[3] <- w:
// 					addtoBuff(3, w)
// 				case sinks[4] <- w:
// 					addtoBuff(4, w)
// 				}
// 			}
// 		}

// 		// empty all remaining buffer
// 		for k, v := range buffer {
// 			if sinks[k] != nil {
// 				if len(v) > 0 {
// 					for _, item := range v {
// 						sinks[k] <- item
// 					}
// 				}
// 				close(sinks[k])
// 			}
// 		}

// 	}

// 	wg.Add(1)
// 	go worker(source, sendqueue, wg)

// 	return nil
// }

// Splits one source channel to sinks with given threshold(buff reset)
func SplitChannel[T any](ctx context.Context, src chan T, sinks ...chan T) error {
	if src == nil {
		return fmt.Errorf("source channel is nil")
	}

	sinkqueue := []chan<- T{}
	// create blocks of size 5
	for _, v := range sinks {
		if v == nil {
			return fmt.Errorf("given sink channel is nil")
		}
		sinkqueue = append(sinkqueue, v)
	}
	fmt.Printf("actual sink queue size is %v\n", len(sinkqueue))

	if len(sinkqueue)%5 != 0 {
		remaining := 5 - (len(sinkqueue) % 5)
		for i := 0; i < remaining; i++ {
			// add nil channels these are automatically kicked out of select
			sinkqueue = append(sinkqueue, nil)
		}
	}

	if len(sinkqueue)/5 == 1 {
		wg.Add(1)
		go splitchanworker(ctx, wg, Threshold, MaxDrain, src, sinkqueue...)
		fmt.Printf("Starting worker chan with sinksize %v\n", len(sinkqueue))
		return nil
	}

	fmt.Printf("sinkqueue > 5 starting relay channels with %v relays for %v sinks\n", len(sinkqueue)/5, len(sinkqueue))

	relaychannels := []chan T{}
	// create sink groups with each group of length 5
	groups := [][]chan<- T{}
	tmp := []chan<- T{}
	for i, v := range sinkqueue {
		if i != 0 && i%5 == 0 {
			groups = append(groups, tmp)
			tmp = []chan<- T{}
		}
		tmp = append(tmp, v)
	}
	if len(tmp) > 0 {
		groups = append(groups, tmp)
	}

	for _, v := range groups {
		wg.Add(1)
		relay := make(chan T)
		relaychannels = append(relaychannels, relay)
		go splitchanworker(ctx, wg, Threshold, MaxDrain, relay, v...)
	}
	fmt.Printf("Total relays %v , total groups %v\n", len(relaychannels), len(groups))
	fmt.Printf("starting apex split with %v relays\n", len(relaychannels))
	// recursion use sources to feed relays
	return SplitChannel(ctx, src, relaychannels...)
}

// splitchanworker is actual worker
func splitchanworker[T any](ctx context.Context, sg *sync.WaitGroup, threshold int, MaxDrains int, src chan T, sinkchans ...chan<- T) {
	defer sg.Done()
	sink := map[int]chan<- T{}
	count := 0
	for _, v := range sinkchans {
		sink[count] = v
		count++
	}
	if count != 5 {
		panic("something went wrong count is " + strconv.Itoa(count))
	}

	// MaxDrains if multiple buffers are full
	if MaxDrains == 0 {
		// set it to 3
		MaxDrains = 3
	}
	if threshold == 0 {
		threshold = 5
	}

	// backlog tracks sink channels whosse buffers have reached threshold
	backlog := map[int]struct{}{}
	buffer := map[int][]T{}

	// Helper Functions
	// addToBuff adds data to buffers where data was not sent
	addToBuff := func(id int, value T) {
		for sid, s := range sink {
			if s != nil && id != sid {
				if buffer[sid] == nil {
					buffer[sid] = []T{}
				}
				// add to buffer
				buffer[sid] = append(buffer[sid], value)
				if len(buffer[sid]) == threshold {
					backlog[sid] = struct{}{}
				}
			}
		}
	}

	//drain buffer of given channel
	drainAndReset := func() {
		// drain buffer of given channel since threshold has been breached
		// get pseudo random
		count := 0
		for chanID := range backlog {
			if sink[chanID] != nil {
				for _, item := range buffer[chanID] {
					select {
					case <-ctx.Done():
						return
					case sink[chanID] <- item:
					}
				}
				buffer[chanID] = []T{}
				delete(backlog, chanID)
				count++
				if count == MaxDrains {
					// skip for now
					return
				}
			}
		}
	}

forloop:
	for {
		switch {
		case len(backlog) > 0:
			// if buffer of any channel has reached threshold
			drainAndReset()
			// if it is true
		default:
			// send to sinks
			w, ok := <-src
			if !ok {
				break forloop
			}
			select {
			case <-ctx.Done():
				return
			case sink[0] <- w:
				addToBuff(0, w)
			case sink[1] <- w:
				addToBuff(1, w)
			case sink[2] <- w:
				addToBuff(2, w)
			case sink[3] <- w:
				addToBuff(3, w)
			case sink[4] <- w:
				addToBuff(4, w)
			}
		}
	}

	// empty all remaining buffer
	for id, ch := range sink {
		if ch != nil {
			if len(buffer[id]) > 0 {
				for _, item := range buffer[id] {
					select {
					case <-ctx.Done():
						return
					case ch <- item:
					}
				}
			}
			close(ch)
		}
	}
}
