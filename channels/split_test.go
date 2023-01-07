package channels_test

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/tarunKoyalwar/gochan/channels"
)

var FillSource = func(work chan<- struct{}, count int) {
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

func TestSinkCount(t *testing.T) {
	for n := 1; n < 101; n++ {
		source := make(chan struct{})
		sinks := []chan struct{}{}
		for i := 0; i < n; i++ {
			sinks = append(sinks, make(chan struct{}))
		}
		t.Logf("N Count : %v\n", n)
		channels.SplitChannel(context.TODO(), source, sinks...)
	}
}

func TestExecution(t *testing.T) {
	for n := 2; n < 44; n++ {
		source := make(chan struct{})
		sinks := []chan struct{}{}
		wg := &sync.WaitGroup{}
		for i := 0; i < n; i++ {
			sinks = append(sinks, make(chan struct{}))
			wg.Add(1)
			go DrainSinkChan(sinks[i], time.Millisecond, strconv.Itoa(i), wg)
		}
		channels.SplitChannel(context.TODO(), source, sinks...)
		FillSource(source, 10)
		wg.Wait()
	}
}
