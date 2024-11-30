package server

import (
	"fmt"
	"glorpn/types"
	"log"
	"net"
)

type Server struct {
	PublicIP net.IP
	Iface    net.Interface
	Key      string
}

func NewServer() types.Server {
	return &Server{}
}

func (s *Server) Start() error {
	// Listen on main interface port 3000, if we get a data packet, send to main eth
	// ip, err := getIfaceIP("tun0")
	// if err != nil {
	// 	return err
	// }

	// listener, err := net.Listen("tcp", net.JoinHostPort(ip.String(), "3000"))
	// if err != nil {
	// 	return err
	// }
	hostIP := "192.168.1.250"
	listener, err := net.Listen("tcp", net.JoinHostPort(hostIP, "3000"))
	if err != nil {
		return err
	}

	for {
		fmt.Println("Listening...")
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("err while accept")
		}
		err = handle(conn)
		if err != nil {
			fmt.Printf("Error handling packet: %v\n", err)
		}
	}
}

func getIfaceIP(ifaceName string) (net.IP, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, err
	}
	// Server pub ip -> routes to servers tun iface:3000 -> Listen on this, if is data packet -> send to main iface 

	// listen on tun for traffic, 
	// incoming gets unwrapped and redirected to target iface
	// start goroutine lisening for incoming
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	if len(addrs) < 1 {
		return nil, fmt.Errorf("tun interface doesn't have an IP")
	}

	ip, _, err := net.ParseCIDR(addrs[0].String())
	if err != nil {
		return nil, err
	}
	return ip, nil
}

// What do I want?
// Server: Exit node, client's traffic is routed throuhg here

func handle(conn net.Conn) error {
	fmt.Println("Connected to: ", conn.RemoteAddr())
	buf := make([]byte, 2048)
	for {
		_, err := conn.Read(buf[:])
		if err != nil {
			return fmt.Errorf("err while reading from remote conn, closing conn and waiting again: %w\n", err)
		}
		packet := types.NewGlorpNPacket(buf[0], buf[1:len(buf)-1])
		if packet.Header == 1 {
			fmt.Println("Client Hello packet")	
			sendAck(conn)
		} else if packet.Header == 7 {
			fmt.Println("Data Packet, data: ", string(packet.Data))
			// sendMain(packet.Data)
		}
	}
}

// Basic tunnel
// Send data down it across the internet
// 

func sendAck(conn net.Conn) error {
	data := []byte("")
	keyPacket := types.NewGlorpNPacket(0x02, data)
	_, err := conn.Write(keyPacket.Serialize())
	if err != nil {
		return err
	}
	log.Println("Sending ack to client")
	return nil
}

func sendMain(data []byte) error {
	// ip, err := getIfaceIP("eth0")
	// if err != nil {
	// 	return err
	// }

	conn, err := net.Dial("tcp", "127.0.0.1:80")
	if err != nil {
		return err
	}

	conn.Write(data)

	// keyPacket := types.NewGlorpNPacket(0x07, data)
	// _, err := conn.Write(keyPacket.Serialize())
	return nil
}
