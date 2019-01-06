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
	fmt.Println("Please touch FeliCa")
	idm, err := pasori.GetID(VID, PID)
	if err != nil {
		panic(err)
	}
	fmt.Println("IDm:", dump_buffer(idm))
}
