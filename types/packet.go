package types

type GlorpNPacket struct {
	Header byte
	Data   []byte
}

func NewGlorpNPacket(header byte, data []byte) *GlorpNPacket {
	return &GlorpNPacket{
		Header: header,
		Data: data,
	}
}

func (g *GlorpNPacket) Serialize() []byte {
	packetBytes := []byte{g.Header}
	packetBytes = append(packetBytes, g.Data...)
	return packetBytes
}
