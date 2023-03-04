package network

import (
	"os"
	"os/signal"
)

func ExitGracefully(signalChan chan os.Signal, doneChan chan bool) {
	<-signalChan
	doneChan <- true
}

func RegisterExitSignal() chan bool {
	signalChan := make(chan os.Signal, 1)
	doneChan := make(chan bool, 1)

	signal.Notify(signalChan, os.Interrupt, os.Kill)

	go ExitGracefully(signalChan, doneChan)

	return doneChan
}
