package node

import "github.com/kparkins/yarbit/database"

type Config struct {
	DataDir      string
	IpAddress    string
	Port         uint64
	Protocol     string
	Bootstrap    PeerNode
	MinerAccount database.Account
}
