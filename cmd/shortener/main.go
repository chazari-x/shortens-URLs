package main

import (
	"log"

	"main/internal/app/server"
)

func main() {
	err := server.StartSever()
	if err != nil {
		log.Fatal(err)
	}
}
