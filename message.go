package ssdprecv

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func (self *SsdpReceiver) parseRequest(message string) (shouldRespond bool, st string) {
	split := strings.Split(message, "\r\n")
	if len(split) == 0 {
		return false, ""
	}
	if !strings.EqualFold(split[0], "M-SEARCH * HTTP/1.1") {
		return false, ""
	}

	headers := map[string]string{}
	for _, line := range split {
		sline := strings.Split(line, ": ")
		if len(sline) != 2 {
			continue
		}
		headers[sline[0]] = sline[1]
	}

	for key, val := range headers {
		if !strings.EqualFold(key, "ST") {
			continue
		}
		if self.isAcceptableSt(val) {
			return true, val
		}
	}

	return false, ""
}

func (self *SsdpReceiver) createAliveAdvertisement(ip *net.IPAddr, nt string) string {
	var buffer bytes.Buffer

	buffer.WriteString("NOTIFY * HTTP/1.1\r\n")
	buffer.WriteString("HOST: 239.255.255.250:1900\r\n")
	buffer.WriteString("CACHE-CONTROL: max-age=" + strconv.Itoa(int(self.AdvInterval.Seconds()*4)) + "\r\n")
	buffer.WriteString("LOCATION: http://" + ip.String() + ":" + strconv.Itoa(self.http.Port) + self.http.Path + "\r\n")
	buffer.WriteString("NT: " + nt + "\r\n")
	buffer.WriteString("NTS: ssdp:alive\r\n")
	buffer.WriteString("SERVER: Golang/1.4.1 UPnP/1.1 Golang/1.4.1\r\n")
	buffer.WriteString("USN: " + self.http.Uuid + ":" + nt + "\r\n")
	buffer.WriteString("\r\n")

	message := buffer.String()

	fmt.Println("alive message created")
	fmt.Println(message)

	return message
}

func (self *SsdpReceiver) createByebyeAdvertisement(nt string) string {
	var buffer bytes.Buffer

	buffer.WriteString("NOTIFY * HTTP/1.1\r\n")
	buffer.WriteString("HOST: 239.255.255.250:1900\r\n")
	buffer.WriteString("NT: " + nt + "\r\n")
	buffer.WriteString("NTS: ssdp:byebye\r\n")
	buffer.WriteString("USN: " + self.http.Uuid + ":" + nt + "\r\n")
	buffer.WriteString("\r\n")

	message := buffer.String()

	fmt.Println("byebye message created")
	fmt.Println(message)

	return message
}

func (self *SsdpReceiver) createResponse(ip *net.IPAddr, st string) string {
	var buffer bytes.Buffer

	buffer.WriteString("HTTP/1.1 200 OK\r\n")
	buffer.WriteString("CACHE-CONTROL: max-age=" + strconv.Itoa(int(self.AdvInterval.Seconds()*4)) + "\r\n")
	buffer.WriteString("EXT:  \r\n")
	buffer.WriteString("LOCATION: http://" + ip.String() + ":" + strconv.Itoa(self.http.Port) + self.http.Path + "\r\n")
	buffer.WriteString("SERVER: Golang/1.4.1 UPnP/1.1 Golang/1.4.1\r\n")
	buffer.WriteString("ST: ")
	switch st {
	case self.http.Uuid:
		buffer.WriteString(self.http.Uuid)
	default:
		buffer.WriteString(st)
	}
	buffer.WriteString("\r\n")
	buffer.WriteString("USN: " + self.http.Uuid)
	switch st {
	case self.http.Uuid:
		break
	default:
		buffer.WriteString(":" + st)
	}
	buffer.WriteString("\r\n")
	buffer.WriteString("\r\n")

	message := buffer.String()

	fmt.Println("Response message created")
	fmt.Println(message)

	return message
}

func (self *SsdpReceiver) isAcceptableSt(st string) bool {
	if st == "ssdp:all" {
		return true
	}
	if st == self.http.Uuid {
		return true
	}
	for _, acceptable := range self.http.ST {
		if st == acceptable {
			return true
		}
	}
	return false
}
