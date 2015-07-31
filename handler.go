package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// HandleSentiment takes in a POST with JSON
// containing a 'text' param. The handler runs
// sentiment analysis on the text and returns
// the results.
func HandleSentiment(r http.ResponseWriter, req *http.Request) {
	r.Header().Add("Content-Type", "application/json")

	if req.ContentLength < 1 {
		r.WriteHeader(http.StatusBadRequest)
		r.Write([]byte(fmt.Sprintf(`{"message": "no text passed. Cannot run sentiment analysis"}`)))
		return
	}

	data := make([]byte, req.ContentLength)
	_, err := req.Body.Read(data)
	if err != nil && err != io.EOF {
		r.WriteHeader(http.StatusInternalServerError)
		r.Write([]byte(fmt.Sprintf(`{"message": "ERROR: error reading request body", "error": "%v"}`, err.Error())))
		return
	}

	j := AnalyseJSON{}
	err = json.Unmarshal(data, &j)
	if err != nil {
		r.WriteHeader(http.StatusBadRequest)
		r.Write([]byte(fmt.Sprintf(`{"message": "ERROR: error unmarshalling given JSON into expected format", "error": "%v"}`, err.Error())))
		return
	}

	analysis := model.SentimentAnalysis(j.Text)
	resp, err := json.Marshal(analysis)
	if err != nil {
		r.WriteHeader(http.StatusInternalServerError)
		r.Write([]byte(fmt.Sprintf(`{"message": "ERROR: unable to marshal sentiment analysis into JSON", "error": "%v"}`, err.Error())))
		return
	}

	r.WriteHeader(http.StatusOK)
	r.Write(resp)

	log.Printf("POST /analyze [len(text) = %v]", len(j.Text))
}

// HandleHookedRequest is an http.HandlerFunc
// which will take a POST request with some id
// string and a hook_id string, post a GET request
// to the hook number with the parameters provided
// in the configuration, and run sentiment analysis
// on the returned item (it's expected to have a 'text'
// param in the JSON body)
func HandleHookedRequest(r http.ResponseWriter, req *http.Request) {

}
