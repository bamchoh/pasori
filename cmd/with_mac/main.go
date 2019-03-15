package main

import (
	"crypto/cipher"
	"crypto/des"
	"errors"
	"fmt"

	"github.com/bamchoh/pasori"
)

func reverse(b []byte) []byte {
	a := make([]byte, len(b))
	for i, j := 0, len(b)-1; i < len(b); i, j = i+1, j-1 {
		a[i] = b[j]
	}
	return a
}

func xor(a []byte, b []byte) []byte {
	c := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		c[i] = a[i] ^ b[i]
	}
	return c
}

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

func my3des(c1, c2 cipher.Block, dst, src []byte) {
	enc1 := make([]byte, des.BlockSize)
	enc2 := make([]byte, des.BlockSize)
	c1.Encrypt(enc1, src)
	c2.Decrypt(enc2, enc1)
	c1.Encrypt(dst, enc2)
}

func des2(ck [16]byte, rc [16]byte) ([16]byte, error) {
	var sk [16]byte

	key1 := reverse(ck[:8])
	key2 := reverse(ck[8:])
	c1, err := des.NewCipher(key1)
	if err != nil {
		return sk, err
	}
	c2, err := des.NewCipher(key2)
	if err != nil {
		return sk, err
	}

	rc1 := rc[:8]
	rc2 := rc[8:]
	enc1 := make([]byte, des.BlockSize)
	enc2 := make([]byte, des.BlockSize)

	my3des(c1, c2, enc1, reverse(rc1))
	my3des(c1, c2, enc2, xor(enc1, reverse(rc2)))

	sk1 := reverse(enc1)
	sk2 := reverse(enc2)

	for i := 0; i < len(sk1); i++ {
		sk[i] = sk1[i]
	}

	for i := 0; i < len(sk2); i++ {
		sk[i+8] = sk2[i]
	}
	return sk, nil
}

func des1(ck []byte, rc []byte) ([]byte, error) {
	if len(ck) < 16 {
		return nil, errors.New("Length of CK is less than 16")
	}
	if len(rc) < 16 {
		return nil, errors.New("Length of RC is less than 16")
	}

	key := make([]byte, 24)
	copy(key[:8], reverse(ck[:8]))
	copy(key[8:16], reverse(ck[8:]))
	copy(key[16:], reverse(ck[:8]))
	c, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	rc1 := rc[:8]
	rc2 := rc[8:16]
	enc1 := make([]byte, des.BlockSize)
	enc2 := make([]byte, des.BlockSize)

	c.Encrypt(enc1, reverse(rc1))
	c.Encrypt(enc2, xor(enc1, reverse(rc2)))
	sk1 := reverse(enc1)
	sk2 := reverse(enc2)

	sk := make([]byte, 16)
	copy(sk[:8], sk1)
	copy(sk[8:], sk2)
	return sk, nil
}

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

	ck := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	// err = psr.FelicaWriteWithoutEncryption(pasori.CK, ck[:])
	// if err != nil {
	// 	panic(err)
	// }

	id, err := psr.FelicaReadWithoutEncryption(pasori.SERVICE_RO, pasori.S_PAD0, pasori.MAC)
	if err != nil {
		panic(err)
	}
	sk, _ := des1(ck, wb)
	bd := make([]byte, 16)
	copy(bd[:8], xor(id[0][:8], wb[:8]))
	copy(bd[8:], id[0][8:])
	mac, _ := des1(sk, bd[:])
	fmt.Println("ck:  ", ck)
	fmt.Println("id:  ", id[0])
	fmt.Println("mac: ", id[1][:8])
	fmt.Println("mac1:", mac[8:16])

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
