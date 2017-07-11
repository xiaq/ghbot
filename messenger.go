package main

import (
	"fmt"
)

var messengerMakers = map[string]func(*IRCClient, string) Messenger{
	"en": MakeIRCMessengerEn,
	"zh": MakeIRCMessengerZh,
}

type Messenger interface {
	OnPush(PushEvent)
	OnIssues(IssuesEvent)
	OnIssueComment(IssueCommentEvent)
	OnPullRequest(PullRequestEvent)
}

type IRCMessengerBase struct {
	*IRCClient
	Channel string
}

func (m *IRCMessengerBase) Message(msg string) {
	m.Sendf("PRIVMSG #%s :%s", m.Channel, msg)
}

func (m *IRCMessengerBase) Messagef(s string, a ...interface{}) {
	m.Message(fmt.Sprintf(s, a...))
}
