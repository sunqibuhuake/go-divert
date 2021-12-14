package main

import (
	"bytes"
	"fmt"
	"github.com/sunqibuhuake/go-divert"
)

func checkPacket(handle *divert.Handle, packetChan <-chan *divert.Packet) {
	for packet := range packetChan {

		var target = []byte("lilk")
		var hack = []byte("hihi")

		if packet.Addr.IsOutbound() {
			fmt.Println("OUT")
			if bytes.Contains(packet.Raw, target) {
				fmt.Println("Origin:")
				fmt.Println(string(packet.Raw))
				packet.Raw = bytes.ReplaceAll(packet.Raw, target, hack)
				fmt.Println("Modified:")
				fmt.Println(string(packet.Raw))
				divert.HelperCalcChecksum(packet,0)
			}
		} else {
			fmt.Println("IN")
			//packet.Raw = bytes.ReplaceAll(packet.Raw, hack, target)
			//handle.HelperCalcChecksum(packet)
			fmt.Println(string(packet.Raw))
		}
		handle.Send(packet.Raw, packet.Addr)
	}
}

func main() {
	var filter = "(outbound and tcp.PayloadLength > 0 and tcp.DstPort == 1800) or (inbound and  tcp.PayloadLength > 0 and  tcp.SrcPort == 1800)"
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
