package types

type Client interface {
	Start(wanIface, peerIP string) error 
}

type Server interface {
	Start() error 
}
