package divert

import (
	"fmt"
	"net"

	"github.com/sunqibuhuake/go-divert/header"
)

type Packet struct {
	Raw       []byte
	Addr      *Address
	IpHdr     header.IPHeader
	ipVersion int
	parsed    bool
}

func (p *Packet) verifyParsed() {
	if !p.parsed {
		p.parseHeaders()
	}
}

func (p *Packet) parseHeaders() {
	p.ipVersion = int(p.Raw[0] >> 4)

	if p.ipVersion == 4 {
		p.IpHdr = header.NewIPv4Header(p.Raw)
	} else {
		p.IpHdr = header.NewIPv6Header(p.Raw)
	}

	p.parsed = true
}

func (p *Packet) String() string {
	p.verifyParsed()

	return fmt.Sprintf("Packet {\n"+
		"\tIPHeader=%v\n"+
		"\tWinDivertAddr=%v\n"+
		"\tRawData=%v\n"+
		"}",
		p.IpHdr, p.Addr, p.Raw)
}

// Shortcut for ipHdr.Version()
func (p *Packet) IpVersion() int {
	return p.ipVersion
}

// Shortcut for IpHdr.SrcIP()
func (p *Packet) SrcIP() net.IP {
	p.verifyParsed()

	return p.IpHdr.SrcIP()
}

// Shortcut for IpHdr.SetSrcIP()
func (p *Packet) SetSrcIP(ip net.IP) {
	p.verifyParsed()

	p.IpHdr.SetSrcIP(ip)
}

// Shortcut for IpHdr.DstIP()
func (p *Packet) DstIP() net.IP {
	p.verifyParsed()

	return p.IpHdr.DstIP()
}

// Shortcut for IpHdr.SetDstIP()
func (p *Packet) SetDstIP(ip net.IP) {
	p.verifyParsed()

	p.IpHdr.SetDstIP(ip)
}
