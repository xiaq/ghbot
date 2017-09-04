package main

import (
	"fmt"
	"log"
	"strings"
)

type IRCMessengerEn struct {
	IRCMessengerBase
}

func MakeIRCMessengerEn(ircClient *IRCClient, channel string) Messenger {
	return &IRCMessengerEn{IRCMessengerBase{ircClient, channel}}
}

func (m *IRCMessengerEn) OnPush(event PushEvent) {
	m.Messagef("%s pushed %s to %s (%s):",
		event.Sender.Login,
		withNumEn(len(event.Commits), "commit", "commits"),
		humanizeRefEn(event.Ref), event.Compare)
	for _, commit := range event.Commits {
		m.Messagef("  %s (by %s)",
			firstLine(commit.Message), commit.Author.Name)
	}
}

func (m *IRCMessengerEn) OnIssues(event IssuesEvent) {
	switch event.Action {
	case "opened", "closed":
		m.Messagef("%s %s issue #%d %s (%s)",
			event.Sender.Login, event.Action,
			event.Issue.Number, event.Issue.Title, event.Issue.URL)
	default:
		log.Println("ignored issue being", event.Action)
	}
}

func (m *IRCMessengerEn) OnIssueComment(event IssueCommentEvent) {
	switch event.Action {
	case "created":
		m.Messagef("%s commented on issue #%d %s (%s):",
			event.Sender.Login,
			event.Issue.Number, event.Issue.Title, event.Issue.URL)
		m.Messagef("  %s", abbrCommentEn(event.Comment.Body))
	default:
		log.Println("ignored issue comment being", event.Action)
	}
}

func (m *IRCMessengerEn) OnPullRequest(event PullRequestEvent) {
	switch event.Action {
	case "opened", "closed", "reopened":
		action := event.Action
		if action == "closed" && event.PullRequest.Merged {
			action = "merged"
		}
		m.Messagef("%s %s pull request #%d %s (%s)",
			event.Sender.Login, action,
			event.PullRequest.Number, event.PullRequest.Title,
			event.PullRequest.URL)
	default:
		log.Println("ignored pull request being", event.Action)
	}
}

func humanizeRefEn(ref string) string {
	if strings.HasPrefix(ref, "refs/heads/") {
		return "branch " + ref[len("refs/heads/"):]
	}
	return ref
}

func abbrCommentEn(s string) string {
	nrune := 0
	for i, r := range s {
		nrune++
		if r == '\r' || r == '\n' || (r == ' ' && nrune > 100) {
			return fmt.Sprintf("%s ... (%d bytes omitted)", s[:i], len(s)-i)
		}
		if nrune > 120 {
			return fmt.Sprintf("%s ...(%d bytes omitted)", s[:i], len(s)-i)
		}
	}
	return s
}

func firstLine(s string) string {
	return strings.SplitN(s, "\n", 2)[0]
}

func withNumEn(n int, single, plural string) string {
	if n == 1 {
		return "1 " + single
	}
	return fmt.Sprintf("%d %s", n, plural)
}
