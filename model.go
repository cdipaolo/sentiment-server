package main

// AnalyzeJSON holds the expected JSON
// request info for the POST /analyze
// endpoint
type AnalyseJSON struct {
	Text string `json:"text"`
}

// TaskJSON holds a generic request
// for the POST /task endpoint, where
// the consumer can set a URL to make
// a GET request from (with the id
// specified within this struct) and
// run analysis on that returned value
type TaskJSON struct {
	ID     string `json:"recordingId"`
	HookID string `json:"hookId,omitempty"`
}
