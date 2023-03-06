package network

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"

	s "github.com/nddat1811/SES-algorithm/SES"
	c "github.com/nddat1811/SES-algorithm/constant"
)

type ReceiverWorker struct {
	Connection   net.Conn
	Address      net.Addr
	SesClock     *s.SES
	ShutdownFlag chan bool
	MessageCount int
	Noise        [][]byte
}

func NewReceiverWorker(connection net.Conn, address net.Addr, sesClock *s.SES) *ReceiverWorker {
	return &ReceiverWorker{
		Connection:   connection,
		Address:      address,
		SesClock:     sesClock,
		ShutdownFlag: make(chan bool),
		Noise:        [][]byte{},
	}
}
func (rw *ReceiverWorker) WriteLog(data string) {
	file, err := os.OpenFile("./static/logs/connect.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open general log file: %v", err)
	}
	
	defer file.Close()

	log.SetOutput(file)
	log.Println(data)
}
func (rw *ReceiverWorker) Start() {
	defer rw.Connection.Close()
	for {
		// Check for shutdown signal
		select {
		case <-rw.ShutdownFlag:
			rw.Connection.Close()
			e := rw.Connection.Close()
			if e != nil {
				fmt.Println("\n\n\n\n err: ", e)
			}
			fmt.Printf("RECEIVER : close connection to %s", rw.Address.String())
			return
		default:
			dataSizeBytes := make([]byte, c.INT_SIZE)
			_, err := rw.Connection.Read(dataSizeBytes)
			if err != nil {
				log.Printf("RECEIVER %s: error reading data size: %v\n", rw.Address.String(), err)
				rw.Connection.Close()
				return
			}

			dataSize := int(binary.BigEndian.Uint32(dataSizeBytes))
			rw.MessageCount++
			if rw.MessageCount == c.MAX_MESSAGE {
				// Close connection
				packet := make([]byte, dataSize)
				_, err = rw.Connection.Read(packet)
				if err != nil {
					log.Printf("RECEIVER #%s: error reading data packet: %v", rw.Address.String(), err)
					return
				}

				rw.Noise = append(rw.Noise, packet)
				for i := len(rw.Noise) - 1; i >= 0; i-- {
					packet := rw.Noise[i]
					rw.SesClock.Deliver(packet)
				}
				close(rw.ShutdownFlag)
				e := rw.Connection.Close()
				if e != nil {
					fmt.Println("err close connection : ", e)
				}
				rw.WriteLog(fmt.Sprintf("RECEIVER : close connection to %s", rw.Address.String()))
				return
			}

			packet := make([]byte, dataSize)
			_, err = rw.Connection.Read(packet)
			if err != nil {
				log.Printf("RECEIVER #%s: error reading data packet: %v\n", rw.Address.String(), err)
				return
			}

			rw.Noise = append(rw.Noise, packet)
			if rand.Float32() > 0.1 && len(rw.Noise) > 0 {
				for i := len(rw.Noise) - 1; i >= 0; i-- {
					packet := rw.Noise[i]
					rw.SesClock.Deliver(packet)
				}
				rw.Noise = nil
			}
		}
	}
}

func (rw *ReceiverWorker) Stop() {
	rw.ShutdownFlag <- true
}
