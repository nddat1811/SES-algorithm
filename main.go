package main

import (
	"log"
	"os"
	"sync"

	"github.com/nddat1811/SES-algorithm/network"
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

// func initLog(iid int) {
// 	generalLogFile, err := os.OpenFile(fmt.Sprintf("./static/logs/%02d__general.log", iid), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer generalLogFile.Close()

// 	senderLogFile, err := os.OpenFile(fmt.Sprintf("./static/logs/%02d__sender.log", iid), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer senderLogFile.Close()

// 	receiverLogFile, err := os.OpenFile(fmt.Sprintf("./static/logs/%02d__receiver.log", iid), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer receiverLogFile.Close()

// 	consoleHandler := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
// 	fileHandler := log.New(io.MultiWriter(generalLogFile, receiverLogFile), "", log.LstdFlags)

// 	log.SetOutput(io.MultiWriter(consoleHandler, fileHandler))
// }

func main() {
	numberProcess := 2
	network.RegisterExitSignal()
	var wg sync.WaitGroup
	for i := 0; i < numberProcess; i++ {
		wg.Add(1)
		go func(instanceID int) {
			defer wg.Done()

			network := network.NewNetwork(instanceID, numberProcess)
			defer network.SafetyClose()
			network.StartSending()
			network.StartListening()
		}(i)
	}

	wg.Wait()
}
