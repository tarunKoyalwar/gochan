package channels

import (
	"context"
	"fmt"
	"sync"
)

// JoinChannels Joins Many Channels to Create One
func JoinChannel[T any](ctx context.Context, sink chan T, sources ...chan T) error {
	if sink == nil {
		return fmt.Errorf("join: sink is nil")
	}
	if len(sources) == 0 {
		return fmt.Errorf("join: sources cannot be zero")
	}
	for _, v := range sources {
		if v == nil {
			return fmt.Errorf("join: source is nil")
		}
	}
	// check length of sources
	srcCount := len(sources)

	if srcCount%5 != 0 {
		for i := 0; i < 5-(srcCount%5); i++ {
			sources = append(sources, nil)
		}
	}

	if len(sources) == 5 {
		wg.Add(1)
		go joinWorker(ctx, wg, sink, sources...)
		return nil
	}

	groups := [][]chan T{}
	tmp := []chan T{}
	for i, v := range sources {
		if i != 0 && i%5 == 0 {
			groups = append(groups, tmp)
			tmp = []chan T{}
		}
		tmp = append(tmp, v)
	}
	if len(tmp) > 0 {
		groups = append(groups, tmp)
	}

	relays := []chan T{}
	for _, v := range groups {
		wg.Add(1)
		relay := make(chan T)
		relays = append(relays, relay)
		go joinWorker(ctx, wg, relay, v...)
	}

	// Recursion
	return JoinChannel(ctx, sink, relays...)
}

func joinWorker[T any](ctx context.Context, sg *sync.WaitGroup, sink chan T, sources ...chan T) {
	defer sg.Done()
	if len(sources) != 5 {
		panic(fmt.Sprintf("join: worker only supports 5 sources got %v", len(sources)))
	}
	if sink == nil {
		panic("join: sink cannot be nil")
	}

	// recieve only channels
	src := map[int]<-chan T{}
	for k, v := range sources {
		src[k] = v
	}

	for src[0] != nil || src[1] != nil || src[2] != nil || src[3] != nil || src[4] != nil {
		select {
		case <-ctx.Done():
			close(sink)
			return
		case w, ok := <-src[0]:
			if !ok {
				src[0] = nil
				continue
			}
			sink <- w
		case w, ok := <-src[1]:
			if !ok {
				src[1] = nil
				continue
			}
			sink <- w
		case w, ok := <-src[2]:
			if !ok {
				src[2] = nil
				continue
			}
			sink <- w
		case w, ok := <-src[3]:
			if !ok {
				src[3] = nil
				continue
			}
			sink <- w
		case w, ok := <-src[4]:
			if !ok {
				src[4] = nil
				continue
			}
			sink <- w
		}
	}
	close(sink)
}
