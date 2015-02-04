package ssdprecv

import (
	"errors"
	"fmt"
	"net"
	"time"
)

var mcastAddr, _ = net.ResolveUDPAddr("udp4", "239.255.255.250:1900")

type HttpServer struct {
	NIF  *net.Interface // Interface to listen
	Port int            // Port of the HTTP server to provide description file
	Path string         // Path of the description file from root
	Uuid string         // starts from "uuid:"
	ST   []string       // List of ST to respond
}

type SsdpReceiver struct {
	http        *HttpServer
	running     bool
	buff        []byte
	AdvInterval time.Duration // Interval for advertisement max-age will be by 4 of this value
}

func New(server *HttpServer) *SsdpReceiver {
	self := new(SsdpReceiver)

	self.http = server
	self.running = false
	self.buff = make([]byte, 4096)

	self.AdvInterval = 900 * time.Second

	return self
}

func (self *SsdpReceiver) Boot(ch chan int) error {
	if self.running {
		fmt.Println("Already running")
		return errors.New("running")
	}

	conn, err := net.ListenMulticastUDP("udp4", self.http.NIF, mcastAddr)
	if err != nil {
		fmt.Println("Failed to join multi-cast address")
		return err
	}

	self.running = true

	go self.listen(conn, ch)
	go self.loopAliveAdvertisement()

	fmt.Println("Successfully booted")
	return nil
}

func (self *SsdpReceiver) loopAliveAdvertisement() {
	time.Sleep(500 * time.Millisecond)
	self.advertise("alive")
	time.Sleep(100 * time.Millisecond)
	self.advertise("alive")
	time.Sleep(200 * time.Millisecond)
	self.advertise("alive")

	timer := time.Tick(self.AdvInterval)
	for _ = range timer {
		if !self.running {
			break
		}
		self.advertise("alive")
	}
}

func (self *SsdpReceiver) Shutdown() {
	fmt.Println("Shutting down")
	self.running = false
}
