package types

type Client interface {
	Start() error 
}

type Server interface {
	Start() error 
}
