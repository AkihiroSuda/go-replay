# GoReplay: record-less semi-deterministic replayer for Go programs

[![GoDoc](https://godoc.org/github.com/AkihiroSuda/go-replay?status.svg)](https://godoc.org/github.com/AkihiroSuda/go-replay)
[![Build Status](https://travis-ci.org/AkihiroSuda/go-replay.svg?branch=master)](https://travis-ci.org/AkihiroSuda/go-replay)
[![Go Report Card](https://goreportcard.com/badge/github.com/AkihiroSuda/go-replay)](https://goreportcard.com/report/github.com/AkihiroSuda/go-replay)

GoReplay replays concurrent Go programs _semi-deterministically_ without recording concrete events.

## How it works
GoReplay injects `time.Sleep(time.Duration(hash(seed, context, stack)) % maxInterval)` to arbitrary execution points of the target Go program.

 * `seed`: environmental variable `GRSEED`. You can replay the execution by remembering and setting `GRSEED`. The default value is empty value and it disables GoReplay.
 * `context`: optional `[]byte` slice.
 * `stack`: runtime stack. (planned)
 * `maxInterval`: environmental variable `GRMAX` (should be `time.Duration` string). The default value is `10ms`.

## Usage

Follow the example: [`example/ex01`](example/ex01).

```go
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
```

The result is non-deterministic without GoReplay:
```
$ go run example/ex01/main.go
i=7
i=5
i=1
i=6
i=0
i=2
i=3
i=4
$ (for f in $(seq 1 10);do go run example/ex01/main.go | sha512sum -; done) | sort | uniq | wc -l
4
```

If you set `GRSEED`, the result becomes deterministic:
```
$ (for f in $(seq 1 10);do GRSEED=foo go run example/ex01/main.go | sha512sum -; done) | sort | uniq
7ea818bd9e800609ab8e360688d975189033a9b9277d1ad9c9c96c9013f4ffeb1cf05c16ace9de737dcedaa68bc99162d611ac60b7169a43b4f9b17b1665f121  -

$ (for f in $(seq 1 10);do GRSEED=bar go run example/ex01/main.go | sha512sum -; done) | sort | uniq
fc1c177ebbc58baa7e4960102cf26da25173559de0d7a08e00c37a4512eb0a579fc6b4b70f2c708fe1166a1c0641e8839305d5fe7e26624cabbce2cb8274d963  -
```

## Practical Examples

 - [example/etcd-5155](example/etcd-5155): data race ([coreos/etcd#5155](https://github.com/coreos/etcd/issues/5155))

## Hint
You may use [AspectGo](https://github.com/AkihiroSuda/golang-exp-aspectgo/tree/aspectgo-master/aspectgo)(WIP) for automatic instrumentation.

## Related project
 * [Earthquake: a programmable fuzzy scheduler for testing distributed systems](https://github.com/osrg/earthquake)

GoReplay is planned to be merged to Earthquake.
