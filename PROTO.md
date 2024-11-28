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