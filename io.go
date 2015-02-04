package ssdprecv

import (
	"errors"
	"fmt"
	"net"
)

func (self *SsdpReceiver) advertise(subtype string) {
	switch subtype {
	case "alive":
	case "byebye":
		break
	default:
		return
	}

	ip, err := self.getLocalIpAddress()
	if err != nil {
		fmt.Println("Failed to detect valid local address")
		self.running = false
		return
	}

	conn, err := net.DialUDP("udp4", nil, mcastAddr)
	if err != nil {
		fmt.Println("Failed to create connection for advertisement")
		return
	}

	defer conn.Close()

	for _, nt := range self.http.ST {
		switch subtype {
		case "alive":
			_, err = conn.Write([]byte(self.createAliveAdvertisement(ip, nt)))
		case "byebye":
			_, err = conn.Write([]byte(self.createByebyeAdvertisement(nt)))
		}
		if err != nil {
			fmt.Println("Failed send advertisement")
		} else {
			fmt.Println("Sent advertisement as " + nt)
		}
	}
}

func (self *SsdpReceiver) listen(conn *net.UDPConn, ch chan int) {
	defer func() {
		fmt.Println("Closing listening connection")
		self.advertise("byebye")
		conn.Close()
		ch <- 1
	}()

	fmt.Println("Server listening process started")

	for self.running {
		n, addr, err := conn.ReadFromUDP(self.buff)
		if err != nil {
			fmt.Println("Failed to read message from connection")
			self.running = false
		}
		message := string(self.buff[:n])

		shouldRespond, st := self.parseRequest(message)

		if shouldRespond {
			self.respond(conn, addr, st)
		}
	}
}

func (self *SsdpReceiver) respond(conn *net.UDPConn, raddr *net.UDPAddr, st string) {
	ip, err := self.getLocalIpAddress()
	if err != nil {
		fmt.Println("Failed to detect valid local address")
		self.running = false
		return
	}

	if st == "ssdp:all" {
		for _, val := range self.http.ST {
			fmt.Println("Respond as " + val)
			_, err = conn.WriteToUDP([]byte(self.createResponse(ip, val)), raddr)
			if err != nil {
				fmt.Println("Failed to respond")
			}
		}
	} else {
		fmt.Println("Respond as " + st)
		_, err = conn.WriteToUDP([]byte(self.createResponse(ip, st)), raddr)
		if err != nil {
			fmt.Println("Failed to respond")
		}
	}
}

func (self *SsdpReceiver) getLocalIpAddress() (*net.IPAddr, error) {
	addrs, err := self.http.NIF.Addrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if addr.String() != "0.0.0.0" {
			ip, err := net.ResolveIPAddr("ip4", addr.String())
			if err != nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("No valid IP address")
}
