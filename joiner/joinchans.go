package joiner

import "sync"

/*
This package should join channels
https://medium.com/justforfunc/two-ways-of-merging-n-channels-in-go-43c0b57cd1de
*/

var Wg *sync.WaitGroup = &sync.WaitGroup{}

// Join all channels and exit when all channels are closed
func Joiner(allchans map[string]chan struct{}) <-chan struct{} {
	x := make(chan struct{})

	ch1 := allchans["1"]
	ch2 := allchans["2"]
	ch3 := allchans["3"]

	go func(c1 chan struct{}, c2 chan struct{}, c3 chan struct{}) {
		for c1 != nil || c2 != nil || c3 != nil {
			select {
			case w, ok := <-c1:
				if !ok {
					c1 = nil
					continue
				}
				x <- w
			case w, ok := <-c2:
				if !ok {
					c2 = nil
					continue
				}
				x <- w
			case w, ok := <-c3:
				if !ok {
					c3 = nil
					continue
				}
				x <- w
			}
		}
		close(x)
	}(ch1, ch2, ch3)
	return x
}
