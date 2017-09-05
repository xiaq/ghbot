package main

import (
	"fmt"
	"log"
	"strings"
)

type IRCMessengerZh struct {
	IRCMessengerBase
}

func MakeIRCMessengerZh(ircClient *IRCClient, channel string) Messenger {
	return &IRCMessengerZh{IRCMessengerBase{ircClient, channel}}
}

func (m *IRCMessengerZh) OnPush(event PushEvent) {
	m.Messagef("%s 向 %s推了 %d 个 commit：",
		event.Sender.Login, humanizeRefZh(event.Ref), len(event.Commits))
	for _, commit := range event.Commits {
		m.Messagef("  %s (by %s)",
			firstLine(commit.Message), commit.Author.Name)
	}
}

var issueActionMap = map[string]string{
	"opened": "提出了",
	"closed": "关闭了",
}

func (m *IRCMessengerZh) OnIssues(event IssuesEvent) {
	switch event.Action {
	case "opened", "closed":
		m.Messagef("%s %s issue #%d %s (%s)",
			event.Sender.Login, issueActionMap[event.Action],
			event.Issue.Number, event.Issue.Title, event.Issue.HTMLURL)
	default:
		log.Println("ignored issue being", event.Action)
	}
}

func (m *IRCMessengerZh) OnIssueComment(event IssueCommentEvent) {
	switch event.Action {
	case "created":
		m.Messagef("%s 评论了 issue #%d %s (%s):",
			event.Sender.Login,
			event.Issue.Number, event.Issue.Title, event.Issue.HTMLURL)
		m.Messagef("  %s", abbrCommentZh(event.Comment.Body))
	default:
		log.Println("ignored issue comment being", event.Action)
	}
}

var pullRequestActionMap = map[string]string{
	"opened":   "提出了",
	"closed":   "关闭了",
	"merged":   "合并了",
	"reopened": "重开了",
}

func (m *IRCMessengerZh) OnPullRequest(event PullRequestEvent) {
	switch event.Action {
	case "opened", "closed", "reopened":
		action := event.Action
		if action == "closed" && event.PullRequest.Merged {
			action = "merged"
		}
		m.Messagef("%s %s PR #%d %s (%s)",
			event.Sender.Login, pullRequestActionMap[action],
			event.PullRequest.Number, event.PullRequest.Title,
			event.PullRequest.HTMLURL)
	default:
		log.Println("ignored pull request being", event.Action)
	}
}

func humanizeRefZh(ref string) string {
	if strings.HasPrefix(ref, "refs/heads/") {
		return ref[len("refs/heads/"):] + " 分支"
	}
	return ref
}

func abbrCommentZh(s string) string {
	nrune := 0
	for i, r := range s {
		nrune++
		if r == '\r' || r == '\n' || (r == ' ' && nrune > 100) {
			return fmt.Sprintf("%s ... (略去 %d 字节)", s[:i], len(s)-i)
		}
		if nrune > 120 {
			return fmt.Sprintf("%s... (略去 %d 字节)", s[:i], len(s)-i)
		}
	}
	return s
}
