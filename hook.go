package main

import (
	"encoding/json"
	"log"
)

// The following Event structs and the objects they use only contain the fields
// that are used in this program.

type PushEvent struct {
	Sender  Sender   `json:"sender"`
	Ref     string   `json:"ref"`
	Compare string   `json:"compare"`
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
	Number  int    `json:"number"`
	Title   string `json:"title"`
	HTMLURL string `json:"html_url"`
}

type Comment struct {
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
}

type PullRequest struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Merged  bool   `json:"merged"`
	HTMLURL string `json:"html_url"`
}

type Sender struct {
	Login string `json:"login"`
}

func eventToMessage(eventType string, req []byte, m Messenger) {
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
	case "push":
		var event PushEvent
		if parse(&event) {
			m.OnPush(event)
		}
	case "issues":
		var event IssuesEvent
		if parse(&event) {
			m.OnIssues(event)
		}
	case "issue_comment":
		var event IssueCommentEvent
		if parse(&event) {
			m.OnIssueComment(event)
		}
	case "pull_request":
		var event PullRequestEvent
		if parse(&event) {
			m.OnPullRequest(event)
		}
	default:
		log.Println("ignored event", eventType)
	}
}
