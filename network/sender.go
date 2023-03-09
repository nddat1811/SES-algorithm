package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	s "github.com/nddat1811/SES-algorithm/SES"
	c "github.com/nddat1811/SES-algorithm/constant"
)

type SenderWorker struct {
	IP            string
	Port          int
	InstanceID    int
	DestinationID int
	SesClock      *s.SES
	MessageCount  int
	ShutdownFlag  chan bool
}

func NewSenderWorker(instanceID int, destinationID int, ip string, port int, sesClock *s.SES) *SenderWorker {
	return &SenderWorker{
		IP:            ip,
		Port:          port,
		InstanceID:    instanceID,
		DestinationID: destinationID,
		SesClock:      sesClock,
		MessageCount:  0,
		ShutdownFlag:  make(chan bool),
	}
}

func (sw *SenderWorker) Start() {
	sender, err := net.Dial("tcp", fmt.Sprintf("%s:%d", sw.IP, sw.Port))
	for err != nil {
		log.Printf("SENDER #%d: connection to %s:%d failed, retry after 500 milliseconds.", sw.InstanceID, sw.IP, sw.Port)
		time.Sleep(500 * time.Millisecond)
		sender, err = net.Dial("tcp", fmt.Sprintf("%s:%d", sw.IP, sw.Port))
	}
	defer sender.Close()

	for {
		select {
		case <-sw.ShutdownFlag:
			sender.Close()
			return
		default:
			if sw.MessageCount == c.MAX_MESSAGE {
				sender.Close()
				return
			}

			sw.MessageCount++

			message := []byte(fmt.Sprintf("Message number %d from process %d", sw.MessageCount, sw.InstanceID))
			message = sw.SesClock.Send(sw.DestinationID, message)

			dataSize := len(message)
			buf := new(bytes.Buffer)
			err := binary.Write(buf, binary.BigEndian, int32(dataSize))
			if err != nil {
				log.Println("binary.Write failed:", err)
				sender.Close()
				return
			}
			message = append(buf.Bytes(), message...)

			_, err = sender.Write(message)
			if err != nil {
				log.Println("sender.Write failed:", err)
				sender.Close()
				return
			}
			
			time.Sleep(time.Duration(rand.Int63n(int64(c.MAX_DELAY-c.MIN_DELAY)) + int64(c.MIN_DELAY))) 
			// Stop sending for 100 - 1000 time.Millisecond
		}
	}
}

func (sw *SenderWorker) Stop() {
	close(sw.ShutdownFlag)
}
