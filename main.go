package main

import (
	"github.com/codecodify/randimages/router"
	"log"
)

func main() {
	r := router.Start()
	log.Panic(r.Run(":8080"))
}
