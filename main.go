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
	PushEvent(msg)
	for i := range msg {
		fmt.Println(msg[i].Message)
	}
}

func getUsername() string {
	if len(os.Args) < 2 {
		log.Fatal("You must write a username")
	}
	return os.Args[1]
}

func jsonHandle(username string) []byte {
	client := &http.Client{}
	url := fmt.Sprintf("https://api.github.com/users/%s/events", username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Accept", "application/vnd.github+json")

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

func PushEvent(msg []Event) {
	var (
		login     any
		repo_name any
		message   any
	)
	j := 0
	for i := range mp {
		if mp[i]["type"] == "PushEvent" {
			if actor, ok := mp[i]["actor"].(map[string]any); ok {
				login = actor["login"].(string)
			}
			if repo, ok := mp[i]["repo"].(map[string]any); ok {
				repo_name = repo["name"].(string)
			}
			if payload, ok := mp[i]["payload"].(map[string]any); ok {
				k := 0
				if commits, ok := payload["commits"].([]any); ok && len(commits) > 0 {
					if commitData, ok := commits[k].(map[string]any); ok {
						message, _ = commitData["message"].(string)
					}
					k++
				}
			}
			msg[j].Message = fmt.Sprintf("%s pushed message: \"%s\" to repository: \"%s\"", login, message, repo_name)
			j++
		}
	}
}
