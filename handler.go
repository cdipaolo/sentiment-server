package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

var (
	client    *http.Client
	count     int64
	hookCount int64
)

func init() {
	client = &http.Client{}
	count = 0
	hookCount = 0
}

// HandleStatus is a simple health-check endpoint
// that will tell the user the total number of
// successful analyses and the number of successful
// hooked requests (a subset of the former) made
func HandleStatus(r http.ResponseWriter, req *http.Request) {
	r.Header().Add("Content-Type", "application/json")
	r.WriteHeader(http.StatusOK)

	// send the total successful count
	// and total error count
	r.Write([]byte(fmt.Sprintf(`{
		"status": "Up",
		"totalSuccessfulAnalyses": %v,
		"hookedRequests": %v
	}`, count, hookCount)))

	log.Printf("GET / [totalSuccessfulAnalyses = %v]\n", count)
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
		log.Printf("POST /analyze > ERROR: no text passed\n")
		return
	}

	data := make([]byte, req.ContentLength)
	_, err := req.Body.Read(data)
	if err != nil && err != io.EOF {
		r.WriteHeader(http.StatusInternalServerError)
		r.Write([]byte(fmt.Sprintf(`{"message": "ERROR: error reading request body", "error": "%v"}`, err.Error())))
		log.Printf("POST /analyze > ERROR: couldn't read request body\n\t%v\n", err)
		return
	}

	j := AnalyzeJSON{}
	err = json.Unmarshal(data, &j)
	if err != nil {
		r.WriteHeader(http.StatusBadRequest)
		r.Write([]byte(fmt.Sprintf(`{"message": "ERROR: error unmarshalling given JSON into expected format", "error": "%v"}`, err.Error())))
		log.Printf("POST /analyze > ERROR: error unmarshalling given JSON\n\t%v\n", err)
		return
	}

	analysis := model.SentimentAnalysis(j.Text)
	resp, err := json.Marshal(analysis)
	if err != nil {
		r.WriteHeader(http.StatusInternalServerError)
		r.Write([]byte(fmt.Sprintf(`{"message": "ERROR: unable to marshal sentiment analysis into JSON", "error": "%v"}`, err.Error())))
		log.Printf("POST /analyze > ERROR: unable to unmarshal sentiment analysis into JSON\n\t%v\n", err)
		return
	}

	r.WriteHeader(http.StatusOK)
	r.Write(resp)

	count++
	log.Printf("POST /analyze [len(text) = %v]\n", len(j.Text))
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
		log.Printf("POST /task > ERROR: no text passed\n")
		return
	}

	data := make([]byte, req.ContentLength)
	_, err := req.Body.Read(data)
	if err != nil && err != io.EOF {
		r.WriteHeader(http.StatusInternalServerError)
		r.Write([]byte(fmt.Sprintf(`{"message": "ERROR: error reading request body", "error": "%v"}`, err.Error())))
		log.Printf("POST /task > ERROR: error reading request\n\t%v\n", err)
		return
	}

	j := TaskJSON{}
	err = json.Unmarshal(data, &j)
	if err != nil {
		r.WriteHeader(http.StatusBadRequest)
		r.Write([]byte(fmt.Sprintf(`{"message": "ERROR: error unmarshalling given JSON into expected format", "error": "%v"}`, err.Error())))
		log.Printf("POST /task > ERROR: error unmarshalling given JSON\n\t%v\n", err)
		return
	}

	// * Perform the GET hook * //
	series, text, err := GetHookResponse(j)
	if err != nil {
		r.WriteHeader(http.StatusInternalServerError)
		r.Write([]byte(fmt.Sprintf(`{"message": "ERROR: unable to get text from hook request with configured parameters", "error": %v}`, err)))
		log.Printf("POST /task > ERROR: error getting hooked response\n\t%v\n", err)
		return
	}

	resp := []byte{}

	analysis := model.SentimentAnalysis(text)

	if series == nil {
		resp, err = json.Marshal(analysis)
	} else {
		for i := range series {
			series[i].Score = model.SentimentOfSentence(series[i].Text)
		}
		resp, err = json.Marshal(TimeSeriesResponse{
			Metadata: analysis,
			Series:   series,
		})
	}

	if err != nil {
		r.WriteHeader(http.StatusInternalServerError)
		r.Write([]byte(fmt.Sprintf(`{"message": "ERROR: unable to marshal sentiment analysis into JSON", "error": "%v"}`, err.Error())))
		log.Printf("POST /task > ERROR: error marshalling sentiment analysis into JSON\n\t%v\n", err)
		return
	}

	r.WriteHeader(http.StatusOK)
	r.Write(resp)

	hookCount++
	count++
	log.Printf("POST /task [len(text) = %v]\n", len(text))
}

// GetHookResponse takes in a TaskJSON and
// returns the text of the response given by
// the hook. The text is found by returning
// the JSON field from the param specified
// within the hook declaration (and expecting
// plain text result if the param is blank
func GetHookResponse(j TaskJSON) ([]TimeSeries, string, error) {
	id := Config.DefaultHook
	if j.HookID != "" {
		id = j.HookID
	}

	hook, ok := Config.Hooks[id]
	if !ok {
		return nil, "", fmt.Errorf(`{"message": "ERROR: hook given was not found in your configured hooks!", "hookId": "%v", "defaultHook": "%v"}`, id, Config.DefaultHook)
	}

	url, err := url.Parse(fmt.Sprintf(hook.URL, j.ID))
	if err != nil {
		return nil, "", fmt.Errorf(`{"message": "ERROR: unable to format the given ID into the configured hook ID", "hookUrl": "%v", "id":"%v"}`, hook.URL, id)
	}

	request := &http.Request{
		Method: "GET",
		Header: http.Header(hook.Headers),
		URL:    url,
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, "", fmt.Errorf(`{"message": "ERROR: could not complete HOOK request", "hook": "%v", "error": "%v"}`, id, err)
	}

	data := make([]byte, resp.ContentLength)
	n, err := resp.Body.Read(data)
	if err != nil && err != io.EOF {
		return nil, "", fmt.Errorf(`{"message": "ERROR: could not read the body from HOOK GET request", "hook": "%v", "error": "%v"}`, id, err)
	}

	text := string(data)
	var timeSeries []TimeSeries
	if hook.Key != "" && !hook.Time {
		bod := make(map[string]interface{})
		err = json.Unmarshal(data[:n], &bod)
		if err != nil {
			return nil, "", fmt.Errorf(`{"message": "ERROR: could not unmarshal body from HOOK GET request", "hook": "%v", "error": "%v"}`, id, err)
		}

		tmp, ok := bod[hook.Key]
		if !ok {
			return nil, "", fmt.Errorf(`{"message": "ERROR: could not get text with the given key from HOOK GET request", "hook": "%v", "expectedId": "%v"}`, id, hook.Key)
		}

		text, ok = tmp.(string)
		if !ok {
			return nil, "", fmt.Errorf(`{"message": "ERROR: could not assert HOOK GET request body to type string", "hook": "%v"}`, id)
		}
	}
	if hook.Key != "" && hook.Time {
		bod := make(map[string]interface{})
		err = json.Unmarshal(data[:n], &bod)
		if err != nil {
			return nil, "", fmt.Errorf(`{"message": "ERROR: could not unmarshal body from HOOK GET request", "hook": "%v", "error": "%v"}`, id, err)
		}

		// have to marshal into JSON and back out
		// of JSON to ignore the extra fields possibly
		// in the interface{} for type assertion
		timeSeriesJSON, err := json.Marshal(bod[hook.Key])
		if err != nil {
			return nil, "", fmt.Errorf(`{"message": "ERROR: could not marshal (what should be an) array with the given key from HOOK GET request", "hook": "%v", "expectedId": "%v", "body": %v}`, id, hook.Key, string(data[:n]))
		}

		// and back out again!
		t := []TimeSeriesRequest{}
		err = json.Unmarshal(timeSeriesJSON, &t)
		if err != nil {
			return nil, "", fmt.Errorf(`{"message": "ERROR: could not unmarshal series with the given key from HOOK GET request", "hook": "%v", "series": %v}`, id, hook.Key, string(timeSeriesJSON))
		}

		timeSeries = []TimeSeries{}
		for i := range t {
			timeSeries = append(timeSeries, TimeSeries{
				Start: t[i].Start * 1000.0,
				End:   t[i].End * 1000.0,
				Text:  t[i].Text,
			})
		}

		text = TurnTimeSeriesIntoText(timeSeries)
	}
	if hook.Key == "" && hook.Time {
		err = json.Unmarshal(data[:n], &timeSeries)
		if err != nil {
			return nil, "", fmt.Errorf(`{"message": "ERROR: could not unmarshal body from HOOK GET request into type []TimeSeries", "hook": "%v", "error": "%v"}`, id, err)
		}

		text = TurnTimeSeriesIntoText(timeSeries)
	}

	return timeSeries, text, nil
}

// TurnTimeSeriesIntoText compiles all the text values
// within an []TimeSeries into one string for use
// with regular analysis
func TurnTimeSeriesIntoText(ts []TimeSeries) string {
	var text bytes.Buffer
	for i := range ts {
		text.WriteString(ts[i].Text)
		text.WriteString(" ")
	}

	return text.String()
}
