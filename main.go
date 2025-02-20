package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	var (
		username string = getUsername()
	)
	fmt.Print(username)
}

func getUsername() string {
	if len(os.Args) < 2 {
		log.Fatal("You must write a username")
	}
	return os.Args[1]
}
