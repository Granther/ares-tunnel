package client

import (
	"glorpn/types"
	"net"
)

type Client struct {
	PublicIP net.IP
	Iface    net.Interface
	Key      string
}

func NewClient() types.Client {
	return &Client{}
}

func (c *Client) Start() error {
	return nil
} 