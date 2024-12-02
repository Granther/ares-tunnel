package client

import (
	"fmt"
	"glorpn/types"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/milosgajdos/tenus"
	"github.com/net-byte/water"
)

type Client struct {
	PublicIP       net.IP
	Iface          net.Interface
	Authenticated  bool
	WANIfaceHandle *pcap.Handle
}

func NewClient() types.Client {
	return &Client{
		Authenticated: false,
	}
}

func monitorExiting(tun *os.File) {
	buf := make([]byte, 1024)
	for {
		_, err := (*tun).Read(buf[:])
		if err != nil {
			log.Fatalf("Error in monitor exiting: %w \n", err)
		}
		fmt.Println("Data", string(buf))
	}
}

func execIP(command []string) {
	exec.Command("ip", command...)
}

func (c *Client) connect(ip string) error {
	conn, err := net.Dial("tcp", net.JoinHostPort(ip, "3000"))
	if err != nil {
		return err
	}
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

	err = c.handleIncoming(conn, ip)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) sendData(conn net.Conn, data string) error {
	dataPack := types.NewGlorpNPacket(0x07, []byte(data))
	_, err := conn.Write(dataPack.Serialize())
	return err
}

func (c *Client) awaitAck(conn net.Conn) error {
	buf := make([]byte, 2048)
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

func (c *Client) handleIncoming(conn net.Conn, ip string) error {
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

	ipAddr, ipNet, err := net.ParseCIDR("20.0.0.1/24")
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

	packets := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packets.Packets() {
		networkLayer := packet.NetworkLayer()

		ipv4Layers, ok := networkLayer.(*layers.IPv4)
		if ok {
			fmt.Println(ipv4Layers.SrcIP.String())
			fmt.Println("is ipv4")
		}

		fmt.Println(packet.String())

		glorpPack := types.NewGlorpNPacket(0x07, packet.Data())
		_, err = conn.Write(glorpPack.Serialize())
		if err != nil {
			return err
		}
		if !c.isAuthenicated() {
			fmt.Println("Not authenicated, skipping...")
			continue
		}
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
		return addrs[0].String(), nil
	}
	return "", fmt.Errorf("iface did not have any addrs")
}

func (c *Client) serve(wanIfaceName string) error {
	ip, err := getIfaceIP(wanIfaceName)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(ip, "3000"))
	if err != nil {
		return err
	}

	// Create handle for main iface
	c.WANIfaceHandle, err = pcap.OpenLive(wanIfaceName, 1600, true, pcap.BlockForever)
	if err != nil {
		fmt.Errorf("failed to create %v handle: %w", wanIfaceName, err)
	}

	for {
		fmt.Println("Listening...")
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("err while accept")
		}
		err = c.handle(conn)
		if err != nil {
			fmt.Printf("Error handling packet: %v\n", err)
		}
	}
}

func (c *Client) handle(conn net.Conn) error {
	fmt.Println("Connected to: ", conn.RemoteAddr())
	buf := make([]byte, 2048)
	for {
		_, err := conn.Read(buf[:])
		if err != nil {
			return fmt.Errorf("err while reading from remote conn, closing conn and waiting again: %v", err)
		}
		packet := types.NewGlorpNPacket(buf[0], buf[1:len(buf)-1])
		if packet.Header == 1 {
			fmt.Println("Client Hello packet")
			c.sendAck(conn)
		} else if packet.Header == 7 {
			err = c.WANIfaceHandle.WritePacketData(packet.Data)
			if err != nil {
				return err
			}
		}
	}
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

func (c *Client) Start(wanIfaceName, peerIP string) error {
	// connect to server pub ip

	go c.serve(wanIfaceName)

	if peerIP == "" {
		fmt.Println("No peer, only listening")
		for {
		}
	}

	err := c.connect(peerIP)
	if err != nil {
		return fmt.Errorf("failed to connect on peerip: %v: %w", peerIP, err)
	}

	return nil

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
}
