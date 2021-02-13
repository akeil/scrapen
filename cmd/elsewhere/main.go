package main

import (
	"log"
	"os"

	"akeil.net/akeil/elsewhere"
)

func main() {

	url := os.Args[1]

	err := elsewhere.Run(url)
	if err != nil {
		log.Fatal(err)
	}
}
