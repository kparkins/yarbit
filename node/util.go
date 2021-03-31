package node

func FilterPeers(strings map[string]PeerNode, match func(s string) bool) map[string]PeerNode {
	out := make(map[string]PeerNode)
	for k, v := range strings {
		if match(k) {
			out[k] = v
		}
	}
	return out
}
