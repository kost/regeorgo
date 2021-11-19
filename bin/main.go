package main

import (
	"net/http"
	"github.com/kost/regeorgo"
)

func main() {
	// initialize regeorgo
	gh := &regeorgo.GeorgHandler{LogLevel: 0}
	gh.initHandler()

	// use it as standard handler for http
	http.HandleFunc("/regeorgo", gh.regHandler)
	http.ListenAndServe(":8111", nil)
}

