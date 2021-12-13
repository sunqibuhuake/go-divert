package header

import "net"

type IPHeader interface {
	String() string

	Version() int
	HeaderLen() uint8
	NextHeader() uint8
	SrcIP() net.IP
	DstIP() net.IP
	SetSrcIP(net.IP)
	SetDstIP(net.IP)
	Checksum() (uint16, error)
	NeedNewChecksum() bool
}

func ProtocolName(protocol uint8) string {
	switch protocol {
	case ICMPv4:
		return "ICMPv4"
	case TCP:
		return "TCP"
	case UDP:
		return "UDP"
	case ICMPv6:
		return "ICMPv6"
	default:
		return "Unimplemented Protocol"
	}
}
