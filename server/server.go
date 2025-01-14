package server

import (
	"fmt"
	"glorpn/types"
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type Server struct {
	PublicIP        net.IP
	Iface           net.Interface
	MainIfaceHandle *pcap.Handle
	Key             string
}

func NewServer() types.Server {
	return &Server{}
}

func (s *Server) Start() error {
	hostIP := "18.0.0.1"
	listener, err := net.Listen("tcp", net.JoinHostPort(hostIP, "3000"))
	if err != nil {
		return err
	}

	// Create handle for main iface
	s.MainIfaceHandle, err = pcap.OpenLive("eth0", 1600, true, pcap.BlockForever)
	if err != nil {
		log.Fatalf("Failed to create eth0 handle: %v", err)
	}

	for {
		fmt.Println("Listening...")
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("err while accept")
		}
		err = s.handle(conn)
		if err != nil {
			fmt.Printf("Error handling packet: %v\n", err)
		}
	}
}

// What do I want?
// Server: Exit node, client's traffic is routed throuhg here

func (s *Server) handle(conn net.Conn) error {
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
			sendAck(conn)
		} else if packet.Header == 7 {
			// fmt.Println("Data Packet, data: ", string(packet.Data))
			// newPack, err := s.resourcePacket(packet.Data)
			// if err != nil {
			// 	fmt.Println("Got error while resourcing packet, ignoring: ", err)
			// 	continue
			// }

			err = s.MainIfaceHandle.WritePacketData(packet.Data)
			if err != nil {
				return err
			}
		}
	}
}

func (s *Server) resourcePacket(data []byte) ([]byte, error) {
	packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)

	// fmt.Println(packet.String())

	return packet.Data(), nil

	// for i, lay := range packet.Layers() {

	// }

	// ipLayer := packet.Layer(layers.LayerTypeIPv4)
	// icmpLayer := packet.Layer(layers.LayerTypeICMPv4)

	// // payloadLayer := packet.Layer(layers.LayerTypePayload)

	// if ipLayer == nil {
	// 	return nil, fmt.Errorf("eth or ip layer was nil")
	// }

	// // eth := ethLayer.(*layers.Ethernet)
	// ip := ipLayer.(*layers.IPv4)
	// icmp := icmpLayer.(*layers.ICMPv4)

	// // newEth := *eth
	// oldIP := *ip
	// oldIcmp := *icmp

	// newIP := layers.IPv4{
	// 	Version:    oldIP.Version,
	// 	IHL:        oldIP.IHL,
	// 	TOS:        oldIP.TOS,
	// 	Length:     oldIP.Length,
	// 	Id:         oldIP.Id,
	// 	Flags:      oldIP.Flags,
	// 	FragOffset: oldIP.FragOffset,
	// 	TTL:        oldIP.TTL,
	// 	Protocol:   oldIP.Protocol,
	// 	Checksum:   oldIP.Checksum,
	// 	SrcIP:      net.IP{192, 168, 1, 14},
	// 	DstIP:      net.IP{8, 8, 8, 8},
	// 	Options:    oldIP.Options,
	// 	Padding:    oldIP.Padding,
	// }

	// newIcmp := layers.ICMPv4{
	// 	TypeCode: oldIcmp.TypeCode,
	// 	Checksum: oldIcmp.Checksum,
	// 	Id:       oldIcmp.Id,
	// 	Seq:      oldIcmp.Seq,
	// }

	// newIcmp.Payload = oldIcmp.Payload

	// l, ok := packet.ApplicationLayer().(*layers.Layer)

	// // pay := layers.ICMPv4

	// // newIP.SrcIP = net.IP{192, 168, 1, 14}
	// // newIP.DstIP = net.IP{8, 8, 8, 8}

	// buffer := gopacket.NewSerializeBuffer()
	// opts := gopacket.SerializeOptions{
	// 	FixLengths:       true,
	// 	ComputeChecksums: true,
	// }

	// err := gopacket.SerializeLayers(buffer, opts, &newIP, icmp)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to serialize layers")
	// }

	// return buffer.Bytes(), nil
}

// Whan data arrives at server, fix packet to be from src of server vpn IP and dst of main iface

// Basic tunnel
// Send data down it across the internet

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
