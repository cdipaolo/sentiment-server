package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"testing"

	"github.com/cdipaolo/sentiment"
)

const (
	URL      = "127.0.0.1:8080"
	Protocol = "http://"
)

var (
	TestComment  = []byte(`{"text": "The anti-immigration people have to invent some explanation to account for all the effort technology companies have expended trying to make immigration easier. So they claim it's because they want to drive down salaries. But if you talk to startups, you find practically every one over a certain size has gone through legal contortions to get programmers into the US, where they then paid them the same as they'd have paid an American. Why would they go to extra trouble to get programmers for the same price? The only explanation is that they're telling the truth: there are just not enough great programmers to go around"}`)
	TestPost     = []byte(`The anti-immigration people have to invent some explanation to account for all the effort technology companies have expended trying to make immigration easier. So they claim it's because they want to drive down salaries. But if you talk to startups, you find practically every one over a certain size has gone through legal contortions to get programmers into the US, where they then paid them the same as they'd have paid an American. Why would they go to extra trouble to get programmers for the same price? The only explanation is that they're telling the truth: there are just not enough great programmers to go around`)
	TestTemporal = []byte(`{
		"series": [
			{
				"start": 0,
				"end": 1.123,
				"text": "I am a happy guy!"
			},
			{
				"start": 1.123,
				"end": 12.0162,
				"text": "But not when I am sad :("
			},
			{
				"start": 12.0162,
				"end": 13.12,
				"text": "I think as a whole I have a good life, though..."
			}
		]
	}`)
	TestTemporalArray = []byte(`[
		{
			"start": 0,
			"end": 1.123,
			"text": "I am a happy guy!"
		},
		{
			"start": 1.123,
			"end": 12.0162,
			"text": "But not when I am sad :("
		},
		{
			"start": 12.0162,
			"end": 13.12,
			"text": "I think as a whole I have a good life, though..."
		}
	]`)
)

func init() {
	// create test handlers for hooks
	http.HandleFunc("/test/comment/", func(r http.ResponseWriter, req *http.Request) {
		r.Header().Add("Content-Type", "application/json")

		if len(req.Header["Auth"]) == 0 || req.Header["Auth"][0] != "SUPER_SECRET" {
			r.WriteHeader(http.StatusUnauthorized)
			r.Write([]byte(`{"message": "ERROR: you didn't pass the auth!!!!!!!!!!!!!!!!!!!!!"}`))
			return
		}

		r.WriteHeader(http.StatusOK)
		r.Write(TestComment)
	})

	http.HandleFunc("/test/post/", func(r http.ResponseWriter, req *http.Request) {
		r.Header().Add("Content-Type", "text/plain")

		r.WriteHeader(http.StatusOK)
		r.Write(TestPost)
	})

	http.HandleFunc("/test/temporal/", func(r http.ResponseWriter, req *http.Request) {
		r.Header().Add("Content-Type", "application/json")
		r.WriteHeader(http.StatusOK)

		if len(req.Header["Array"]) != 0 && req.Header["Array"][0] == "true" {
			r.Write(TestTemporalArray)
			return
		}

		r.Write(TestTemporal)
	})

	go main()
}

// post takes a path and a json to post, performs a
// POST request, and returns the status, the body,
// and any errors
func post(pth string, json string) (int, []byte, error) {
	resp, err := http.Post(Protocol+path.Join(URL, pth), "application/json", bytes.NewBuffer([]byte(json)))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, body, nil
}

// get makes a get request to the server
// at the specified path
func get(pth string) (int, []byte, error) {
	resp, err := http.Get(Protocol + path.Join(URL, pth))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, body, nil
}

// * GET / tests * //

func TestStatusShouldPass1(t *testing.T) {
	statusCode, body, err := get("/")
	if err != nil {
		t.Errorf("ERROR: error trying to get\n\t%v\n", err)
	}
	if statusCode != http.StatusOK {
		t.Errorf("ERROR: status returned should be 200 OK\n\t%v\n", string(body))
	}
	if len(body) == 0 {
		t.Fatalf("ERROR: body should not be nil!\n")
	}

	status := struct {
		Status    string `json:"status"`
		Count     int64  `json:"totalSuccessfulAnalyses"`
		HookCount int64  `json:"hookedRequests"`
	}{}
	err = json.Unmarshal(body, &status)
	if err != nil {
		t.Fatalf("ERROR: error unmarshalling JSON response\n\t%v\n", err)
	}

	if status.Status != "Up" {
		t.Errorf("ERROR: health check status should be 'Up'\n\t%+v\n", status)
	}
}

// * POST /analyze tests * //

func TestSentimentShouldPass1(t *testing.T) {
	text := `The anti-immigration people have to invent some explanation to account for all the effort technology companies have expended trying to make immigration easier. So they claim it's because they want to drive down salaries. But if you talk to startups, you find practically every one over a certain size has gone through legal contortions to get programmers into the US, where they then paid them the same as they'd have paid an American. Why would they go to extra trouble to get programmers for the same price? The only explanation is that they're telling the truth: there are just not enough great programmers to go around`
	txt := fmt.Sprintf(`{
		"text": "%v"
	}`, text)

	status, body, err := post("analyze", txt)
	if err != nil {
		t.Errorf("ERROR: error trying to post\n\t%v\n", err)
	}
	if status != http.StatusOK {
		t.Errorf("ERROR: status returned should be 200 OK\n\t%v\n", string(body))
	}
	if len(body) == 0 {
		t.Fatalf("ERROR: body should not be nil!\n")
	}

	analysis := sentiment.Analysis{}
	err = json.Unmarshal(body, &analysis)
	if err != nil {
		t.Fatalf("ERROR: error unmarshalling JSON response\n\t%v\n", err)
	}

	should := model.SentimentAnalysis(text)
	if should.Score != analysis.Score {
		t.Errorf("ERROR: responded text sentiment score should equal the same score from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Score, analysis.Score)
	}
	if len(should.Words) != len(analysis.Words) {
		t.Errorf("ERROR: responded individual word sentiment should equal in length the same response from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Words, analysis.Words)
	}
}

func TestSentimentShouldFail1(t *testing.T) {
	text := `The anti-immigration people have to invent some explanation to account for all the effort technology companies have expended trying to make immigration easier. So they claim it's because they want to drive down salaries. But if you talk to startups, you find practically every one over a certain size has gone through legal contortions to get programmers into the US, where they then paid them the same as they'd have paid an American. Why would they go to extra trouble to get programmers for the same price? The only explanation is that they're telling the truth: there are just not enough great programmers to go around`
	txt := fmt.Sprintf(`{
		"text": "%v
	}`, text)

	status, body, err := post("analyze", txt)
	if err != nil {
		t.Errorf("ERROR: error trying to post\n\t%v\n", err)
	}
	if status != http.StatusBadRequest {
		t.Errorf("ERROR: status returned should be 400 BAD REQUEST\n\t%v\n", string(body))
	}
	if len(body) == 0 {
		t.Fatalf("ERROR: body should not be nil!\n")
	}
}

func TestSentimentShouldFail2(t *testing.T) {
	txt := ``

	status, body, err := post("analyze", txt)
	if err != nil {
		t.Errorf("ERROR: error trying to post\n\t%v\n", err)
	}
	if status != http.StatusBadRequest {
		t.Errorf("ERROR: status returned should be 400 BAD REQUEST\n\t%v\n", string(body))
	}
	if len(body) == 0 {
		t.Fatalf("ERROR: body should not be nil!\n")
	}
}

// * Hooked Requests * //

func TestHookedSentimentShouldPass1(t *testing.T) {
	ts, text, err := GetHookResponse(TaskJSON{
		ID:     "1",
		HookID: "comment",
	})
	if err != nil {
		t.Fatalf("ERROR: could not get hooked response!\n\t%v\n", err)
	}

	status, body, err := post("task", `{
		"recordingId": "1",
		"hookId": "comment"
	}`)
	if err != nil {
		t.Errorf("ERROR: error trying to post\n\t%v\n", err)
	}
	if status != http.StatusOK {
		t.Errorf("ERROR: status returned should be 200 OK\n\t%v\n", string(body))
	}
	if len(body) == 0 {
		t.Fatalf("ERROR: body should not be nil!\n")
	}

	analysis := sentiment.Analysis{}
	err = json.Unmarshal(body, &analysis)
	if err != nil {
		t.Fatalf("ERROR: error unmarshalling JSON response\n\t%v\n", err)
	}

	if ts != nil {
		t.Errorf("ERROR: time series should be nil!\n\t%v\n", ts)
	}

	should := model.SentimentAnalysis(text)
	if should.Score != analysis.Score {
		t.Errorf("ERROR: responded text sentiment score should equal the same score from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Score, analysis.Score)
	}
	if len(should.Words) != len(analysis.Words) {
		t.Errorf("ERROR: responded individual word sentiment should equal in length the same response from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Words, analysis.Words)
	}
}

func TestHookedSentimentShouldPass2(t *testing.T) {
	ts, text, err := GetHookResponse(TaskJSON{
		ID:     "1",
		HookID: "post",
	})
	if err != nil {
		t.Fatalf("ERROR: could not get hooked response!\n\t%v\n", err)
	}

	status, body, err := post("task", `{
		"recordingId": "1",
		"hookId": "post"
	}`)
	if err != nil {
		t.Errorf("ERROR: error trying to post\n\t%v\n", err)
	}
	if status != http.StatusOK {
		t.Errorf("ERROR: status returned should be 200 OK\n\t%v\n", string(body))
	}
	if len(body) == 0 {
		t.Fatalf("ERROR: body should not be nil!\n")
	}

	analysis := sentiment.Analysis{}
	err = json.Unmarshal(body, &analysis)
	if err != nil {
		t.Fatalf("ERROR: error unmarshalling JSON response\n\t%v\n", err)
	}

	if ts != nil {
		t.Errorf("ERROR: time series should be nil!\n\t%v\n", ts)
	}

	should := model.SentimentAnalysis(text)
	if should.Score != analysis.Score {
		t.Errorf("ERROR: responded text sentiment score should equal the same score from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Score, analysis.Score)
	}
	if len(should.Words) != len(analysis.Words) {
		t.Errorf("ERROR: responded individual word sentiment should equal in length the same response from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Words, analysis.Words)
	}
}

// test default hook settings
func TestHookedSentimentShouldPass3(t *testing.T) {
	ts, text, err := GetHookResponse(TaskJSON{
		ID:     "1",
		HookID: "post",
	})
	if err != nil {
		t.Fatalf("ERROR: could not get hooked response!\n\t%v\n", err)
	}

	status, body, err := post("task", `{
		"recordingId": "1"
	}`)
	if err != nil {
		t.Errorf("ERROR: error trying to post\n\t%v\n", err)
	}
	if status != http.StatusOK {
		t.Errorf("ERROR: status returned should be 200 OK\n\t%v\n", string(body))
	}
	if len(body) == 0 {
		t.Fatalf("ERROR: body should not be nil!\n")
	}

	analysis := sentiment.Analysis{}
	err = json.Unmarshal(body, &analysis)
	if err != nil {
		t.Fatalf("ERROR: error unmarshalling JSON response\n\t%v\n", err)
	}

	if ts != nil {
		t.Errorf("ERROR: time series should be nil!\n\t%v\n", ts)
	}

	should := model.SentimentAnalysis(text)
	if should.Score != analysis.Score {
		t.Errorf("ERROR: responded text sentiment score should equal the same score from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Score, analysis.Score)
	}
	if len(should.Words) != len(analysis.Words) {
		t.Errorf("ERROR: responded individual word sentiment should equal in length the same response from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Words, analysis.Words)
	}
}

// test temporal data handling

func TestHookedSentimentShouldPass4(t *testing.T) {
	ts, text, err := GetHookResponse(TaskJSON{
		ID:     "1",
		HookID: "temporal",
	})
	if err != nil {
		t.Fatalf("ERROR: could not get hooked response!\n\t%v\n", err)
	}

	status, body, err := post("task", `{
		"recordingId": "1",
		"hookId": "temporal"
	}`)
	if err != nil {
		t.Errorf("ERROR: error trying to post\n\t%v\n", err)
	}
	if status != http.StatusOK {
		t.Errorf("ERROR: status returned should be 200 OK\n\t%v\n", string(body))
	}
	if len(body) == 0 {
		t.Fatalf("ERROR: body should not be nil!\n")
	}

	analysis := TimeSeriesResponse{}
	err = json.Unmarshal(body, &analysis)
	if err != nil {
		t.Fatalf("ERROR: error unmarshalling JSON response\n\t%v\n", err)
	}

	if ts == nil {
		t.Errorf("ERROR: time series from hook should not be nil!\n\t%v\n", ts)
	}
	if analysis.Series == nil {
		t.Fatalf("ERROR: time series from response should not be nil!\n\t%v\n", analysis.Series)
	}
	if analysis.Metadata == nil {
		t.Fatalf("ERROR: analysis metadata from response should not be nil!\n\t%v\n", analysis.Metadata)
	}

	should := model.SentimentAnalysis(text)
	if should.Score != analysis.Metadata.Score {
		t.Errorf("ERROR: responded text sentiment score should equal the same score from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Score, analysis.Metadata.Score)
	}
	if len(should.Words) != len(analysis.Metadata.Words) {
		t.Errorf("ERROR: responded individual word sentiment should equal in length the same response from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Words, analysis.Metadata.Words)
	}
	if len(analysis.Series) != len(ts) {
		t.Errorf("ERROR: responded time series data should equal, in length, the time series data given from the request!\n\tShould be %v\n\tReturned: %v\n", len(ts), len(analysis.Series))
	}
}

func TestHookedSentimentShouldPass5(t *testing.T) {
	ts, text, err := GetHookResponse(TaskJSON{
		ID:     "1",
		HookID: "temporalArray",
	})
	if err != nil {
		t.Fatalf("ERROR: could not get hooked response!\n\t%v\n", err)
	}

	status, body, err := post("task", `{
		"recordingId": "1",
		"hookId": "temporalArray"
	}`)
	if err != nil {
		t.Errorf("ERROR: error trying to post\n\t%v\n", err)
	}
	if status != http.StatusOK {
		t.Errorf("ERROR: status returned should be 200 OK\n\t%v\n", string(body))
	}
	if len(body) == 0 {
		t.Fatalf("ERROR: body should not be nil!\n")
	}

	analysis := TimeSeriesResponse{}
	err = json.Unmarshal(body, &analysis)
	if err != nil {
		t.Fatalf("ERROR: error unmarshalling JSON response\n\t%v\n", err)
	}

	if ts == nil {
		t.Errorf("ERROR: time series from hook should not be nil!\n\t%v\n", ts)
	}
	if analysis.Series == nil {
		t.Fatalf("ERROR: time series from response should not be nil!\n\t%v\n", analysis.Series)
	}
	if analysis.Metadata == nil {
		t.Fatalf("ERROR: analysis metadata from response should not be nil!\n\t%v\n", analysis.Metadata)
	}

	should := model.SentimentAnalysis(text)
	if should.Score != analysis.Metadata.Score {
		t.Errorf("ERROR: responded text sentiment score should equal the same score from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Score, analysis.Metadata.Score)
	}
	if len(should.Words) != len(analysis.Metadata.Words) {
		t.Errorf("ERROR: responded individual word sentiment should equal in length the same response from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Words, analysis.Metadata.Words)
	}
	if len(analysis.Series) != len(ts) {
		t.Errorf("ERROR: responded time series data should equal, in length, the time series data given from the request!\n\tShould be %v\n\tReturned: %v\n", len(ts), len(analysis.Series))
	}
}

func TestHookedSentimentShouldFail1(t *testing.T) {
	status, body, err := post("task", `{
		"recordingId": "1",
		"hookId": "does-not-exist"
	}`)
	if err != nil {
		t.Errorf("ERROR: error trying to post\n\t%v\n", err)
	}
	if status != http.StatusInternalServerError {
		t.Errorf("ERROR: status returned should be 500 SERVER ERROR\n\t%v\n", string(body))
	}
	if len(body) == 0 {
		t.Fatalf("ERROR: body should not be nil!\n")
	}
}

func TestHookedSentimentShouldFail2(t *testing.T) {
	status, body, err := post("task", ``)
	if err != nil {
		t.Errorf("ERROR: error trying to post\n\t%v\n", err)
	}
	if status != http.StatusBadRequest {
		t.Errorf("ERROR: status returned should be 400 BAD REQUEST\n\t%v\n", string(body))
	}
	if len(body) == 0 {
		t.Fatalf("ERROR: body should not be nil!\n")
	}
}

// * Benchmarks * //

func BenchmarkPOSTAnalyze(b *testing.B) {
	txt := `{"text":"The anti-immigration people have to invent some explanation to account for all the effort technology companies have expended trying to make immigration easier. So they claim it's because they want to drive down salaries. But if you talk to startups, you find practically every one over a certain size has gone through legal contortions to get programmers into the US, where they then paid them the same as they'd have paid an American. Why would they go to extra trouble to get programmers for the same price? The only explanation is that they're telling the truth: there are just not enough great programmers to go around"}`

	for i := 0; i < b.N; i++ {
		post("analyze", txt)
	}
}
