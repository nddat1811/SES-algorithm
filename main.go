package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	ses "github.com/nddat1811/SES-algorithm/SES"
	"github.com/nddat1811/SES-algorithm/network"
)

func registerExitSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-c
		os.Exit(1)
	}()
}

func main() {
	numberProcess, _ := strconv.Atoi(os.Args[1])
	instanceID, _ := strconv.Atoi(os.Args[2])
	registerExitSignal()
	fmt.Println("n: ", numberProcess)
	fmt.Println("i: ", instanceID)

	ses.InitLog(instanceID, numberProcess)
	network := network.NewNetwork(instanceID, numberProcess)
	defer func() {
		if r := recover(); r != nil {
			network.SafetyClose()
		}
	}()
	network.StartSending()
	network.StartListening()
}
