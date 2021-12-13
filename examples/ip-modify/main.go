package main

import (
	"fmt"
	"github.com/sunqibuhuake/go-divert"
	"net"
)

func checkPacket(handle *divert.Handle, packetChan <-chan *divert.Packet) {
	for packet := range packetChan {

		packet.String()
		if packet.Addr.IsOutbound() {
			fmt.Println("OUT")
			packet.SetDstIP(net.ParseIP("43.154.110.254"))
		} else {
			fmt.Println("IN")
			packet.SetSrcIP(net.ParseIP("192.168.1.104"))
		}

		handle.HelperCalcChecksum(packet)
		handle.Send(packet.Raw, packet.Addr)
	}
}

func main() {
	var filter = "(outbound  and tcp.DstPort == 7777) or (inbound  and  tcp.SrcPort == 7777)"
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
	defer handle.Close()
	checkPacket(handle, packetChan)

}
