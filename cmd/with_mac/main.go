package main

import (
	"fmt"

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

	wb := make([]byte, 16)
	for i := 0; i < len(wb); i++ {
		wb[i] = byte(i + 1)
	}

	err = psr.FelicaWriteWithoutEncryption(pasori.RC, wb)
	if err != nil {
		panic(err)
	}

	var b [][16]byte
	b, err = psr.FelicaReadWithoutEncryption(pasori.SERVICE_RO, pasori.ID, pasori.CKV, pasori.MAC)
	if err != nil {
		panic(err)
	}
	fmt.Println(b)

	b, err = psr.FelicaReadWithoutEncryption(pasori.SERVICE_RO, pasori.S_PAD0)
	if err != nil {
		panic(err)
	}
	fmt.Println(b)
}
