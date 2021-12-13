package divert

import (
	"fmt"
	"golang.org/x/sys/windows"
	"unsafe"
)

func (h *Handle) recvLoop(packetChan chan<- *Packet) {
	for h.Open == HandleOpen {
		addr := Address{}
		buff := make([]byte, PacketBufferSize)

		_, err := h.Recv(buff, &addr)
		if err != nil {
			fmt.Println("Recv loop error: " +err.Error())
			close(packetChan)
			break
		}

		packet := &Packet{
			Raw:  buff,
			Addr: &addr,
		}

		packetChan <- packet
	}
}

func ioControlEx(h windows.Handle, code CtlCode, ioctl unsafe.Pointer, buf *byte, bufLen uint32, overlapped *windows.Overlapped) (iolen uint32, err error) {
	err = windows.DeviceIoControl(h, uint32(code), (*byte)(ioctl), uint32(unsafe.Sizeof(ioCtl{})), buf, bufLen, &iolen, overlapped)
	if err != windows.ERROR_IO_PENDING {
		return
	}

	err = windows.GetOverlappedResult(h, overlapped, &iolen, true)
	return
}

func ioControl(h windows.Handle, code CtlCode, ioctl unsafe.Pointer, buf *byte, bufLen uint32) (iolen uint32, err error) {
	event, _ := windows.CreateEvent(nil, 0, 0, nil)

	overlapped := windows.Overlapped{
		HEvent: event,
	}

	iolen, err = ioControlEx(h, code, ioctl, buf, bufLen, &overlapped)

	windows.CloseHandle(event)
	return
}