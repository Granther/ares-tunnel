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
	iface, err := net.InterfaceByName("tun0")
	if err != nil {
		return err
	}
	fmt.Println("Got tun interface")

	// Server pub ip -> routes to servers tun iface:3000 -> Listen on this, if is data packet -> send to main iface 

	// listen on tun for traffic, 
	// incoming gets unwrapped and redirected to target iface
	// start goroutine lisening for incoming
	addrs, err := iface.Addrs()
	if err != nil {
		return err
	}
	if len(addrs) < 1 {
		return fmt.Errorf("tun interface doesn't have an IP")
	}

	listener, err := net.Listen(addrs[0].String(), "127.0.0.1:3000")
	if err != nil {
		return err
	}
	// l, err := net.Listen("tcp", "localhost:3000")
	// if err != nil {
	// 	log.Fatalln("couldn't listen on network")
	// }

	for {
		fmt.Println("Listening...")
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("err while accept")
		}
		go handle(conn)
	}
}

func attachTun(ifaceName string) error {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return err
	}
	fmt.Println("Got tun interface")

	// Server pub ip -> routes to servers tun iface:3000 -> Listen on this, if is data packet -> send to main iface 

	// listen on tun for traffic, 
	// incoming gets unwrapped and redirected to target iface
	// start goroutine lisening for incoming
	addrs, err := iface.Addrs()
	if err != nil {
		return err
	}
	if len(addrs) < 1 {
		return fmt.Errorf("tun interface doesn't have an IP")
	}

	ip, _, _ := net.ParseCIDR(addrs[0])

	listener, err := net.Listen(ip.String(), "127.0.0.1:3000")
	if err != nil {
		return err
	}

	for {
		_, err := listener.Accept()
		if err != nil {
			return err
		}
	}
	// outgoing gets wrapped and send out target iface
	return nil
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
		packet := types.NewGlorpNPacket(buf[0], buf[1:len(buf)-1])
		if packet.Header == 1 {
			fmt.Println("Client Hello packet")
		} else if packet.Header == 7 {
			fmt.Println("Data Packet")
		}
		fmt.Println(string(packet.Data), packet.Header)

		fmt.Println("Sending key...")
		sendKey(conn)
		fmt.Println("Sent key")
	}
}

func sendKey(conn net.Conn) error {
	data := []byte("1234")
	keyPacket := types.NewGlorpNPacket(0x02, data)
	_, err := conn.Write(keyPacket.Serialize())
	return err
}
