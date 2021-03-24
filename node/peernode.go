package node

import (
	"fmt"
)

type PeerNode struct {
	IpAddress   string `json:"ip_address"`
	Port        uint64 `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`
	IsActive    bool   `json:"is_active"`
}

func (p PeerNode) SocketAddress() string {
	return fmt.Sprintf("%s:%d", p.IpAddress, p.Port)
}
