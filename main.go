package main

import (
	"fmt"
	"log"
	"os"
)

type Event struct {
	Id    int    `json:"id"`
	Type  string `json:"type"`
	Actor struct {
		Id            int    `json:"id"`
		Login         string `json:"login"`
		Display_login string `json:"display_login"`
	}
}

func main() {
	username := getUsername()
	fmt.Print(username)
}

func getUsername() string {
	if len(os.Args) < 2 {
		log.Fatal("You must write a username")
	}
	return os.Args[1]
}
