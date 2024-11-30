package client

import (
	"fmt"
	"glorpn/types"
	"log"
	"net"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

type Client struct {
	PublicIP      net.IP
	Iface         net.Interface
	Authenticated bool
	Key           string
}

func NewClient() types.Client {
	return &Client{
		Authenticated: false,
	}
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

func (c *Client) handleIncoming(ip string) error {
	for {
		_, err := net.Listen("tcp", ip)
		if err != nil {
			return err
		}
		if !c.isAuthenicated() {
			continue
		}
	}
}

func (c *Client) isAuthenicated() bool {
	return c.Authenticated
}

func (c *Client) Start() error {
	// connect to server pub ip
	serverIP := "192.168.1.250"
	clientIP := "192.168.1.250"

	c.connect(serverIP)

	c.handleIncoming(clientIP)

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

	return nil

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
