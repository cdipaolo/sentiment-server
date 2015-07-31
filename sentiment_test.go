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

func init() {
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
	text, err := GetHookResponse(TaskJSON{
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

	should := model.SentimentAnalysis(text)
	if should.Score != analysis.Score {
		t.Errorf("ERROR: responded text sentiment score should equal the same score from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Score, analysis.Score)
	}
	if len(should.Words) != len(analysis.Words) {
		t.Errorf("ERROR: responded individual word sentiment should equal in length the same response from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Words, analysis.Words)
	}
}

func TestHookedSentimentShouldPass2(t *testing.T) {
	text, err := GetHookResponse(TaskJSON{
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

	should := model.SentimentAnalysis(text)
	if should.Score != analysis.Score {
		t.Errorf("ERROR: responded text sentiment score should equal the same score from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Score, analysis.Score)
	}
	if len(should.Words) != len(analysis.Words) {
		t.Errorf("ERROR: responded individual word sentiment should equal in length the same response from the library!\n\tShould be: %v\n\tReturned: %v\n", should.Words, analysis.Words)
	}
}

// * Benchmarks * //

func BenchmarkPOSTAnalyze(b *testing.B) {
	txt := `{"text":"The anti-immigration people have to invent some explanation to account for all the effort technology companies have expended trying to make immigration easier. So they claim it's because they want to drive down salaries. But if you talk to startups, you find practically every one over a certain size has gone through legal contortions to get programmers into the US, where they then paid them the same as they'd have paid an American. Why would they go to extra trouble to get programmers for the same price? The only explanation is that they're telling the truth: there are just not enough great programmers to go around"}`

	for i := 0; i < b.N; i++ {
		post("analyze", txt)
	}
}
