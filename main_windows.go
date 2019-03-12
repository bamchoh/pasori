// build windows

package pasori

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/win"
)

const (
	MAX_AREA_CODE    = 16
	MAX_SERVICE_CODE = 256
)

const (
	S_PAD0 = iota
	S_PAD1
	S_PAD2
	S_PAD3
	S_PAD4
	S_PAD5
	S_PAD6
	S_PAD7
	S_PAD8
	S_PAD9
	S_PAD10
	S_PAD11
	S_PAD12
	S_PAD13
	REG
)

const (
	RC = iota + 0x80
	MAC
	ID
	D_ID
	SER_C
	SYS_C
	CKV
	CK
	MC
)

const (
	WCNT = iota + 0x90
	MAC_A
	STATE
	CRC_CHECK = 0xA0
)

const (
	SERVICE_RW = 0x09
	SERVICE_RO = 0x0B
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

type instrReadBlock struct {
	cardIdm          uintptr
	numberOfServices uint8
	serviceCodeList  uintptr
	numberOfBlocks   uint8
	blockList        uintptr
}

type outstrReadBlock struct {
	statusFlag1          uintptr
	statusFlag2          uintptr
	resultNumberOfBlocks uintptr
	blockData            uintptr
}

type instrWriteBlock struct {
	cardIdm          uintptr
	numberOfServices uint8
	serviceCodeList  uintptr
	numberOfBlocks   uint8
	blockList        uintptr
	blockData        uintptr
}

type outstrWriteBlock struct {
	statusFlag1 uintptr
	statusFlag2 uintptr
}

type instrSearchService struct {
	bufferSizeOfAreaCodes    int
	bufferSizeOfServiceCodes int
	offsetOfAreaServiceIndex int
}

type outstrSearchService struct {
	numServiceCodes    int
	serviceCodeList    uintptr
	numAreaCodes       int
	areaCodeList       uintptr
	endServiceCodeList uintptr
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
	dll                          *syscall.DLL
}

func (p *pasori) felicaEnumService(systemcode uint16) (*felica, error) {
	f := felica{}
	f.SystemCode = h2ns(systemcode)
	poll := polling{
		systemcode: &f.SystemCode,
		timeslot:   0,
	}
	card := cardinfo{
		cardIdm: uintptr(unsafe.Pointer(&f.IDm)),
		cardPmm: uintptr(unsafe.Pointer(&f.PMm)),
	}
	iss := instrSearchService{
		bufferSizeOfAreaCodes:    MAX_AREA_CODE,
		bufferSizeOfServiceCodes: MAX_SERVICE_CODE,
		offsetOfAreaServiceIndex: 0,
	}
	oss := outstrSearchService{
		numAreaCodes:       10,
		numServiceCodes:    10,
		serviceCodeList:    uintptr(unsafe.Pointer(&f.ServiceCode[0])),
		areaCodeList:       uintptr(unsafe.Pointer(&f.AreaCode[0])),
		endServiceCodeList: uintptr(unsafe.Pointer(&f.EndServiceCode[0])),
	}
	ret, _, err := p.pollingAndSearchServiceCode.Call(uintptr(unsafe.Pointer(&poll.systemcode)), uintptr(unsafe.Pointer(&iss.bufferSizeOfAreaCodes)), uintptr(unsafe.Pointer(&card.cardIdm)), uintptr(unsafe.Pointer(&oss.numAreaCodes)))
	fmt.Println(ret)
	fmt.Println(err)
	fmt.Println(f)
	if ret == 0 {
		if err.(syscall.Errno) == 0 {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
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

func (p *pasori) felicaReadWithoutEncryption(idm *[8]uint8, servicecode uint16, blknum ...uint8) ([][16]byte, error) {
	var serviceCode [2]uint8
	serviceCode[0] = uint8(servicecode & 0xff)
	serviceCode[1] = uint8(servicecode >> 8)

	blockList := make([]byte, len(blknum)*2)
	for i := 0; i < len(blknum); i++ {
		blockList[2*i] = 0x80
		blockList[2*i+1] = blknum[i]
	}

	irb := instrReadBlock{
		cardIdm:          uintptr(unsafe.Pointer(&idm[0])),
		numberOfServices: 1,
		serviceCodeList:  uintptr(unsafe.Pointer(&serviceCode[0])),
		numberOfBlocks:   uint8(len(blknum)),
		blockList:        uintptr(unsafe.Pointer(&blockList[0])),
	}

	var statusFlag1 uint8
	var statusFlag2 uint8
	var resultNumberOfBlocks uint8
	blockData := make([][16]uint8, len(blknum))

	orb := outstrReadBlock{
		statusFlag1:          uintptr(unsafe.Pointer(&statusFlag1)),
		statusFlag2:          uintptr(unsafe.Pointer(&statusFlag2)),
		resultNumberOfBlocks: uintptr(unsafe.Pointer(&resultNumberOfBlocks)),
		blockData:            uintptr(unsafe.Pointer(&blockData[0][0])),
	}

	ret, _, err := p.readBlockWithoutEncryption.Call(uintptr(unsafe.Pointer(&irb.cardIdm)), uintptr(unsafe.Pointer(&orb.statusFlag1)))

	if ret == 0 {
		if err.(syscall.Errno) == 0 {
			return blockData[:], nil
		}
		return nil, err
	}

	return blockData[:], nil
}

func (p *pasori) felicaWriteWithoutEncryption(idm *[8]uint8, servicecode uint16, blknum uint8, data []byte) error {
	var serviceCode [2]uint8
	serviceCode[0] = uint8(servicecode & 0xff)
	serviceCode[1] = uint8(servicecode >> 8)

	var blockList [2]uint8
	blockList[0] = 0x80
	blockList[1] = blknum

	var blockData [16]uint8
	for i := 0; i < len(blockData); i++ {
		blockData[i] = data[i]
	}

	iwr := instrWriteBlock{
		cardIdm:          uintptr(unsafe.Pointer(&idm[0])),
		numberOfServices: 1,
		serviceCodeList:  uintptr(unsafe.Pointer(&serviceCode[0])),
		numberOfBlocks:   1,
		blockList:        uintptr(unsafe.Pointer(&blockList[0])),
		blockData:        uintptr(unsafe.Pointer(&blockData[0])),
	}

	var statusFlag1 uint8
	var statusFlag2 uint8

	owr := outstrWriteBlock{
		statusFlag1: uintptr(unsafe.Pointer(&statusFlag1)),
		statusFlag2: uintptr(unsafe.Pointer(&statusFlag2)),
	}
	ret, _, err := p.writeBlockWithoutEncryption.Call(uintptr(unsafe.Pointer(&iwr.cardIdm)), uintptr(unsafe.Pointer(&owr.statusFlag1)))
	if ret == 0 {
		if err.(syscall.Errno) == 0 {
			return nil
		}
		return err
	}
	return nil
}

func (p *pasori) Release() {
	defer p.disposeLibrary.Call()
	// defer p.dll.Release()
}

func (p *pasori) FelicaEnumService() (*felica, error) {
	win.SetLastError(0)
	return p.felicaEnumService(0xFFFF)
}

func (p *pasori) FelicaReadWithoutEncryption(servicecode uint16, blkno ...uint8) ([][16]byte, error) {
	win.SetLastError(0)
	idm, err := p.GetIdm()
	if err != nil {
		return nil, err
	}
	var idmary [8]uint8
	for i := 0; i < len(idmary); i++ {
		idmary[i] = idm[i]
	}
	return p.felicaReadWithoutEncryption(&idmary, servicecode, blkno...)
}

func (p *pasori) FelicaWriteWithoutEncryption(blkno uint8, data []byte) error {
	win.SetLastError(0)
	idm, err := p.GetIdm()
	if err != nil {
		return err
	}
	var idmary [8]uint8
	for i := 0; i < len(idmary); i++ {
		idmary[i] = idm[i]
	}
	return p.felicaWriteWithoutEncryption(&idmary, 9, blkno, data)
}

func (p *pasori) GetIdm() ([]byte, error) {
	var err error
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

type felica struct {
	IDm            [8]uint8
	PMm            [8]uint8
	SystemCode     uint16
	AreaCode       [MAX_AREA_CODE]uint16
	EndServiceCode [MAX_AREA_CODE]uint16
	ServiceCode    [MAX_SERVICE_CODE]uint16
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

func InitPasori() (*pasori, error) {
	basepath := "C:\\Program Files\\Common Files\\Sony Shared\\FeliCaLibrary"
	dll, err := syscall.LoadDLL(basepath + "\\" + "felica.dll")
	if err != nil {
		return nil, err
	}

	p := pasori{dll: dll}

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

	p.pollingAndSearchServiceCode, err = dll.FindProc("polling_and_search_service_code")
	if err != nil {
		return nil, err
	}

	p.readBlockWithoutEncryption, err = dll.FindProc("read_block_without_encryption")
	if err != nil {
		return nil, err
	}

	p.writeBlockWithoutEncryption, err = dll.FindProc("write_block_without_encryption")
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
	return &p, nil
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

	p.pollingAndSearchServiceCode, err = dll.FindProc("polling_and_search_service_code")
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
