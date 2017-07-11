package main

import "fmt"

type Messenger interface {
	Message(string)
	Messagef(string, ...interface{})
}

type IRCMessenger struct {
	*IRCClient
	Channel string
}

func (im *IRCMessenger) Message(msg string) {
	im.Sendf("PRIVMSG #%s :%s", im.Channel, msg)
}

func (im *IRCMessenger) Messagef(s string, a ...interface{}) {
	im.Message(fmt.Sprintf(s, a...))
}
