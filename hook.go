package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// The following Event structs and the objects they use only contain the fields
// that are used in this program.

type PushEvent struct {
	Sender  Sender   `json:"sender"`
	Ref     string   `json:"ref"`
	Commits []Commit `json:"commits"`
}

type IssuesEvent struct {
	Sender Sender `json:"sender"`
	Action string `json:"action"`
	Issue  Issue  `json:"issue"`
}

type IssueCommentEvent struct {
	Sender  Sender  `json:"sender"`
	Action  string  `json:"action"`
	Issue   Issue   `json:"issue"`
	Comment Comment `json:"comment"`
}

type PullRequestEvent struct {
	Sender      Sender      `json:"sender"`
	Action      string      `json:"action"`
	Number      int         `json:"number"`
	PullRequest PullRequest `json:"pull_request"`
}

type Commit struct {
	Author  GitAuthor `json:"author"`
	Message string    `json:"message"`
}

type GitAuthor struct {
	Name string `json:"name"`
}

type Issue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
}

type Comment struct {
	Body string `json:"body"`
}

type PullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Merged bool   `json:"merged"`
}

type Sender struct {
	Login string `json:"login"`
}

func eventToMessage(eventType string, req []byte) []string {
	parse := func(event interface{}) bool {
		err := json.Unmarshal(req, event)
		if err == nil {
			return true
		}
		log.Printf("cannot decode %s event: %v", eventType, err)
		log.Printf("request body was: %s", req)
		return false
	}

	switch eventType {
	case "ping":
		log.Println("pinged")
		return nil
	case "push":
		var event PushEvent
		if !parse(&event) {
			return nil
		}
		lines := []string{
			fmt.Sprintf("@%s pushed %s to %s:",
				event.Sender.Login,
				withNum(len(event.Commits), "commit", "commits"),
				humanizeRef(event.Ref)),
		}
		for _, commit := range event.Commits {
			lines = append(lines, fmt.Sprintf("  %s (by %s)",
				firstLine(commit.Message), commit.Author.Name))
		}
		return lines
	case "issues":
		var event IssuesEvent
		if !parse(&event) {
			return nil
		}
		switch event.Action {
		case "opened", "closed":
			return []string{
				fmt.Sprintf("@%s %s issue #%d (%s)",
					event.Sender.Login, event.Action,
					event.Issue.Number, event.Issue.Title)}
		default:
			log.Println("ignored issue being", event.Action)
		}
		return nil
	case "issue_comment":
		var event IssueCommentEvent
		if !parse(&event) {
			return nil
		}
		switch event.Action {
		case "created":
			return []string{
				fmt.Sprintf("@%s commented on issue #%d (%s):",
					event.Sender.Login, event.Issue.Number, event.Issue.Title),
				fmt.Sprintf("  %s", abbrComment(event.Comment.Body)),
			}
		default:
			log.Println("ignored issue comment being", event.Action)
			return nil
		}
	case "pull_request":
		var event PullRequestEvent
		if !parse(&event) {
			return nil
		}
		switch event.Action {
		case "opened", "closed", "reopened":
			action := event.Action
			if action == "closed" && event.PullRequest.Merged {
				action = "merged"
			}
			return []string{
				fmt.Sprintf("@%s %s pull request #%d (%s)",
					event.Sender.Login, action,
					event.PullRequest.Number, event.PullRequest.Title),
			}
		default:
			log.Println("ignored pull request being", event.Action)
			return nil
		}
	default:
		log.Println("ignored event", eventType)
		return nil
	}
}

func humanizeRef(ref string) string {
	if strings.HasPrefix(ref, "refs/heads/") {
		return "branch " + ref[len("refs/heads/"):]
	}
	return ref
}

func abbrComment(s string) string {
	nrune := 0
	for i, r := range s {
		nrune++
		if r == '\r' || r == '\n' || (r == ' ' && nrune > 100) {
			return s[:i] + " ... (omitted)"
		}
		if nrune > 120 {
			return s[:i] + "... (omitted)"
		}
	}
	return s
}

func firstLine(s string) string {
	return strings.SplitN(s, "\n", 2)[0]
}

func withNum(n int, single, plural string) string {
	if n == 1 {
		return "1 " + single
	}
	return fmt.Sprintf("%d %s", n, plural)
}
