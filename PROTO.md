# Custom VPN Protocol
- TCP handshake
- Client connects to server with hello
- Server generate and sends key which will encrypt all traffic

### OpCodes
- Client Hello
- Server Key Send
- Client Ack
- Client Discon
- Server Discon
- Data

### Backend
- Creates interface on both systems
- Client knows its public IP addr and the server's
- Client sends connection req to server and the protocol does its thing
- All data sent to the client's local iface/ip will be wrapped by the encryption of the vpn and stamped for delivery at the server's IP via the client's actual logical network
    - Any data that hits the slected interface

#### Right now
- Any data exiting with a src of the selected iface gets wrapped
- Any data destined for the tun0 iface gets unwrapped 

- 18.0.0.0/24 is the vpn network
- 

### TUN
sudo ip tuntap add dev tun0 mode tun
sudo ip addr add 10.0.0.1/24 dev tun0
sudo ip link set up dev tun0

sudo ip tuntap add dev tun0 mode tun
sudo ip addr add 10.10.0.1/24 dev tun1
sudo ip link set up dev tun1

sudo ip addr add 10.9.0.1/24 dev tap0
sudo ip link set up dev tap0

sudo ip link add dev dummy0 type dummy
sudo ip link set dummy0 up
sudo ip addr add 18.0.0.1/24 dev dummy0

### 2 Clients
- Client 1 (server), listens without peer ip
    - When connection is made on port 3000, 