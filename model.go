package main

import (
	"github.com/cdipaolo/sentiment"
)

// TimeSeriesResponse returns the
// normal analysis response within
// a "metadata" key, as well as
// the analysis of the time series
// data within a key called "series"
type TimeSeriesResponse struct {
	Metadata *sentiment.Analysis `json:"metadata,omitempty"`
	Series   []TimeSeries        `json:"series"`
}

// AnalyzeJSON holds the expected JSON
// request info for the POST /analyze
// endpoint
type AnalyzeJSON struct {
	Text     string             `json:"text"`
	Language sentiment.Language `json:"lang,omitempty"`
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

	// Lanugage holds the language code expected
	// to come from the hook. Defaults to 'en'
	// for English. Look at the available codes
	// at:
	//     https://github.com/cdipaolo/sentiment
	// in its model.go file
	Language sentiment.Language `json:"lang,omitempty"`

	// Key is the key the URL will look for
	// in returned JSON. If not provided the
	// Hook will expect the returned values
	// to be plain text. Following format
	// expected:
	//
	// {
	//   "key": "this is my text to be analyzed!"
	// }
	Key string `json:"key,omitempty"`

	// Time tells the hook request that the
	// hook will return text within time
	// buckets. This expects the following
	// format for a response:
	//
	// {
	//   "key": [
	//      {
	//        "start": 0,
	//        "end": 16.016,
	//        "text": "This is some great text!"
	//      },
	//      {
	//        "start": 16.016,
	//        "end": 24.014,
	//        "text": "I really hate this sentence though..."
	//      }
	//   ]
	//   ...
	// }
	//
	// If this is given the format returned
	// to the API consumer will be the same,
	// but also add in a "timed" section
	// which maps each time bucket to the
	// corresponding text within it to the
	// sentiment of the text within that
	// bucket. All the normal analysis will
	// be moved into a "metadata" parameter.
	// Example:
	//
	// {
	//   "series": [
	//     {
	//       "start": 0,
	//       "end": 16.016,
	//       "text": "This is some great text!",
	//       "score": 1
	//     },
	//     {
	//       "start": 16.016,
	//       "end": 24.014,
	//       "text": "I really hate this sentence though...",
	//       "score": 0
	//     }
	//   ]
	//   "metadata": ...
	// }
	//
	// If this flag is passed and there is
	// no Key given, the hook will be expected
	// to return an array of TimeSeries as the
	// top level JSON object.
	Time bool `json:"time,omitempty"`
}

// TimeSeries holds the expected format
// for time series data response. Look at
// the Hook docs for Time to get a sense
// of how this works.
type TimeSeries struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`

	Text  string `json:"text"`
	Score uint8  `json:"score"`
}

// TimeSeriesRequest holds the expected
// format for temporal data _returned
// from a GET hook_
type TimeSeriesRequest struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}
