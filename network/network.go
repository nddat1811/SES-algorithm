package network

import (
	"fmt"
	"log"
	"net"

	s "github.com/nddat1811/SES-algorithm/SES"
	c "github.com/nddat1811/SES-algorithm/constant"
)

type Network struct {
	IP            string
	InstanceID    int
	NumberProcess int
	Port          int
	Socket        net.Listener
	SenderList    []*SenderWorker
	ReceiverList  []*ReceiverWorker
	SesClock      *s.SES
	c             int
}

func NewNetwork(instanceID int, numberProcess int) *Network {
	port := c.PORT_OFFSET + instanceID
	sesClock := s.NewSES(instanceID, numberProcess)

	return &Network{
		IP:            c.IP_ADDR,
		InstanceID:    instanceID,
		NumberProcess: numberProcess,
		Port:          port,
		SenderList:    []*SenderWorker{},
		ReceiverList:  []*ReceiverWorker{},
		SesClock:      sesClock,
		c:             0,
	}
}

func (n *Network) StartListening() {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", n.IP, n.Port))
	if err != nil {
		log.Fatal(err)
	}
	n.Socket = listen
	defer listen.Close()

	log.Printf("Process %d: Server is Listening on %s:%d\n", n.InstanceID, n.IP, n.Port)
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		addr := conn.RemoteAddr()

		log.Printf("Open Connection with %s\n", addr)
		receiver := NewReceiverWorker(conn, addr, n.SesClock)
		go receiver.Start()
		n.ReceiverList = append(n.ReceiverList, receiver)
	}
}

func (n *Network) StartSending() {
	for instanceID := 0; instanceID < n.NumberProcess; instanceID++ {
		if n.InstanceID == instanceID {
			continue
		}
		sender := NewSenderWorker(n.InstanceID, instanceID, n.IP, c.PORT_OFFSET+instanceID, n.SesClock)
		go sender.Start()
		n.SenderList = append(n.SenderList, sender)
	}
}

func (n *Network) SafetyClose() {
	n.Socket.Close()
	fmt.Println("STOP")
	for _, sender := range n.SenderList {
		sender.Stop()
	}
	for _, receiver := range n.ReceiverList {
		receiver.Stop()
	}
}
