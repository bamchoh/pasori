package main

import (
	"fmt"

	"crypto/rand"

	"github.com/bamchoh/pasori"
)

func dump_buffer(buf []byte) string {
	str := ""
	for _, b := range buf {
		str += fmt.Sprintf("%02X", b)
	}
	return str
}

var (
	VID uint16 = 0x054C // SONY
	PID uint16 = 0x06C3 // RC-S380
)

func main() {
	var err error
	fmt.Println("Please touch FeliCa")
	psr, err := pasori.InitPasori()
	if err != nil {
		panic(err)
	}
	defer psr.Release()

	psr.RC = make([]byte, 16)
	_, err = rand.Read(psr.RC)
	if err != nil {
		panic(err)
	}
	err = psr.FelicaWriteWithoutEncryption(pasori.RC, psr.RC)
	if err != nil {
		panic(err)
	}

	psr.CK = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	err = psr.FelicaWriteWithMAC_A(pasori.S_PAD0, []byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3})
	if err != nil {
		panic(err)
	}

	blocks, err := psr.FelicaReadWithMAC_A(pasori.SERVICE_RO, pasori.S_PAD0, pasori.S_PAD1)
	if err != nil {
		panic(err)
	}
	fmt.Println("s_pad0", blocks[0])
	fmt.Println("s_pad1", blocks[1])
}
