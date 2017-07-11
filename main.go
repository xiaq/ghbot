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
	languagesFlag = flag.String("languages", "",
		"A common-separated list of languages for each channel.")
)

func main() {
	flag.Parse()
	channels := strings.Split(*channelsFlag, ",")
	languages := strings.Split(*languagesFlag, ",")
	require := func(s string, arg string) {
		if s == "" {
			log.Println(arg, "is required")
			os.Exit(1)
		}
	}
	require(*initFlag, "-init")
	require(*channelsFlag, "-channels")
	require(*languagesFlag, "-languages")

	if len(channels) != len(languages) {
		log.Println("-channels and -languages should have the same number of elements")
		os.Exit(1)
	}
	for _, language := range languages {
		if _, ok := messengerMakers[language]; !ok {
			log.Printf("language %q not supported", language)
			os.Exit(1)
		}
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
		messengers[i] = messengerMakers[languages[i]](ircClient, channel)
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
