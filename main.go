package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nddat1811/SES-algorithm/SES"
	"github.com/nddat1811/SES-algorithm/network"
)

func main() {
	numberProcess := 3
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Create a context to signal cancellation to the network goroutines
	ctx, cancel := context.WithCancel(context.Background())
	
	var wg sync.WaitGroup
	for i := 0; i < numberProcess; i++ {
		wg.Add(1)
		go func(instanceID int) {
			defer wg.Done()
			ses.InitLog(instanceID)
			network := network.NewNetwork(instanceID, numberProcess)
			defer network.SafetyClose()
			for {
				select {
				case <-ctx.Done():
					network.SafetyClose()
					return
				default:
					network.StartSending()
					network.StartListening()
				}
			}
		}(i)
	}
	// Wait for a signal from the OS
	<-c

	// Send a quit signal to the network goroutines
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()

	os.Exit(0)
}
