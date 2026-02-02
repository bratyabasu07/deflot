package ui

import (
	"fmt"
	"sync"
	"time"
)

// PrintOutro displays the end-of-run visual.
// It ensures no sleeps happen in the main thread by using a goroutine waiter.
func PrintOutro(isJSON, isStdout bool) {
	if isJSON || isStdout || !isTTY() {
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		fmt.Println("========== DEFLOT ==========")
		time.Sleep(50 * time.Millisecond)
		fmt.Println("Scan complete.")
		time.Sleep(50 * time.Millisecond)
		fmt.Println("============================")
	}()

	wg.Wait()
}
