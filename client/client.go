package client

import (
	"fmt"
	"glorpn/types"
	"log"
	"net"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/milosgajdos/tenus"
	"github.com/net-byte/water"
)

const (
	BUFSIZE = 1500
)

type Client struct {
	WANIfaceName   string
	WANIp          string
	Iface          net.Interface
	TunSource      *gopacket.PacketSource
	Authenticated  bool
	WANIfaceHandle *pcap.Handle
	TunnelConn     net.Conn
}

func NewClient() types.Client {
	return &Client{
		Authenticated: false,
	}
}

func (c *Client) connectServer(ip string) error { // Called from 'client' to pubip
	conn, err := net.Dial("tcp", net.JoinHostPort(ip, "3000")) // Dial is remote, listen is local
	if err != nil {
		return err
	}

	c.TunnelConn = conn

	log.Printf("Dialed server at ip %s\n", ip)

	err = c.sendHello(conn)
	if err != nil {
		return err
	}

	err = c.awaitAck(conn)
	if err != nil {
		return err
	}

	c.Authenticated = true

	err = c.sendData(conn, "this is data")
	if err != nil {
		return err
	}
	log.Println("Send some data to server")

	err = c.handle(conn)
	if err != nil {
		return fmt.Errorf("failed to handle conn as the client")
	}

	// err = c.handleIncoming(conn) {
	// }

	return nil
}

func (c *Client) sendData(conn net.Conn, data string) error {
	dataPack := types.NewGlorpNPacket(0x07, []byte(data))
	_, err := conn.Write(dataPack.Serialize())
	return err
}

func (c *Client) awaitAck(conn net.Conn) error {
	buf := make([]byte, BUFSIZE)
	for {
		_, err := conn.Read(buf[:])
		if err != nil {
			return err
		}
		ackPack := types.NewGlorpNPacket(buf[0], buf[1:len(buf)-1])
		if ackPack.Header == 2 {
			log.Println("Got Ack from server")
			return nil
		} else {
			log.Printf("Did not get Ack from server, restart connection. %v\n", ackPack.Header)
			panic("Didnt get ack")
		}
	}
}

func (c *Client) sendHello(conn net.Conn) error {
	helloPack := types.NewGlorpNPacket(0x01, []byte("Hello"))
	_, err := conn.Write(helloPack.Serialize())
	if err != nil {
		return err
	}
	log.Println("Sent hello to server")
	return nil
}

func (c *Client) handleIncoming() error {
	// if c.isAuthenicated() && c.TunnelConn == nil {
	// 	var err error
	// 	*c.TunnelConn, err = net.Dial("tcp", net.JoinHostPort(c.WANIp, "3000"))
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	for packet := range c.TunSource.Packets() {
		networkLayer := packet.NetworkLayer()

		_, ok := networkLayer.(*layers.IPv4)
		if ok {
			fmt.Println("is ipv4")
		}

		fmt.Println(packet.String())

		if !c.isAuthenicated() {
			fmt.Println("Not authenicated, skipping...")
			continue
		}

		fmt.Println("Writing to tunnel conn")
		glorpPack := types.NewGlorpNPacket(0x07, packet.Data())
		_, err := c.TunnelConn.Write(glorpPack.Serialize())
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) isAuthenicated() bool {
	return c.Authenticated
}

func getIfaceIP(ifaceName string) (string, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", fmt.Errorf("failed to get interface %v by name: %w", ifaceName, err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("failed to get interface %v's addrs: %w", ifaceName, err)
	}

	if len(addrs) >= 1 {
		ip, _, err := net.ParseCIDR(addrs[0].String())
		if err != nil {
			return "", fmt.Errorf("failed to parse iface attached cidr: %w", err)
		}
		return ip.String(), nil
	}
	return "", fmt.Errorf("iface did not have any addrs")
}

// Listen on main wan iface for port 3000 traffic
// Called by 'server'
func (c *Client) serve(wanIfaceName string) error {
	ip, err := getIfaceIP(wanIfaceName)
	if err != nil {
		return err
	}

	c.WANIp = ip

	listener, err := net.Listen("tcp", net.JoinHostPort(ip, "3000"))
	if err != nil {
		return fmt.Errorf("failed to start listener on wan iface: %w", err)
	}

	// Create handle for main iface
	c.WANIfaceHandle, err = pcap.OpenLive(wanIfaceName, 1600, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("failed to create %v handle: %w", wanIfaceName, err)
	}

	for {
		fmt.Println("Listening...")
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("err while reading conn")
		}
		err = c.handle(conn)
		if err != nil {
			fmt.Printf("Error handling packet: %v\n", err)
		}
	}
}

func (c *Client) handle(conn net.Conn) error {
	// c.sendData(conn, "Sent from hadle before loop")
	fmt.Println("Connected to: ", conn.RemoteAddr())
	c.TunnelConn = conn

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		return fmt.Errorf("failed to create raw socket in handle: %w", err)
	}

	// han, err := pcap.OpenLive(c.WANIfaceName, 1600, false, time.Millisecond)
	// if err != nil {
	// 	return fmt.Errorf("failed to create %v handle: %w", "Eth0", err)
	// }
	buf := make([]byte, BUFSIZE)
	for {
		_, err := conn.Read(buf[:])
		if err != nil {
			return fmt.Errorf("err while reading from remote conn, closing conn and waiting again: %v", err)
		}
		packet := types.NewGlorpNPacket(buf[0], buf[1:len(buf)-1])
		if packet.Header == 1 {
			fmt.Println("Client Hello packet")
			err = c.sendAck(conn)
			if err != nil {
				fmt.Println("Ack returned errer, dont care")
			} else {
				fmt.Println("Ack returned fine. authenicating")
				c.Authenticated = true
			}
		} else if packet.Header == 7 {
			recvPack := gopacket.NewPacket(packet.Data, layers.LayerTypeIPv4, gopacket.Default)
			// fmt.Println("Recieved Packet: \n: ", p.String())

			resourcedBytes, err := c.resourcePack(recvPack)
			if err != nil {
				fmt.Println("Error resourcing packet, err: %w", err)
				continue
				// return fmt.Errorf("unable to re-source packet received from peer: %w", err)
			}

			pack := gopacket.NewPacket(resourcedBytes, layers.LayerTypeIPv4, gopacket.Default)

			ipv4Layer, ok := pack.NetworkLayer().(*layers.IPv4)
			if !ok {
				fmt.Println("Not ipv4, skipping")
				continue
			}

			dstAddr := &syscall.SockaddrInet4{
				// Port: 80,
				Addr: [4]byte(ipv4Layer.DstIP),
			}

			err = syscall.Sendto(fd, pack.Data(), 0, dstAddr)
			if err != nil {
				fmt.Println("failed to write data to wire: ", err)
				continue
			}

			// err = han.WritePacketData(pack.Data())
			// if err != nil {
			// 	fmt.Println("failure to write data to wire: ", err)
			// 	continue
			// 	// return fmt.Errorf("failure to write data to wire: %w", err)
			// }
		}
	}
}

func (c *Client) resourcePack(packet gopacket.Packet) ([]byte, error) {
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return nil, fmt.Errorf("packet does not have ipv4 layer")
	}

	ip, _ := ipLayer.(*layers.IPv4)

	newPacket := gopacket.NewSerializeBuffer()
	newSrcIP := net.IPv4(20, 0, 0, 10)
	newDstIP := net.IPv4(8, 8, 8, 8)

	ip.SrcIP = newSrcIP
	ip.DstIP = newDstIP
	ip.Checksum = 0

	var transportLayer gopacket.SerializableLayer
	switch packet.TransportLayer().LayerType() {
	case layers.LayerTypeTCP:
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer == nil {
			return nil, fmt.Errorf("tcp is empty")
		}
		tcp, _ := tcpLayer.(*layers.TCP)
		// Update checksums later
		tcp.SetNetworkLayerForChecksum(ip)
		transportLayer = tcp
	case layers.LayerTypeUDP:
		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer == nil {
			return nil, fmt.Errorf("udp is empty")
		}
		udp, _ := udpLayer.(*layers.UDP)
		// Update checksums later
		udp.SetNetworkLayerForChecksum(ip)
		transportLayer = udp
	default:
		return nil, fmt.Errorf("non-tcp non-udp transport layer")
	}

	options := gopacket.SerializeOptions{
		FixLengths: true,
		ComputeChecksums: true,
	}

	err := gopacket.SerializeLayers(newPacket, options, ip, transportLayer)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize layers: %w", err)
	}

	return newPacket.Bytes(), nil
}

func (c *Client) sendAck(conn net.Conn) error {
	data := []byte("")
	keyPacket := types.NewGlorpNPacket(0x02, data)
	_, err := conn.Write(keyPacket.Serialize())
	if err != nil {
		return err
	}
	log.Println("Sending ack to client")
	return nil
}

func (c *Client) createTun(cidr string) error {
	// Create TUN interface
	config := water.Config{
		DeviceType: water.TUN, // Use water.TAP for a TAP device
	}
	config.Name = "tun0" // Optional: Specify the interface name

	iface, err := water.New(config)
	if err != nil {
		log.Fatalf("Failed to create TUN interface: %v", err)
	}

	link, err := tenus.NewLinkFrom(config.Name)
	if err != nil {
		log.Fatalf("Failed to get link for interface %s: %v", config.Name, err)
	}

	ipAddr, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Fatalf("Failed to parse CIDR: %v", err)
	}

	// Assign the IP address
	err = link.SetLinkIp(ipAddr, ipNet)
	if err != nil {
		log.Fatalf("Failed to set IP address: %v", err)
	}

	err = link.SetLinkUp()
	if err != nil {
		log.Fatalf("Failed to set tun link up: %v", err)
	}

	fmt.Printf("Interface %s is up\n", iface.Name())

	handle, err := pcap.OpenLive("tun0", 1500, true, pcap.BlockForever)
	if err != nil {
		log.Fatalf("Faield to open live pcap: %v", err)
	}

	c.TunSource = gopacket.NewPacketSource(handle, handle.LinkType())

	return nil
}

func (c *Client) Start(wanIfaceName, peerIP string) error {
	// Create tun iface
	cidr := "20.0.0.1/24"
	err := c.createTun(cidr)
	if err != nil {
		return fmt.Errorf("failed to create tun iface at runtime, cidr: %v: %w", cidr, err)
	}

	c.WANIfaceName = wanIfaceName
	go c.handleIncoming()

	// Are you the server or client?

	// Client. Has pub ip
	if peerIP != "" {
		err = c.connectServer(peerIP)
		if err != nil {
			return err
		}
	} else { // Server
		err = c.serve(wanIfaceName)
		if err != nil {
			return err
		}
	}

	endch := make(chan int)

	<-endch

	return nil

	// go c.serve(wanIfaceName)
	// // if err != nil {
	// // 	return err
	// // }

	// // if peerIP == "" {
	// // 	fmt.Println("No peer, only listening")
	// // } else {
	// // 	err = c.connect(peerIP)
	// // 	if err != nil {
	// // 		return fmt.Errorf("failed to connect on peerip: %v: %w", peerIP, err)
	// // 	}
	// // }

	// err = c.connect(peerIP)
	// if err != nil {
	// 	return fmt.Errorf("failed to connect on peerip: %v: %w", peerIP, err)
	// }

	// err = c.handleIncoming()
	// if err != nil {
	// 	return fmt.Errorf("failed to handle incoming non-3000 traffic: %w", err)
	// }

}

// Process packets
// packet := make([]byte, 1500) // MTU size
// for {
// 	n, err := iface.Read(packet)
// 	if err != nil {
// 		log.Fatalf("Error reading packet: %v", err)
// 	}
// 	fmt.Printf("Received packet: %x\n", packet[:n])

// 	gopack := gopacket.NewPacket(packet[:n], layers.LayerTypeEthernet, gopacket.Default)
// 	networkLayer := gopack.NetworkLayer()

// 	_, ok := networkLayer.(*gopacket.IPv4)
// 	if ok {
// 		fmt.Println("is ipv4")
// 	} else {
// 		fmt.Println("is not")
// 	}

// 	glorpPack := types.NewGlorpNPacket(0x07, packet[:n])
// 	_, err = conn.Write(glorpPack.Serialize())
// 	if err != nil {
// 		return err
// 	}
// 	if !c.isAuthenicated() {
// 		fmt.Println("Not authenicated, skipping...")
// 		continue
// 	}
// 	// Process the packet
// 	// Write responses back to iface.Write(packet[:n]) if needed
// }

// handle, err := pcap.OpenLive("eth0", 1500, true, pcap.BlockForever)
// if err != nil {
// 	return err
// }

// packetSrc := gopacket.NewPacketSource(handle, handle.LinkType())
// for packet := range packetSrc.Packets() {
// 	fmt.Println("Got packet on dummy")
// 	glorpPack := types.NewGlorpNPacket(0x07, packet.Data())
// 	_, err := conn.Write(glorpPack.Serialize())
// 	if err != nil {
// 		return err
// 	}
// 	if !c.isAuthenicated() {
// 		continue
// 	}
// }

// listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP(ip), Port: 80})
// if err != nil {
// 	return err
// }
// //buf := make([]byte, 2048)
// fmt.Println("Made it here 2")
// for {
// 	_, err = listener.Accept()
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("Accepted")
// 	// fmt.Println("Got data: ", string(buf))
// 	if !c.isAuthenicated() {
// 		continue
// 	}
// }

// iface, err := water.New(water.Config{DeviceType: water.TUN})
// if err != nil {
// 	return err
// }

// fmt.Println("Iface name: ", iface.Name())

// link, err := tenus.NewLinkFrom(iface.Name())
// if err != nil {
// 	return err
// }

// err = link.SetLinkMTU(1300)
// if nil != err {
// 	log.Fatalln("Unable to set MTU to 1300 on interface")
// }

// lIp, lNet, err := net.ParseCIDR("10.11.0.1/24")
// if err != nil {
// 	return err
// }

// err = link.SetLinkIp(lIp, lNet)
// if err != nil {
// 	return err
// }

// err = link.SetLinkUp()
// if err != nil {
// 	return err
// }

// buf := make([]byte, 2048)
// for {
// 	_, err := net.Listen("tcp", "192.168.1.250:")
// 	//_, err := iface.Read(buf[:])
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Printf("Data: %s \n", string(buf))
// 	pack := types.NewGlorpNPacket(0x07, buf)
// 	conn.Write(pack.Serialize())
// }

// Every packet that hits tun0 gets sent out destined for port 3000

// conn, err := net.Dial("tcp", "10.0.0.1:3000")
// if err != nil {
// 	log.Fatalln("unable to dial")
// }
// // opcode := []byte(0x01)
// data := []byte("Hello")

// packet := types.NewGlorpNPacket(0x07, data)
// conn.Write(packet.Serialize())

// fmt.Println("Listening for key packet...")
// var buf [1024]byte
// for {
// 	_, err = conn.Read(buf[:])
// 	if err != nil {
// 		log.Fatalln("err reading key packet from server")
// 	}
// 	keyPacket := types.NewGlorpNPacket(buf[0], buf[1:len(buf)-1])
// 	fmt.Println("Key: ", string(keyPacket.Data))
// }
