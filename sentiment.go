package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/cdipaolo/sentiment"
)

var (
	model sentiment.Models
)

func init() {
	var err error
	model, err = sentiment.Restore()
	if err != nil {
		panic(fmt.Sprintf("ERROR: error restoring sentiment model!\n\t%v\n", err))
	}

	http.Handle("/analyze", Post(HandleSentiment))
	http.Handle("/task", Post(HandleHookedRequest))
	http.Handle("/", Get(HandleStatus))
}

func main() {
	flag.Parse()
	err := ParseConfig()
	if err != nil {
		panic(fmt.Sprintf("ERROR: error parsing configuration!\n\t%v\n", err.Error()))
	}

	log.Printf("Listening at http://127.0.0.1%v ...\n", Config.portString)
	log.Fatal(http.ListenAndServe(Config.portString, nil))
}
