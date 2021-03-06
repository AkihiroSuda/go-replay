package main

import (
	"fmt"
	"sync"

	"github.com/AkihiroSuda/go-replay"
)

func main() {
	n := 8
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i interface{}) {
			msg := fmt.Sprintf("i=%d", i)
			replay.Inject([]byte(msg))
			fmt.Printf("%s\n", msg)
			wg.Done()
		}(i)
	}
	wg.Wait()
}
