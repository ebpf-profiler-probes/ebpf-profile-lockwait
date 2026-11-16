// Copyright The ebpf-profile-lockwait Authors
// SPDX-License-Identifier: Apache-2.0

// Mutex contention demo for verifying lockwait (futex wait).
//
//	go run ./example
//	go run ./example -workers=16 -hold=2ms -duration=30s
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	workers := flag.Int("workers", 8, "number of contending goroutines")
	hold := flag.Duration("hold", time.Millisecond, "how long the lock holder keeps the mutex")
	duration := flag.Duration("duration", 0, "stop after this long (0 = until Ctrl-C)")
	flag.Parse()

	if *workers < 2 {
		fmt.Fprintln(os.Stderr, "workers must be >= 2 to create contention")
		os.Exit(1)
	}

	fmt.Printf("lockwait contention demo: workers=%d hold=%s\n", *workers, *hold)
	fmt.Println("press Ctrl-C to stop")

	var mu sync.Mutex
	stop := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				mu.Lock()
				time.Sleep(*hold)
				mu.Unlock()
			}
		}()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	if *duration > 0 {
		select {
		case <-time.After(*duration):
		case <-sig:
		}
	} else {
		<-sig
	}

	close(stop)
	wg.Wait()
	fmt.Println("stopped")
}
