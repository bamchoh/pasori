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

	b, err := psr.FelicaReadWithoutEncryption(0)
	if err != nil {
		panic(err)
	}
	fmt.Println(b)

	wb := make([]byte, len(b))
	for i := 0; i < len(b); i++ {
		wb[i] = b[i] + 1
	}

	err = psr.FelicaWriteWithoutEncryption(0, wb)
	if err != nil {
		panic(err)
	}

	b, err = psr.FelicaReadWithoutEncryption(0)
	if err != nil {
		panic(err)
	}
	fmt.Println(b)
}
