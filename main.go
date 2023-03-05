package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nddat1811/SES-algorithm/network"
	s "github.com/nddat1811/SES-algorithm/SES"
)

func getCustomConsoleHandler() *log.Logger {
	cHandler := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	return cHandler
}

func getCustomFileHandler(directory string) *log.Logger {
	file, err := os.OpenFile(directory, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	fHandler := log.New(file, "", log.Ldate|log.Ltime)
	return fHandler
}

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
			s.InitLog(instanceID)
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
