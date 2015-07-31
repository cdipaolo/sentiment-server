## Sentiment Server
### Web Server For Performing Sentiment Analysis

Sentiment Server performs modular sentiment analysis as a drop-in, easy, open source solution. Getting responses is as easy as `POST /analysis`. The cool part is that you can add in hooks to APIs upon which you can make abbreviated requests. See [hooks](#hooks) for details.

The server uses [this library](http://github.com/cdipaolo/sentiment) for sentiment analysis. Problems with the sentiment engine itself should be registered there.

The server is _fast_! These are all the directly imported dependencies:

![Sentiment Server Dependencies](dependencies.png)

### Installation

```bash
$ go get github.com/cdipaolo/sentiment-server
$ go install github.com/cdipaolo/sentiment-server

# assuming $GOPATH/bin is in your $PATH
$ sentiment-server -C=/path/to/my/configuration
2015/07/31 00:30:45 Listening at http://127.0.0.1:8080 ...
```

<a id="hooks"></a>
### Hooks

Hooks let you specify URLs which you can GET from to retrieve text for sentiment analysis. This could be, for example, resources on a company server which need special headers (for auth, etc,) or any other service you don't want to deal with munging through requests every time you make a request.

Hooks are pretty simple. They hold a URL which can be formatted with Golang's `fmt.Sprintf` (basically just have one `%v` in there... eg `http://jsonplaceholder.typicode.com/posts/%v`,) any special headers you need to pass (headers are a `map[string][]string`,) and a key you want to identify the hook with when you request the `POST /task` endpoint.

Note that if you don't specify a key, the response will be assumed to be in plain text format (the resp.Body will be the analyzed text.)

If you want, you may specify a default header so you don't need to tell the API which hook you want each time. This is dont in the config file. If you only specify one hook this will default to be the hook given.

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
                "Auth": "abcdefg",
                "Another-Header": "Hello!"
            }
        },
        "comment": {
            "url": "http://jsonplaceholder.typicode.com/comments/%v",
        }
    },
    "defaultHook": "comment"
}
```

## Endpoints

### POST /analyze

General text classification. Pass it some body of text in the expected format and it will output the estimated sentiment. Positive numbers indicate positive sentiment and negative numbers indicate negative sentiment. It will also give you the estimated sentiment of each word (given on the interval [-1,1].)

**Expected JSON**

```json
{
    "text": "This is an awesome day!"
}
```

**Returned JSON**

```json
{
  "words": [
    {
      "word": "this",
      "score": -0.07602276959089216
    },
    {
      "word": "is",
      "score": 0
    },
    {
      "word": "an",
      "score": 0
    },
    {
      "word": "awesome",
      "score": 0.513978494623656
    },
    {
      "word": "day",
      "score": 0.1724137931034483
    }
  ],
  "score": 0.6103695181362121
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
      "score": -0.07602276959089216
    },
    {
      "word": "is",
      "score": 0
    },
    {
      "word": "an",
      "score": 0
    },
    {
      "word": "awesome",
      "score": 0.513978494623656
    },
    {
      "word": "day",
      "score": 0.1724137931034483
    }
  ],
  "score": 0.6103695181362121
}
```
