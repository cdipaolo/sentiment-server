package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

var client *http.Client

func init() {
	client = &http.Client{}
}

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

	j := TaskJSON{}
	err = json.Unmarshal(data, &j)
	if err != nil {
		r.WriteHeader(http.StatusBadRequest)
		r.Write([]byte(fmt.Sprintf(`{"message": "ERROR: error unmarshalling given JSON into expected format", "error": "%v"}`, err.Error())))
		return
	}

	// * Perform the GET hook * //
	text, err := GetHookResponse(j)
	if err != nil {
		r.WriteHeader(http.StatusInternalServerError)
		r.Write([]byte(fmt.Sprintf(`{"message": "ERROR: unable to get text from hook request with configured parameters", "error": %v}`, err)))
	}

	analysis := model.SentimentAnalysis(text)
	resp, err := json.Marshal(analysis)
	if err != nil {
		r.WriteHeader(http.StatusInternalServerError)
		r.Write([]byte(fmt.Sprintf(`{"message": "ERROR: unable to marshal sentiment analysis into JSON", "error": "%v"}`, err.Error())))
		return
	}

	r.WriteHeader(http.StatusOK)
	r.Write(resp)

	log.Printf("POST /task [len(text) = %v]", len(text))
}

// GetHookResponse takes in a TaskJSON and
// returns the text of the response given by
// the hook. The text is found by returning
// the JSON field from the param specified
// within the hook declaration (and expecting
// plain text result if the param is blank
func GetHookResponse(j TaskJSON) (string, error) {
	id := Config.DefaultHook
	if j.HookID != "" {
		id = j.HookID
	}

	hook, ok := Config.Hooks[id]
	if !ok {
		return "", fmt.Errorf(`{"message": "ERROR: hook given was not found in your configured hooks!", "hookId": "%v", "defaultHook": "%v"}`, id, Config.DefaultHook)
	}

	url, err := url.Parse(fmt.Sprintf(hook.URL, j.ID))
	if err != nil {
		return "", fmt.Errorf(`{"message": "ERROR: unable to format the given ID into the configured hook ID", "hookUrl": "%v", "id":"%v"}`, hook.URL, id)
	}

	request := &http.Request{
		Method: "GET",
		Header: http.Header(hook.Headers),
		URL:    url,
	}

	resp, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf(`{"message": "ERROR: could not complete HOOK request", "hook": "%v", "error": "%v"}`, id, err)
	}

	data := make([]byte, resp.ContentLength)
	n, err := resp.Body.Read(data)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf(`{"message": "ERROR: could not read the body from HOOK GET request", "hook": "%v", "error": "%v"}`, id, err)
	}

	text := string(data)
	if hook.Key != "" {
		bod := make(map[string]interface{})
		err = json.Unmarshal(data[:n], &bod)
		if err != nil {
			return "", fmt.Errorf(`{"message": "ERROR: could not unmarshal body from HOOK GET request", "hook": "%v", "error": "%v"}`, id, err)
		}

		tmp, ok := bod[hook.Key]
		if !ok {
			return "", fmt.Errorf(`{"message": "ERROR: could not get text with the given key from HOOK GET request", "hook": "%v", "expectedId": "%v"}`, id, hook.Key)
		}

		text, ok = tmp.(string)
		if !ok {
			return "", fmt.Errorf(`{"message": "ERROR: could not assert HOOK GET request body to type string", "hook": "%v"}`, id)
		}
	}

	return text, nil
}
