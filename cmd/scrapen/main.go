package main

import (
	"log"
	"os"

	"github.com/akeil/scrapen"
)

func main() {

	url := os.Args[1]

	err := scrapen.Run(url)
	if err != nil {
		log.Fatal(err)
	}
}
