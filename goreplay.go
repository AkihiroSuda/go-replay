// Package goreplay provides GoReplay.
package goreplay

import (
	"hash/fnv"
	"log"
	"os"
	"time"
)

// Inject calls DefaultGoReplay.Inject().
func Inject(context []byte) {
	if DefaultGoReplay == nil {
		// should not panic here
		log.Printf("GoReplay called before initialization")
		return
	}
	DefaultGoReplay.Inject(context)
}

// GoReplay is an instance of GoReplay.
type GoReplay struct {
	// Enabled is true if this instance is enabled.
	Enabled bool
	// Debug is true if debugging mode.
	Debug bool
	// Seed is an arbitrary string for replaying execution.
	// If seed is an empty string, it disables GoReplay.
	Seed string
	// Max is the max value for delays injected by GoReplay.
	Max time.Duration
}

// Init initializes GoReplay.
func (gr *GoReplay) Init() error {
	// nothing to do currently
	return nil
}

// Inject injects a random delay using context.
// The delay can be replayed with the seed value.
// Context can be nil.
func (gr *GoReplay) Inject(context []byte) {
	if !gr.Enabled {
		return
	}
	t := gr.inject(context)
	if gr.Debug {
		log.Printf("t=%s (seed=%s, context=%s)",
			t, gr.Seed, context)
	}
	time.Sleep(t)
}

func (gr *GoReplay) inject(context []byte) time.Duration {
	h := fnv.New64a()
	h.Write([]byte(gr.Seed))
	if context != nil {
		h.Write([]byte(context))
	}
	// TODO: hash runtime.Stack()
	ui64 := h.Sum64()
	t := time.Duration(ui64 % uint64(gr.Max))
	return t
}

// DefaultGoReplay is the default instance of GoReplay.
var DefaultGoReplay *GoReplay

// DefaultSeed is the default seed for DefaultGoReplay.
const DefaultSeed = ""

// DefaultMax is the default max delay for DefaultGoReplay.
const DefaultMax = "10ms"

func init() {
	enabled := true

	debugStr := os.Getenv("GRDEBUG")
	debug := debugStr != ""

	seed := os.Getenv("GRSEED")
	if seed == "" {
		log.Printf("GRSEED is not set. Disabling GoReplay.")
		enabled = false
	}

	maxStr := os.Getenv("GRMAX")
	if maxStr == "" {
		maxStr = DefaultMax
	}
	max, err := time.ParseDuration(maxStr)
	if err != nil {
		log.Printf("Error while parsing GRMAX: %s", err)
		enabled = false
	}

	DefaultGoReplay = &GoReplay{
		Enabled: enabled,
		Debug:   debug,
		Seed:    seed,
		Max:     max,
	}
	err = DefaultGoReplay.Init()
	if err != nil {
		log.Printf("Error while initializing: %s", err)
		DefaultGoReplay.Enabled = false
	}
}
