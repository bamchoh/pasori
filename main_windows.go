// build windows

package main

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

var (
	POLLING_ANY uint16 = 0xFFFF
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

func (p *pasori) felicaPolling(systemcode uint16, rfu uint8, timeslot uint8) (*felica, error) {
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
		return nil, err
	}
	return &f, nil
}

type felica struct {
	IDm [8]uint8
	PMm [8]uint8
}

func main() {
	basepath := "C:\\Program Files\\Common Files\\Sony Shared\\FeliCaLibrary"
	dll, err := syscall.LoadDLL(basepath + "\\" + "felica.dll")
	if err != nil {
		log.Fatal("LoadDLL: ", err)
	}
	defer dll.Release()

	p := pasori{}

	p.initializeLibrary, err = dll.FindProc("initialize_library")
	if err != nil {
		log.Fatal("FindProc: ", err)
	}

	p.disposeLibrary, err = dll.FindProc("dispose_library")
	if err != nil {
		log.Fatal("FindProc: ", err)
	}

	p.openReaderWriterAuto, err = dll.FindProc("open_reader_writer_auto")
	if err != nil {
		log.Fatal("FindProc: ", err)
	}

	p.closeReaderWriter, err = dll.FindProc("close_reader_writer")
	if err != nil {
		log.Fatal("FindProc: ", err)
	}

	p.pollingAndGetCardInformation, err = dll.FindProc("polling_and_get_card_information")
	if err != nil {
		log.Fatal("FindProc: ", err)
	}

	p.pollingAndRequestSystemCode, err = dll.FindProc("polling_and_request_system_code")
	if err != nil {
		log.Fatal("FindProc: ", err)
	}

	p.pollingAndSearchServiceCode, err = dll.FindProc("polling_and_search_service_code")
	if err != nil {
		log.Fatal("FindProc: ", err)
	}

	p.readBlockWithoutEncryption, err = dll.FindProc("read_block_without_encryption")
	if err != nil {
		log.Fatal("FindProc: ", err)
	}

	p.writeBlockWithoutEncryption, err = dll.FindProc("write_block_without_encryption")
	if err != nil {
		log.Fatal("FindProc: ", err)
	}

	win.SetLastError(0)
	ret, _, err := p.initializeLibrary.Call()
	// errno = 128 が返ってくるが felicalib でも返ってくるのでerrチェックはせずに無視する
	if ret == 0 {
		fmt.Println("Failed to initialize felica")
		fmt.Println(err)
		return
	}

	ret, _, err = p.openReaderWriterAuto.Call()
	if ret == 0 {
		fmt.Println("Failed to open felica")
		fmt.Println(err)
		return
	}

	f, err := p.felicaPolling(POLLING_ANY, 0, 0)
	if err != nil {
		fmt.Println("Faild to poll felica")
		fmt.Println(err)
		return
	}

	fmt.Print("IDm: ")
	for i, v := range f.IDm {
		if i != 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%02X", v)
	}
	fmt.Println()

	fmt.Print("PMm: ")
	for i, v := range f.PMm {
		if i != 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%02X", v)
	}
	fmt.Println()
}
