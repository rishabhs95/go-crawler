# Go concurrent web crawler
A simple intra target domain web crawler.

## Dependencies
  - Native:
    - os
    - fmt
    - sync
    - strings
    - net/url
    - net/http
  - External:
    - golang.org/x/net/html
	- github.com/op/go-logging

## Execution
  - Get all the dependencies installed:
    - `go get golang.org/x/net/html`
    - `go get github.com/op/go-logging`
  - Execute the crawler:
    - `go run crawler.go`

## References
  - [Web Crawler](https://schier.co/blog/2015/04/26/a-simple-web-scraper-in-go.html)
  - [Concurrency Article](https://medium.com/@a4word/webl-a-simple-web-crawler-written-in-go-c1ce50b4f687)

## Limitations
  - Currently doesn't respect the `robots.txt` rule file for websites.
