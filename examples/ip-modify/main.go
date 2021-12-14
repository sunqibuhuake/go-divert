package main

import (
	"fmt"
	"github.com/sunqibuhuake/go-divert"
	"net"
)


var hasOutPacket = false
func checkPacket(handle *divert.Handle, packetChan <-chan *divert.Packet) {
	var originalDstIp = ""
	var proxyHostIp = net.ParseIP("43.154.110.254")
	for packet := range packetChan {
		packet.String()
		if packet.Addr.IsOutbound() {
			hasOutPacket = true
			originalDstIp = packet.DstIP().To4().String()
			packet.SetDstIP(proxyHostIp)
		} else {
			if hasOutPacket != true {
				continue
			}
			if len(originalDstIp) > 0 {
				packet.SetSrcIP(net.ParseIP(originalDstIp))
			}
		}
		divert.HelperCalcChecksum(packet, 0)
		handle.Send(packet.Raw, packet.Addr)
	}
}

func main() {
	var filter = "(outbound  and tcp.DstPort == 1800) or (inbound  and  tcp.SrcPort == 1800)"
	handle, err := divert.Open(filter, divert.LayerNetwork, divert.PriorityLowest, 0)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	handle.SetParam(divert.QueueLength, divert.QueueLengthMax)
	handle.SetParam(divert.QueueTime, divert.QueueTimeMax)


	packetChan, err := handle.Packets()
	if err != nil {
		panic(err)
	}
	//defer handle.Close()
	 checkPacket(handle, packetChan)

}
