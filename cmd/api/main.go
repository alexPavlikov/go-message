package main

import (
	"log"

	server "github.com/alexPavlikov/go-message/cmd"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
