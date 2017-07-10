package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	readBufSize = 4096
)

var (
	ircServerAddr = "irc.freenode.net:6697"
	writeDeadline = 5 * time.Second
	crLf          = []byte("\r\n")
	ping          = []byte("PING")
)

type IRCClient struct {
	*tls.Conn
	sync.Mutex
}

func dialIRC(disconnect chan<- struct{}) *IRCClient {
	conn, err := tls.Dial("tcp", ircServerAddr, nil)
	if err != nil {
		log.Fatal(err)
	}

	c := &IRCClient{Conn: conn}
	go func() {
		var buf [readBufSize]byte
		for {
			nr, err := conn.Read(buf[:])
			if err != nil {
				log.Println("read error:", err)
				break
			}
			log.Printf("server sent %d bytes: %s", nr, buf[:nr])
			if bytes.HasPrefix(buf[:nr], ping) {
				rest := buf[len(ping) : nr-2]
				c.Sendf("PONG%s", rest)
			}
		}
		close(disconnect)
	}()
	return c
}

func (c *IRCClient) Send(msg string) {
	c.Lock()
	defer c.Unlock()
	c.SetWriteDeadline(time.Now().Add(writeDeadline))
	_, err := c.Write(append([]byte(msg), crLf...))
	if err != nil {
		log.Fatal(err)
	}
}

func (c *IRCClient) Sendf(s string, args ...interface{}) {
	c.Send(fmt.Sprintf(s, args...))
}
