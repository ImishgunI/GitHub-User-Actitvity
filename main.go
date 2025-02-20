package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var mp []map[string]any

func main() {
	username := getUsername()
	data := jsonHandle(username)
	parseToMap(data)
	if payload, ok := mp[0]["payload"].(map[string]any); ok {
		fmt.Print(payload["ref_type"])
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
