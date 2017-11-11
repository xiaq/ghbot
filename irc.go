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
	queue chan string
}

func dialIRC(disconnect chan<- struct{}) *IRCClient {
	conn, err := tls.Dial("tcp", ircServerAddr, nil)
	if err != nil {
		log.Fatal(err)
	}

	c := &IRCClient{Conn: conn, queue: make(chan string, 1024)}
	// Start a goroutine to echo server responses and respond to PING.
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
	// Start a goroutine to send message to the server, with at least 2 seconds
	// between two messages.
	go func() {
		for msg := range c.queue {
			c.Lock()
			defer c.Unlock()
			c.SetWriteDeadline(time.Now().Add(writeDeadline))
			_, err := c.Write(append([]byte(msg), crLf...))
			if err != nil {
				log.Println("failed to send message:", err)
			}
			time.Sleep(2 * time.Second)
		}
	}()
	return c
}

func (c *IRCClient) Send(msg string) {
	c.queue <- msg
}

func (c *IRCClient) Sendf(s string, args ...interface{}) {
	c.Send(fmt.Sprintf(s, args...))
}
