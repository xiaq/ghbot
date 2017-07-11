package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	addrFlag     = flag.String("addr", ":9171", "address to listen to GitHub webhooks")
	initFlag     = flag.String("init", "", "File containing initial instructions")
	channelsFlag = flag.String("channels", "",
		"A comma-separated list of channels to join")
)

func main() {
	flag.Parse()
	channels := strings.Split(*channelsFlag, ",")
	if *initFlag == "" {
		log.Println("-init is required")
		os.Exit(1)
	}
	initBytes, err := ioutil.ReadFile(*initFlag)
	if err != nil {
		log.Println("cannot read init file", err)
		os.Exit(1)
	}
	initMessages := strings.Split(string(initBytes), "\n")

	disconnect := make(chan struct{})
	ircClient := dialIRC(disconnect)

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigs:
			ircClient.Send("QUIT")
			log.Println("interrupted, sent QUIT, exiting after disconnect or 1s")
			select {
			case <-time.After(time.Second):
				log.Println("1s timeout, exiting anyway")
			case <-disconnect:
				log.Println("disconnected, exiting")
			}
		case <-disconnect:
			log.Println("server disconnected, quitting")
		}
		os.Exit(0)
	}()

	for _, msg := range initMessages {
		if msg == "" {
			continue
		}
		ircClient.Send(msg)
	}

	for _, channel := range channels {
		ircClient.Sendf("JOIN :#%s", channel)
	}

	messengers := make([]Messenger, len(channels))
	for i, channel := range channels {
		messengers[i] = &IRCMessenger{ircClient, channel}
	}

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Println("cannot read request:", err)
			return
		}
		for _, m := range messengers {
			eventToMessage(req.Header.Get("X-Github-Event"), body, m)
		}
	})

	log.Fatal(http.ListenAndServe(*addrFlag, nil))
}
