package ui

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// StartIntro runs the cinemtaic startup sequence in a non-blocking goroutine.
// It returns a cancel function that must be called to stop the intro
// and clean up the display.
func StartIntro(isJSON, isStdout bool) func() {
	// 1. Rules: Auto-disable for --json, --stdout, non-TTY
	if isJSON || isStdout || !isTTY() {
		return func() {}
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		runSequence(ctx)
	}()

	return func() {
		cancel()
		wg.Wait()
		// Ensure cursor is visible and newline is printed after we're done
		fmt.Print("\033[?25h\n")
	}
}

func runSequence(ctx context.Context) {
	// Hide cursor
	fmt.Print("\033[?25l")

	// Print Logo (Instant)
	fmt.Println(Banner)

	phases := []string{
		"Recon Engine",
		"[ Online ]",
	}

	// Print initial line to reserve space
	fmt.Print("...")

	// Total duration target: ~500ms
	// Char count roughly 30 total chars across phases.
	// 12ms * 30 = 360ms. Plus some logic overhead.
	ticker := time.NewTicker(12 * time.Millisecond)
	defer ticker.Stop()

	phaseIdx := 0
	textIdx := 0
	currentText := ""

	for {
		select {
		case <-ctx.Done():
			// Instant stop: print final state to look clean
			// Leave the logo, just update the status line
			fmt.Print("\r\033[K[ Online ]")
			return
		case <-ticker.C:
			if phaseIdx >= len(phases) {
				return
			}

			target := phases[phaseIdx]

			// If we finished typing current phase
			if textIdx >= len(target) {
				phaseIdx++
				textIdx = 0
				currentText = ""
				continue
			}

			// Typewriter char
			currentText += string(target[textIdx])
			textIdx++

			// Render
			// We clear line \r\033[K then print
			fmt.Printf("\r\033[K%s", currentText)
		}
	}
}

// isTTY checks if stdout is a terminal
func isTTY() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
