package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type Event struct {
	Message string
}

type EventSort struct {
	Push    bool
	Create  bool
	Watch   bool
	Pr      bool
	Delete  bool
	Fork    bool
	Issue   bool
	Release bool
}

var (
	mp  []map[string]any
	rdb *redis.Client
)

func main() {
	username := getUsername()
	data := jsonHandle(username)
	parseToMap(data)
	msg := make([]Event, len(mp))
	var s EventSort
	t := s.getSort()
	if t == nil {
		j := 0
		printMessage(msg, j)
	} else {
		printBySort(msg, t)
	}
}

func chooseEvent(msg *[]Event, j *int) {
	for i := range mp {
		switch mp[i]["type"] {
		case "PushEvent":
			*j = PushEvent(*msg, *j, i)
		case "CreateEvent":
			*j = CreateEvent(*msg, *j, i)
		case "WatchEvent":
			*j = WatchEvent(*msg, *j, i)
		case "PullRequestEvent":
			*j = PullRequestEvent(*msg, *j, i)
		case "DeleteEvent":
			*j = DeleteEvent(*msg, *j, i)
		case "ForkEvent":
			*j = ForkEvent(*msg, *j, i)
		case "IssueEvent":
			*j = IssueEvent(*msg, *j, i)
		case "ReleaseEvent":
			*j = ReleaseEvent(*msg, *j, i)
		}

	}
}

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func printBySort(msg []Event, t *EventSort) {

}

func (s *EventSort) getSort() *EventSort {
	if len(os.Args) > 2 && len(os.Args) < 4 {
		if os.Args[2] == "sort_by" {
			switch os.Args[3] {
			case "push_event":
				s.Push = true
			case "create_event":
				s.Create = true
			case "watch_event":
				s.Watch = true
			case "PR_event":
				s.Pr = true
			case "delete_event":
				s.Delete = true
			case "fork_event":
				s.Fork = true
			case "issue_event":
				s.Issue = true
			case "release_event":
				s.Release = true
			}
		} else {
			log.Fatal("Needs to write sort_by")
		}
	} else {
		return nil
	}
	return s
}

func getUsername() string {
	if len(os.Args) < 2 {
		log.Fatal("You must write a username")
	}
	return os.Args[1]
}

func printMessage(msg []Event, j int) {
	chooseEvent(&msg, &j)
	for _, v := range msg {
		if v.Message != "" {
			fmt.Println(v.Message)
		}
	}
}

func jsonHandle(username string) []byte {
	url := fmt.Sprintf("https://api.github.com/users/%s/events", username)
	ctx := context.Background()
	cachedData, err := rdb.Get(ctx, url).Result()
	if err == nil {
		return []byte(cachedData)
	}
	client := &http.Client{}
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

	err = rdb.Set(ctx, url, data, 600*time.Second).Err()
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

func DeleteEvent(msg []Event, j, i int) int {
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
		ref = payload["ref"].(string)
		ref_type = payload["ref_type"].(string)
	}
	msg[j].Message = fmt.Sprintf("%s delete %s: \"%s\" in repo \"%s\"", login, ref_type, ref, repo_name)
	j++
	return j
}

func ForkEvent(msg []Event, j, i int) int {
	var (
		login     string
		repo_name string
		fork_name string
	)
	if actor, ok := mp[i]["actor"].(map[string]any); ok {
		login = actor["login"].(string)
	}
	if repo, ok := mp[i]["repo"].(map[string]any); ok {
		repo_name = repo["name"].(string)
	}
	if payload, ok := mp[i]["payload"].(map[string]any); ok {
		if forkee, ok := payload["forkee"].(map[string]any); ok {
			fork_name = forkee["full_name"].(string)
		}
	}
	msg[j].Message = fmt.Sprintf("%s forked a repo: \"%s\", fork_name: \"%s\"", login, repo_name, fork_name)
	j++
	return j
}

func IssueEvent(msg []Event, j, i int) int {
	var (
		login     string
		repo_name string
		action    string
		title     string
	)
	if actor, ok := mp[i]["actor"].(map[string]any); ok {
		login = actor["login"].(string)
	}
	if repo, ok := mp[i]["repo"].(map[string]any); ok {
		repo_name = repo["name"].(string)
	}
	if payload, ok := mp[i]["payload"].(map[string]any); ok {
		action = payload["action"].(string)
		if issue, ok := payload["issue"].(map[string]any); ok {
			title = issue["title"].(string)
		}
	}
	msg[j].Message = fmt.Sprintf("%s %s issue in repo \"%s\", issue title: \"%s\"", login, action, repo_name, title)
	j++
	return j
}

func ReleaseEvent(msg []Event, j, i int) int {
	var (
		login     string
		repo_name string
		action    string
		tag_name  string
		version   string
	)
	if actor, ok := mp[i]["actor"].(map[string]any); ok {
		login = actor["login"].(string)
	}
	if repo, ok := mp[i]["repo"].(map[string]any); ok {
		repo_name = repo["name"].(string)
	}
	if payload, ok := mp[i]["payload"].(map[string]any); ok {
		action = payload["action"].(string)
		if release, ok := payload["issue"].(map[string]any); ok {
			tag_name = release["name"].(string)
			version = release["tag_name"].(string)
		}
	}
	msg[j].Message = fmt.Sprintf("%s %s \"%s\" with tag: \"%s %s\"", login, action, repo_name, tag_name, version)
	j++
	return j
}
