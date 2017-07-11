package main

import (
	"fmt"
	"log"
)

type Messenger interface {
	OnPush(PushEvent)
	OnIssues(IssuesEvent)
	OnIssueComment(IssueCommentEvent)
	OnPullRequest(PullRequestEvent)
}

type IRCMessenger struct {
	*IRCClient
	Channel string
}

func (m *IRCMessenger) Message(msg string) {
	m.Sendf("PRIVMSG #%s :%s", m.Channel, msg)
}

func (m *IRCMessenger) Messagef(s string, a ...interface{}) {
	m.Message(fmt.Sprintf(s, a...))
}

func (m *IRCMessenger) OnPush(event PushEvent) {
	m.Messagef("@%s pushed %s to %s:",
		event.Sender.Login,
		withNum(len(event.Commits), "commit", "commits"),
		humanizeRef(event.Ref))
	for _, commit := range event.Commits {
		m.Messagef("  %s (by %s)",
			firstLine(commit.Message), commit.Author.Name)
	}
}

func (m *IRCMessenger) OnIssues(event IssuesEvent) {
	switch event.Action {
	case "opened", "closed":
		m.Messagef("@%s %s issue #%d (%s)",
			event.Sender.Login, event.Action,
			event.Issue.Number, event.Issue.Title)
	default:
		log.Println("ignored issue being", event.Action)
	}
}

func (m *IRCMessenger) OnIssueComment(event IssueCommentEvent) {
	switch event.Action {
	case "created":
		m.Messagef("@%s commented on issue #%d (%s):",
			event.Sender.Login, event.Issue.Number, event.Issue.Title)
		m.Messagef("  %s", abbrComment(event.Comment.Body))
	default:
		log.Println("ignored issue comment being", event.Action)
	}
}

func (m *IRCMessenger) OnPullRequest(event PullRequestEvent) {
	switch event.Action {
	case "opened", "closed", "reopened":
		action := event.Action
		if action == "closed" && event.PullRequest.Merged {
			action = "merged"
		}
		m.Messagef("@%s %s pull request #%d (%s)",
			event.Sender.Login, action,
			event.PullRequest.Number, event.PullRequest.Title)
	default:
		log.Println("ignored pull request being", event.Action)
	}
}
