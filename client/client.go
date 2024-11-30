package client

import (
	"fmt"
	"glorpn/types"
	"log"
	"net"
	"os"
	"syscall"
	"unsafe"
)

type Client struct {
	PublicIP net.IP
	Iface    net.Interface
	Key      string
}

func NewClient() types.Client {
	return &Client{}
}

func getTun(ifaceName string) (*os.File, error) {
	tun, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	var req struct {
		Name  [16]byte
		Flags uint16
	}

	req.Flags = syscall.IFF_TUN | syscall.IFF_NO_PI // TUN mode, no extra packet info
	copy(req.Name[:], ifaceName)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, tun.Fd(), uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&req)))
	if errno != 0 {
		return nil, fmt.Errorf("failed to create TUN interface: %v", errno)
	}

	return tun, nil
}

func monitorExiting(tun *os.File) {
	buf := make([]byte, 1024)
	for {
		_, err := tun.Read(buf[:])
		if err != nil {
			log.Fatalf("Error in monitor exiting: %w \n", err)
		}
		fmt.Println("Data", string(buf))
	}
}

func (c *Client) Start() error {
	file, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		log.Fatalf("error os.Open(): %v\n", err)
	}

	ifr := make([]byte, 18)
	copy(ifr, []byte("tun0"))
	ifr[16] = 0x01
	ifr[17] = 0x10
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(file.Fd()),
		uintptr(0x400454ca), uintptr(unsafe.Pointer(&ifr[0])))
	if errno != 0 {
		log.Fatalf("error syscall.Ioctl(): %v\n")
	}

	// cmd, err := exec.Run("/sbin/ifconfig",
	// 	[]string{"ifconfig", "tun0", "192.168.7.1",
	// 		"pointopoint", "192.168.7.2", "up"},
	// 	nil, ".", 0, 1, 2)
	// if err != nil {
	// 	log.Fatalf("error exec.Run(): %v\n", err)
	// }
	// cmd.Wait(0)

	for {
		buf := make([]byte, 2048)
		read, err := file.Read(buf)
		if err != nil {
			log.Fatalf("error os.Read(): %v\n", err)
		}

		for i := 0; i < 4; i++ {
			buf[i+12], buf[i+16] = buf[i+16], buf[i+12]
		}
		buf[20] = 0
		buf[22] = 0
		buf[23] = 0
		var checksum uint16
		for i := 20; i < read; i += 2 {
			checksum += uint16(buf[i])<<8 + uint16(buf[i+1])
		}
		checksum = ^(checksum + 4)
		buf[22] = byte(checksum >> 8)
		buf[23] = byte(checksum & ((1 << 8) - 1))

		_, err = file.Write(buf)
		if err != nil {
			log.Fatalf("error os.Write(): %v\n", err)
		}
		fmt.Println("Got")
	}

	// tun, err := getTun("tun0")
	// if err != nil {
	// 	return err
	// }

	// monitorExiting(tun)

	// return nil

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
