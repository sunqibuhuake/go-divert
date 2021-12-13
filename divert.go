// +build windows

package divert

import (
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)


type Handle struct {
	sync.Mutex
	windows.Handle
	rOverlapped windows.Overlapped
	wOverlapped windows.Overlapped
	Open        uint16
}
var once = sync.Once{}

func GetVersionInfo() (ver string, err error) {
	h, err := Open("false", LayerNetwork, PriorityDefault, FlagDefault)
	if err != nil {
		return
	}
	defer h.Close()

	major, err := h.GetParam(VersionMajor)
	if err != nil {
		return
	}

	minor, err := h.GetParam(VersionMinor)
	if err != nil {
		return
	}

	ver = strings.Join([]string{strconv.Itoa(int(major)), strconv.Itoa(int(minor))}, ".")
	return
}

func (h *Handle) Recv(buffer []byte, address *Address) (uint, error) {
	addrLen := uint(unsafe.Sizeof(Address{}))
	recv := recv{
		Addr:       uint64(uintptr(unsafe.Pointer(address))),
		AddrLenPtr: uint64(uintptr(unsafe.Pointer(&addrLen))),
	}

	iolen, err := ioControlEx(h.Handle, ioCtlRecv, unsafe.Pointer(&recv), &buffer[0], uint32(len(buffer)), &h.rOverlapped)
	if err != nil {
		return uint(iolen), Error(err.(windows.Errno))
	}

	return uint(iolen), nil
}

func (h *Handle) RecvEx(buffer []byte, address []Address) (uint, uint, error) {
	addrLen := uint(len(address)) * uint(unsafe.Sizeof(Address{}))
	recv := recv{
		Addr:       uint64(uintptr(unsafe.Pointer(&address[0]))),
		AddrLenPtr: uint64(uintptr(unsafe.Pointer(&addrLen))),
	}

	iolen, err := ioControlEx(h.Handle, ioCtlRecv, unsafe.Pointer(&recv), &buffer[0], uint32(len(buffer)), &h.rOverlapped)
	if err != nil {
		return uint(iolen), addrLen / uint(unsafe.Sizeof(Address{})), Error(err.(windows.Errno))
	}

	return uint(iolen), addrLen / uint(unsafe.Sizeof(Address{})), nil
}

func (h *Handle) Send(buffer []byte, address *Address) (uint, error) {
	send := send{
		Addr:    uint64(uintptr(unsafe.Pointer(address))),
		AddrLen: uint64(unsafe.Sizeof(Address{})),
	}

	iolen, err := ioControlEx(h.Handle, ioCtlSend, unsafe.Pointer(&send), &buffer[0], uint32(len(buffer)), &h.wOverlapped)
	if err != nil {
		return uint(iolen), Error(err.(windows.Errno))
	}

	return uint(iolen), nil
}

func (h *Handle) SendEx(buffer []byte, address []Address) (uint, error) {
	send := send{
		Addr:    uint64(uintptr(unsafe.Pointer(&address[0]))),
		AddrLen: uint64(unsafe.Sizeof(Address{})) * uint64(len(address)),
	}

	iolen, err := ioControlEx(h.Handle, ioCtlSend, unsafe.Pointer(&send), &buffer[0], uint32(len(buffer)), &h.wOverlapped)
	if err != nil {
		return uint(iolen), Error(err.(windows.Errno))
	}

	return uint(iolen), nil
}

func (h *Handle) Shutdown(how Shutdown) error {
	h.Open = HandleShutdown

	shutdown := shutdown{
		How: uint32(how),
	}

	_, err := ioControl(h.Handle, ioCtlShutdown, unsafe.Pointer(&shutdown), nil, 0)
	if err != nil {
		return Error(err.(windows.Errno))
	}

	return nil
}

func (h *Handle) Close() error {
	h.Open = HandleClosed

	windows.CloseHandle(h.rOverlapped.HEvent)
	windows.CloseHandle(h.wOverlapped.HEvent)

	err := windows.CloseHandle(h.Handle)
	if err != nil {
		return Error(err.(windows.Errno))
	}

	return nil
}

func (h *Handle) End() (err error) {
	h.Open = HandleEnded

	err = h.Shutdown(ShutdownBoth)
	if err != nil {
		return
	}
	err = h.Close()
	if err != nil {
		return
	}

	return nil
}

func (h *Handle) Packets() (chan *Packet, error) {
	if h.Open != HandleOpen {
		return nil, errNotOpen
	}

	packetChan := make(chan *Packet, PacketChanCapacity)
	go h.recvLoop(packetChan)
	return packetChan, nil
}

func (h *Handle) GetParam(p Param) (uint64, error) {
	getParam := getParam{
		Param: uint32(p),
		Value: 0,
	}

	_, err := ioControl(h.Handle, ioCtlGetParam, unsafe.Pointer(&getParam), (*byte)(unsafe.Pointer(&getParam.Value)), uint32(unsafe.Sizeof(getParam.Value)))
	if err != nil {
		return getParam.Value, Error(err.(windows.Errno))
	}

	return getParam.Value, nil
}

func (h *Handle) SetParam(p Param, v uint64) error {
	switch p {
	case QueueLength:
		if v < QueueLengthMin || v > QueueLengthMax {
			return errQueueLength
		}
	case QueueTime:
		if v < QueueTimeMin || v > QueueTimeMax {
			return errQueueTime
		}
	case QueueSize:
		if v < QueueSizeMin || v > QueueSizeMax {
			return errQueueSize
		}
	default:
		return errQueueParam
	}

	setParam := setParam{
		Value: v,
		Param: uint32(p),
	}

	_, err := ioControl(h.Handle, ioCtlSetParam, unsafe.Pointer(&setParam), nil, 0)
	if err != nil {
		return Error(err.(windows.Errno))
	}

	return nil
}



