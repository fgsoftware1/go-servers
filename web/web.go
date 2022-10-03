package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./www"))
	http.Handle("/", fs)

	log.Print("Listening on 80")

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal(err)
	}
}
