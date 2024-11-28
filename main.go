package main

import (
	"fmt"
	"log"
	"net"
)

type Server struct {
	Key string
}

type Client struct {

}

type GlorpNPacket struct {
	Header byte
	Data   []byte
}

func NewGlorpNPacket(header byte, data []byte) *GlorpNPacket {
	return &GlorpNPacket{
		Header: header,
		Data: data,
	}
}

func (g *GlorpNPacket) Serialize() []byte {
	packetBytes := []byte{g.Header}
	packetBytes = append(packetBytes, g.Data...)
	return packetBytes
}

// Server listens on 3000

func main() {
	l, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatalln("couldn't listen on network")
	}

	for {
		fmt.Println("Listening...")
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("err while accept")
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	fmt.Println("Connected to: ", conn.RemoteAddr())
	for {
		var buf [1024]byte
		_, err := conn.Read(buf[:])
		if err != nil {
			log.Println("err while reading from remote conn")
			return
		}
		packet := NewGlorpNPacket(buf[0], buf[1:len(buf)-1])
		if packet.Header == 1 {
			fmt.Println("Client Hello packet")

		}
		fmt.Println(string(packet.Data), packet.Header)

		fmt.Println("Sending key...")
		sendKey(conn)
		fmt.Println("Sent key")
	}
}

func sendKey(conn net.Conn) error {
	data := []byte("1234")
	keyPacket := NewGlorpNPacket(0x02, data)
	_, err := conn.Write(keyPacket.Serialize())
	return err
}