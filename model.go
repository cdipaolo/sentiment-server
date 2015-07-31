package main

// AnalyzeJSON holds the expected JSON
// request info for the POST /analyze
// endpoint
type AnalyzeJSON struct {
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

// Hook holds information for any
// hooked requests the consumer might
// want to make to the POST /task endpoint.
//
// When calling POST /task the user passes
// an ID that is fmt.Sprintf'ed into the URL.
// As such, the URL in this struct must have
// some sort of %v type string formattable
// value!
type Hook struct {
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers,omitempty"`

	// Key is the key the URL will look for
	// in returned JSON. If not provided the
	// Hook will expect the returned values
	// to be plain text
	Key string `json:"key,omitempty"`
}
