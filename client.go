package main

import (
	"fmt"
	"log"
	"net"
)

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

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:3000")
	if err != nil {
		log.Fatalln("unable to dial")
	}
	// opcode := []byte(0x01)
	data := []byte("Hello")

	packet := NewGlorpNPacket(0x01, data)
	conn.Write(packet.Serialize())

	fmt.Println("Listening for key packet...")
	var buf [1024]byte
	for {
		_, err = conn.Read(buf[:])
		if err != nil {
			log.Fatalln("err reading key packet from server")
		}
		keyPacket := NewGlorpNPacket(buf[0], buf[1:len(buf)-1])
		fmt.Println("Key: ", string(keyPacket.Data))
	}
}

