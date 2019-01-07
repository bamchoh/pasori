package main

import (
	"fmt"
	"net/http"

	"github.com/bamchoh/pasori"
	"github.com/gin-gonic/gin"
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
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		fmt.Println("in")
		idm, err := pasori.GetID(VID, PID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id": dump_buffer(idm),
		})
	})
	r.Run()
}
