package network

import (
	"encoding/binary"
	"log"
	"math/rand"
	"net"

	s "github.com/nddat1811/SES-algorithm/SES"
	c "github.com/nddat1811/SES-algorithm/constant"
)

type ReceiverWorker struct {
	Connection   net.Conn
	Address      net.Addr
	SesClock     *s.SES
	ShutdownFlag chan bool
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

func (rw *ReceiverWorker) Start() {
	defer rw.Connection.Close()
	for {
		// Check for shutdown signal
		select {
		case <-rw.ShutdownFlag:
			rw.Connection.Close()
			log.Printf("RECEIVER : close connection to %s\n", rw.Address.String())
			return
		default:
			dataSizeBytes := make([]byte, c.INT_SIZE)
			_, err := rw.Connection.Read(dataSizeBytes)
			if err != nil {
				log.Printf("RECEIVER %s: error reading data size: %v\n", rw.Address.String(), err)
				return
			}

			dataSize := int(binary.BigEndian.Uint32(dataSizeBytes))
			if dataSize == 0 {
				// Close connection
				for i := len(rw.Noise) - 1; i >= 0; i-- {
					packet := rw.Noise[i]
					rw.SesClock.Deliver(packet)
					log.Printf("RECEIVER #%s: Received message %s from %s\n", rw.Address.String(), string(packet), rw.Address.String())
				}
				close(rw.ShutdownFlag)
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
					log.Printf("RECEIVER #%s: Received message %s from %s\n", rw.Address.String(), string(packet), rw.Address.String())
				}
				rw.Noise = nil
			}
		}
	}
}

func (rw *ReceiverWorker) Stop() {
	rw.ShutdownFlag <- true
}
