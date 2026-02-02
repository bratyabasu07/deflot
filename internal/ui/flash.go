package ui

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/elliot/deflot/internal/filters"
)

// Flasher handles "first-time" notifications for specific categories.
type Flasher struct {
	disabled bool

	// Atomic flags (0 = not seen, 1 = seen)
	seenSecret  uint32
	seenConfig  uint32
	seenCloud   uint32
	seenJS      uint32
	seenLog     uint32
	seenArchive uint32
}

// NewFlasher creates a new flasher.
func NewFlasher(isJSON, isStdout bool) *Flasher {
	return &Flasher{
		disabled: isJSON || isStdout,
	}
}

// Notify triggers a flash message if this is the first time seeing this category.
func (f *Flasher) Notify(category string) {
	if f.disabled {
		return
	}

	var ptr *uint32

	switch category {
	case filters.CatSecret:
		ptr = &f.seenSecret
	case filters.CatConfig:
		ptr = &f.seenConfig
	case filters.CatCloud:
		ptr = &f.seenCloud
	case filters.CatJS:
		ptr = &f.seenJS
	case filters.CatLog:
		ptr = &f.seenLog
	case filters.CatArchive:
		ptr = &f.seenArchive
	default:
		return
	}

	// Compare and Swap: If 0, set to 1. Returns true if swapped.
	if atomic.CompareAndSwapUint32(ptr, 0, 1) {
		// [!] FIRST SECRET
		// Use \r to overwrite anything currently on the line (like HUD)
		// and \n to push it to history so HUD can redraw on next line.
		fmt.Printf("\r[!] FIRST %s\n", strings.ToUpper(category))
	}
}
