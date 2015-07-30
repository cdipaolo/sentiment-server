package main

import (
	"log"
	"net/http"

	"github.com/cdipaolo/sentiment"
)

var (
	model *sentiment.Model
)

func init() {
	model = sentiment.Train("/tmp/.sentiment")
}

func main() {
	flag.Parse()
	err := ParseConfig()
	if err != nil {
		panic(fmt.Sprintf("ERROR: error parsing configuration!\n\t%v\n", err.Error()))
	}

	http.Handle("/analyze", Post(HandleSentiment))
	http.Handle("/task", Post(HandleHookedRequest))

	log.Fatal(http.ListenAndServe(Config.portString, nil))
}
