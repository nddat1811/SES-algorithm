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
			return
		default:
			if sw.MessageCount == c.MAX_MESSAGE {
				return
			}

			sw.MessageCount++

			message := []byte(fmt.Sprintf("Message from %d, number %d", sw.InstanceID, sw.MessageCount))
			message = sw.SesClock.Send(sw.DestinationID, message)

			dataSize := len(message)
			buf := new(bytes.Buffer)
			err := binary.Write(buf, binary.BigEndian, int32(dataSize))
			if err != nil {
				log.Println("binary.Write failed:", err)
				return
			}
			message = append(buf.Bytes(), message...)

			_, err = sender.Write(message)
			if err != nil {
				log.Println("sender.Write failed:", err)
				return
			}

			//fmt.Printf("SENDER #%d: send messagesss %v to %s:%d", sw.InstanceID, message, sw.IP, sw.Port)

			time.Sleep(time.Duration(rand.Float64() * float64(time.Second))) // Stop sending for random time
		}
	}
}

func (sw *SenderWorker) Stop() {
	close(sw.ShutdownFlag)
}
