// +build windows,!divert_cgo,!divert_embedded
// +build amd64 arm64

package divert

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	winDivert     					= (*windows.DLL)(nil)
	winDivertOpen 					= (*windows.Proc)(nil)
	winDivertClose  				= (*windows.Proc)(nil)
	winDivertRecv               	= (*windows.Proc)(nil)
	winDivertSend                 	= (*windows.Proc)(nil)
	winDivertHelperCalcChecksums  	= (*windows.Proc)(nil)
	winDivertHelperEvalFilter     	= (*windows.Proc)(nil)
	//winDivertHelperCheckFilter    	= (*windows.Proc)(nil)
)

func Open(filter string, layer Layer, priority int16, flags uint64) (h *Handle, err error) {
	once.Do(func() {
		dll, er := windows.LoadDLL("WinDivert.dll")
		if er != nil {
			err = er
			return
		}
		winDivert = dll

		// WinDivertOpen
		proc, er := winDivert.FindProc("WinDivertOpen")
		if er != nil {
			err = er
			return
		}
		winDivertOpen = proc

		// WinDivertHelperCalcChecksums
		winDivertHelperCalcChecksums,  er = winDivert.FindProc("WinDivertHelperCalcChecksums")
		if er != nil {
			err = er
			return
		}

		// WinDivertHelperCheckFilter
		//winDivertHelperCheckFilter,  er = winDivert.FindProc("WinDivertHelperCheckFilter")
		//if er != nil {
		//	err = er
		//	return
		//}

		vers := map[string]struct{}{
			"2.0": {},
			"2.1": {},
			"2.2": {},
		}
		ver, er := func() (ver string, err error) {
			h, err := open("false", LayerNetwork, PriorityDefault, FlagDefault)
			if err != nil {
				return
			}
			defer func() {
				err = h.Close()
			}()

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
		}()
		if er != nil {
			err = er
			return
		}
		if _, ok := vers[ver]; !ok {
			err = fmt.Errorf("unsupported windivert version: %v", ver)
		}
	})
	if err != nil {
		return
	}

	return open(filter, layer, priority, flags)
}

func open(filter string, layer Layer, priority int16, flags uint64) (h *Handle, err error) {
	if priority < PriorityLowest || priority > PriorityHighest {
		return nil, errPriority
	}

	filterPtr, err := windows.BytePtrFromString(filter)
	if err != nil {
		return nil, err
	}

	runtime.LockOSThread()
	hd, _, err := winDivertOpen.Call(uintptr(unsafe.Pointer(filterPtr)), uintptr(layer), uintptr(priority), uintptr(flags))
	runtime.UnlockOSThread()

	if windows.Handle(hd) == windows.InvalidHandle {
		return nil, Error(err.(windows.Errno))
	}

	rEvent, _ := windows.CreateEvent(nil, 0, 0, nil)
	wEvent, _ := windows.CreateEvent(nil, 0, 0, nil)

	return &Handle{
		Mutex:  sync.Mutex{},
		Handle: windows.Handle(hd),
		rOverlapped: windows.Overlapped{
			HEvent: rEvent,
		},
		wOverlapped: windows.Overlapped{
			HEvent: wEvent,
		},
		Open: HandleOpen,
	}, nil
}


func (h *Handle) HelperCalcChecksum(packet *Packet) {
	winDivertHelperCalcChecksums.Call(
		uintptr(unsafe.Pointer(&packet.Raw[0])),
		uintptr(len(packet.Raw)),
		uintptr(unsafe.Pointer(&packet.Addr)),
		uintptr(0))
}

//
//func HelperCheckFilter(filter string) (bool, error) {
//	var errorPos uint
//
//	filterBytePtr, _ := syscall.BytePtrFromString(filter)
//
//	success, _, _ := winDivertHelperCheckFilter.Call(
//		uintptr(unsafe.Pointer(filterBytePtr)),
//		uintptr(0),
//		uintptr(0), // Not implemented yet
//		uintptr(unsafe.Pointer(&errorPos)))
//
//	if success == 1 {
//		return true, nil
//	}
//	return false, errors.New("invalid filter")
//}
//
//


