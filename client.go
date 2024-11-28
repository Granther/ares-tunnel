package main

// import (
// 	"fmt"
// 	"glorpn/types"
// 	"log"
// 	"net"
// )

// func main() {
// 	conn, err := net.Dial("tcp", "127.0.0.1:3000")
// 	if err != nil {
// 		log.Fatalln("unable to dial")
// 	}
// 	// opcode := []byte(0x01)
// 	data := []byte("Hello")

// 	packet := types.NewGlorpNPacket(0x01, data)
// 	conn.Write(packet.Serialize())

// 	fmt.Println("Listening for key packet...")
// 	var buf [1024]byte
// 	for {
// 		_, err = conn.Read(buf[:])
// 		if err != nil {
// 			log.Fatalln("err reading key packet from server")
// 		}
// 		keyPacket := types.NewGlorpNPacket(buf[0], buf[1:len(buf)-1])
// 		fmt.Println("Key: ", string(keyPacket.Data))
// 	}
// }
