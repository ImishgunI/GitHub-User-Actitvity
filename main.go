package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Event struct {
	Message string
}

var mp []map[string]any

func main() {
	username := getUsername()
	data := jsonHandle(username)
	parseToMap(data)
	msg := make([]Event, len(mp))
	j := 0
	for i := range mp {
		switch mp[i]["type"] {
		case "PushEvent":
			j = PushEvent(msg, j, i)
		case "CreateEvent":
			j = CreateEvent(msg, j, i)
		case "WatchEvent":
			j = WatchEvent(msg, j, i)
		case "PullRequestEvent":
			j = PullRequestEvent(msg, j, i)
		}
	}
	printMessage(msg)
}

func getUsername() string {
	if len(os.Args) < 2 {
		log.Fatal("You must write a username")
	}
	return os.Args[1]
}

func printMessage(msg []Event) {
	for _, v := range msg {
		if v.Message != "" {
			fmt.Println(v.Message)
		}
	}
}

func jsonHandle(username string) []byte {
	client := &http.Client{}
	url := fmt.Sprintf("https://api.github.com/users/%s/events", username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func parseToMap(data []byte) {
	err := json.Unmarshal(data, &mp)
	if err != nil {
		log.Fatal(err)
	}
}

func PushEvent(msg []Event, j, i int) int {
	var (
		login     string
		repo_name string
		message   string
		size      float64
	)
	if actor, ok := mp[i]["actor"].(map[string]any); ok {
		login = actor["login"].(string)
	}
	if repo, ok := mp[i]["repo"].(map[string]any); ok {
		repo_name = repo["name"].(string)
	}

	if payload, ok := mp[i]["payload"].(map[string]any); ok {
		size = payload["size"].(float64)
		if size > 0 {
			k := 0
			if commits, ok := payload["commits"].([]any); ok && len(commits) > 0 {
				if commitData, ok := commits[k].(map[string]any); ok {
					message, _ = commitData["message"].(string)
				}
				k++
			}
		} else {
			return j
		}
	}
	msg[j].Message = fmt.Sprintf("%s pushed message: \"%s\" to repository: \"%s\"", login, message, repo_name)
	j++
	return j
}

func CreateEvent(msg []Event, j, i int) int {
	var (
		login     string
		repo_name string
		ref       string
		ref_type  string
	)
	if actor, ok := mp[i]["actor"].(map[string]any); ok {
		login = actor["login"].(string)
	}
	if repo, ok := mp[i]["repo"].(map[string]any); ok {
		repo_name = repo["name"].(string)
	}
	if payload, ok := mp[i]["payload"].(map[string]any); ok {
		if payload["ref"] != nil {
			ref = payload["ref"].(string)
		}
		ref_type = payload["ref_type"].(string)
	}
	if ref_type == "repository" {
		msg[j].Message = fmt.Sprintf("%s created \"%s\", repository name: \"%s\"", login, ref_type, repo_name)
	} else {
		msg[j].Message = fmt.Sprintf("%s created \"%s\" with name \"%s\" in repo: \"%s\"", login, ref_type, ref, repo_name)
	}
	j++
	return j
}

func WatchEvent(msg []Event, j, i int) int {
	var (
		login     string
		repo_name string
		action    string
	)
	if actor, ok := mp[i]["actor"].(map[string]any); ok {
		login = actor["login"].(string)
	}
	if repo, ok := mp[i]["repo"].(map[string]any); ok {
		repo_name = repo["name"].(string)
	}
	if payload, ok := mp[i]["payload"].(map[string]any); ok {
		action = payload["action"].(string)
	}
	msg[j].Message = fmt.Sprintf("%s %s a repo \"%s\"", login, action, repo_name)
	j++
	return j
}

func PullRequestEvent(msg []Event, j, i int) int {
	var (
		login     string
		repo_name string
		action    string
		title     string
		body      string
	)
	if actor, ok := mp[i]["actor"].(map[string]any); ok {
		login = actor["login"].(string)
	}
	if repo, ok := mp[i]["repo"].(map[string]any); ok {
		repo_name = repo["name"].(string)
	}
	if payload, ok := mp[i]["payload"].(map[string]any); ok {
		action = payload["action"].(string)
		if pr, ok := payload["pull_request"].(map[string]any); ok {
			title = pr["title"].(string)
			if pr["body"] != nil {
				body = pr["body"].(string)
			}
		}
	}
	if body == "" {
		msg[j].Message = fmt.Sprintf("%s %s pull request in repo \"%s\". Title of PR: \"%s\"", login, action, repo_name, title)
	} else {
		msg[j].Message = fmt.Sprintf("%s %s pull request in repo \"%s\". Title of PR: \"%s\", body of PR: \"%s\".", login, action, repo_name, title, body)
	}
	j++
	return j
}
