## Sentiment Server
### Web Server For Performing Sentiment Analysis

[![wercker status](https://app.wercker.com/status/8383bada36ee32ed19d7d635a2cc40ac/s "wercker status")](https://app.wercker.com/project/bykey/8383bada36ee32ed19d7d635a2cc40ac)

Sentiment Server performs modular sentiment analysis as a drop-in, easy, open source solution. Getting responses is as easy as `POST /analysis`. The cool part is that you can add in hooks to APIs upon which you can make abbreviated requests. See [hooks](#hooks) for details.

The server uses [this library](http://github.com/cdipaolo/sentiment) for sentiment analysis. Problems with the sentiment engine itself should be registered there. The model is a Naive Bayes classifier trained on the training set from the IMDB movie review corpus (about 85,000 words!)

The server is _fast_! A simple benchmark of the `POST /analyze` endpoint (run `go test -bench .` in the project dir) gives an average time at the server of 1.227ms, including routing, calculating sentiment, etc. for a few paragraphs from a Paul Graham essay with individual analysis for words and sentences as well as the document as a whole on a 2014 Macbook Air with iTunes, Chrome, a terminal, and a bunch of daemons (including Postgres) running. These are all the directly imported dependencies:

![Sentiment Server Dependencies](dependencies.png)

This is a more legit benchmark of the analyze endpoint (`POST /analyze`; same as above mini-bench) endpoint, using [wrk](https://github.com/wg/wrk):

Note that this is analyzing, again, two paragraphs of a Paul Graham essay. It is giving total document sentiment, individual sentence sentiment, and individual word sentiment. Also, actually logging that many requests to STDOUT is not trivial with that much throughput. I was running the same stuff mentioned above.

``` bash
$ wrk -t12 -c400 -d5m -s post.lua http://127.0.0.1:8080/analyze

Running 5m test @ http://127.0.0.1:8080/analyze
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   230.87ms   91.25ms 844.22ms   74.60%
    Req/Sec   146.41    109.14   485.00     63.89%
  514583 requests in 5.00m, 1.81GB read
  Socket errors: connect 0, read 272, write 13, timeout 0
Requests/sec:   1714.69
Transfer/sec:      6.17MB
```

### Installation

```bash
$ go get github.com/cdipaolo/sentiment-server
$ go install github.com/cdipaolo/sentiment-server

# assuming $GOPATH/bin is in your $PATH
$ sentiment-server -C=/path/to/my/configuration
2015/07/31 00:30:45 Listening at http://127.0.0.1:8080 ...

# you can also give a HTTP or HTTPS URL
# of a config file (just plain GET, no auth)
# and it'll use that as the config
#
# note that you can also use -conf instead of -C
$ sentiment-server -conf=http://config.io/my/config.json
```

<a id="hooks"></a>
### Hooks

Hooks let you specify URLs which you can GET from to retrieve text for sentiment analysis. This could be, for example, resources on a company server which need special headers (for auth, etc,) or any other service you don't want to deal with munging through requests every time you make a request.

Hooks are pretty simple. They hold a URL which can be formatted with Golang's `fmt.Sprintf` (basically just have one `%v` in there... eg `http://jsonplaceholder.typicode.com/posts/%v`,) any special headers you need to pass (headers are a `map[string][]string`,) and a key you want to identify the hook with when you request the `POST /task` endpoint.

Note that if you don't specify a key, the response will be assumed to be in plain text format (the resp.Body will be the analyzed text.)

If you want, you may specify a default header so you don't need to tell the API which hook you want each time. This is dont in the config file. If you only specify one hook this will default to be the hook given.

**Time Series Data**

You can have time series hooks which will let you parse data and return it with a format designed to work with time series requests (for example if you are transcoding audio or something.) You just need to add a param to the hook with `"time":true` which will expect the data from the expected key to be in the specified format. 

If you want to know more about this option read the comments on the `Time bool` param of the [`model.go` file](model.go). They are very elaborate and would clog up the README so I'm abstracting them to there.

### Config

Example Config:

```json
{
    "port": 8080,
    "hooks": {
        "post": {
            "url": "http://jsonplaceholder.typicode.com/posts/%v",
            "key": "body",
            "headers": {
                "Auth": ["abcdefg"],
                "Another-Header": ["Hello!"]
            }
        },
        "comment": {
            "url": "http://jsonplaceholder.typicode.com/comments/%v",
            "lang": "en"
        }
    },
    "defaultHook": "comment"
}
```

## Endpoints

### POST /analyze

General text classification. Pass it some body of text in the expected format and it will output the estimated sentiment. Sentiment values are returned on the range [0,1]. For Individual words, the score is the probability that the word is positive. For sentences and the score of the whole document, the value is returned as a descrete value in {0,1}. This is to prevent float underflow by using logarithmic sums (which predict the same output but won't give a clean probability number.) 

Note that all text is converted to lowercase and only letters in a-z are kept (numbers, etc. are taken out.)

Not giving a language will default it to English. Languages must be implemented in [the engine](https://github.com/cdipaolo/sentiment), else they will default to English as well.

**Expected JSON**

```json
{
    "text": "I'm not sure I like your tone right now. I do love you as a person, though.",
    "lang": "en"
}
```

**Returned JSON**

```json
{
  "words": [
    {
      "word": "im",
      "score": 0
    },
    {
      "word": "not",
      "score": 0
    },
    {
      "word": "sure",
      "score": 1
    },
    {
      "word": "i",
      "score": 0
    },
    {
      "word": "like",
      "score": 1
    },
    {
      "word": "your",
      "score": 1
    },
    {
      "word": "tone",
      "score": 1
    },
    {
      "word": "right",
      "score": 1
    },
    {
      "word": "now",
      "score": 1
    },
    {
      "word": "i",
      "score": 0
    },
    {
      "word": "do",
      "score": 1
    },
    {
      "word": "love",
      "score": 1
    },
    {
      "word": "you",
      "score": 0
    },
    {
      "word": "as",
      "score": 1
    },
    {
      "word": "a",
      "score": 0
    },
    {
      "word": "person",
      "score": 0
    },
    {
      "word": "though",
      "score": 1
    }
  ],
  "sentences": [
    {
      "sentence": "im not sure i like your tone right now",
      "score": 0
    },
    {
      "sentence": " i do love you as a person though",
      "score": 1
    }
  ],
  "score": 1
}
```

### POST /task

This calls GET requests to the configured hooks. It allows you to specify the filler id number (called `recordingId` for legacy reasons) which will be formatted into the [hook's](#hooks) URL. It will then return the analysis (same response structure as `POST /analyze`) of the text returned from the request.

Note that you can omit the `hookId` to just use the default hook instead.

**Expected JSON**

```json
{
    "recordingId": "17",
    "hookId": "comments"
}
```

**Returned JSON**

```json
{
  "words": [
    {
      "word": "this",
      "score": 0
    },
    {
      "word": "is",
      "score": 1
    },
    {
      "word": "an",
      "score": 1
    },
    {
      "word": "awesome",
      "score": 1
    },
    {
      "word": "day",
      "score": 1
    }
  ],
  "score": 1
}
```

###GET /

`GET /` is just a health check endpoint. It returns 'Up' as a status if all is ok (which should be any time it can be called,) as well as the total number of successful analyses (apparently that's the plural of 'analysis') and the total number of successful hooked analyses (which is a subset of the former number.)

**Returned JSON**

```json
{
    "status": "Up",
    "totalSuccessfulAnalyses": 666,
    "hookedRequests": 537
}
```

## LICENSE - MIT

See [LICENSE](LICENSE)
