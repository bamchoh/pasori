package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/bamchoh/pasori"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8080", "http service address")

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

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func clientMain() {
	log.SetFlags(0)

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})
	idm := make(chan string)

	go func() {
		defer close(done)
		for {
			raw, err := pasori.GetID(VID, PID)
			if err != nil {
				log.Println(err)
				return
			}
			idm <- dump_buffer(raw)
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		select {
		case <-done:
			return
		case idmstr := <-idm:
			err := c.WriteMessage(websocket.TextMessage, []byte(idmstr))
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}
}

func main() {
	flag.Parse()
	hub := newHub()
	go hub.run()
	go func() {
		for {
			clientMain()
		}
	}()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAdnServe: ", err)
	}
}
