// build windows

package pasori

import (
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/win"
)

func h2ns(x uint16) uint16 {
	return (((x)>>8)&0xff | ((x)<<8)&0xff00)
}

type polling struct {
	systemcode *uint16
	timeslot   uint8
}

type cardinfo struct {
	cardIdm uintptr
	cardPmm uintptr
}

type pasori struct {
	initializeLibrary            *syscall.Proc
	disposeLibrary               *syscall.Proc
	openReaderWriterAuto         *syscall.Proc
	closeReaderWriter            *syscall.Proc
	pollingAndGetCardInformation *syscall.Proc
	pollingAndRequestSystemCode  *syscall.Proc
	pollingAndSearchServiceCode  *syscall.Proc
	readBlockWithoutEncryption   *syscall.Proc
	writeBlockWithoutEncryption  *syscall.Proc
}

func (p *pasori) felicaPolling(systemcode uint16, timeslot uint8) (*felica, error) {
	f := felica{}
	sc := h2ns(systemcode)
	poll := polling{
		systemcode: &sc,
		timeslot:   timeslot,
	}
	card := cardinfo{
		cardIdm: uintptr(unsafe.Pointer(&f.IDm)),
		cardPmm: uintptr(unsafe.Pointer(&f.PMm)),
	}
	var numberOfCards uint8
	ret, _, err := p.pollingAndGetCardInformation.Call(uintptr(unsafe.Pointer(&poll.systemcode)), uintptr(unsafe.Pointer(&numberOfCards)), uintptr(unsafe.Pointer(&card.cardIdm)))
	if ret == 0 {
		if err.(syscall.Errno) == 0 {
			return nil, nil
		}
		return nil, err
	}
	if f.isAllZero() {
		return nil, nil
	}
	return &f, nil
}

type felica struct {
	IDm [8]uint8
	PMm [8]uint8
}

func (f *felica) isAllZero() bool {
	for _, v := range f.IDm {
		if v != 0 {
			return false
		}
	}
	for _, v := range f.PMm {
		if v != 0 {
			return false
		}
	}
	return true
}

func GetID(vid, pid uint16) ([]byte, error) {
	basepath := "C:\\Program Files\\Common Files\\Sony Shared\\FeliCaLibrary"
	dll, err := syscall.LoadDLL(basepath + "\\" + "felica.dll")
	if err != nil {
		return nil, err
	}
	defer dll.Release()

	p := pasori{}

	p.initializeLibrary, err = dll.FindProc("initialize_library")
	if err != nil {
		return nil, err
	}

	p.disposeLibrary, err = dll.FindProc("dispose_library")
	if err != nil {
		return nil, err
	}

	p.openReaderWriterAuto, err = dll.FindProc("open_reader_writer_auto")
	if err != nil {
		return nil, err
	}

	p.pollingAndGetCardInformation, err = dll.FindProc("polling_and_get_card_information")
	if err != nil {
		return nil, err
	}

	win.SetLastError(0)
	ret, _, err := p.initializeLibrary.Call()
	// errno = 128 が返ってくるが felicalib でも返ってくるのでerrチェックはせずに無視する
	if ret == 0 {
		return nil, err
	}

	ret, _, err = p.openReaderWriterAuto.Call()
	if ret == 0 {
		return nil, err
	}
	defer p.disposeLibrary.Call()

	var f *felica
	isloop := true
	for isloop {
		f, err = p.felicaPolling(0xFFFF, 0) // 0xFFFF is POLLING_ANY
		if err != nil {
			return nil, err
		}
		if f != nil {
			isloop = false
		}
		time.Sleep(1 * time.Millisecond)
	}

	return f.IDm[:], nil
}
