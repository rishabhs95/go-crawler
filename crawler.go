/*  Input:
        [x] URL and the depth to crawl
    Requirements:
        [x] Limited to one domain
    Output:
        [x] Site asset tree
*/
package main

import (
    "os"
    "fmt"
	"sync"
	"strings"
	"net/url"
	"net/http"

	"golang.org/x/net/html"
	"github.com/op/go-logging"
)

const depth = 2
const hostname = "http://facebook.com/"

var log = logging.MustGetLogger("crawler")

var wg sync.WaitGroup

var visited map[string] bool
var pageAssets map[string] []string
var pageStaticAssets map[string] []string

var visitedMutex sync.Mutex
var pageAssetMutex sync.Mutex
var pageStaticAssetMutex sync.Mutex

func main() {
    pageAssets = make(map[string] []string)
    pageStaticAssets = make(map[string] []string)
	visited = make(map[string] bool)

    // Parse url
	parsedURL, err := url.Parse(hostname)
	if err != nil {
		log.Error("Invalid URL :", err)
		os.Exit(1)
	}

    // Mark parent url as visited
	visitedMutex.Lock()
	visited[parsedURL.String()] = true
	visitedMutex.Unlock()

	wg.Add(1)
	go startCrawler(parsedURL.String(), depth)
	wg.Wait()
    marked := make(map[string] bool)
	displayCrawledInfo(hostname, marked, 0)
}

func startCrawler(targetUrlString string, depth int) {
	marked := make(map[string] bool)

	defer wg.Done()
	if depth <= 0 {
		return
	}

    // Parse url and send HTTP requests
	response, httpError := http.Get(targetUrlString)
	if httpError != nil {
		log.Error("Error in fetching HTTP request to %s: %v", targetUrlString, httpError)
		return
	}

	defer response.Body.Close()
	contentType := response.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "text/html") {
		return
	}

    // Setup sync waitgroups for assets
	var assets_wg sync.WaitGroup

    // Parse html response
	tokens := html.NewTokenizer(response.Body)
	for {
		tokenType := tokens.Next()
		if tokenType == html.ErrorToken {
			return
		}
		token := tokens.Token()
		if tokenType == html.StartTagToken {
            dataAtom := token.DataAtom.String()
			if (dataAtom == "a" || dataAtom == "link") {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						_, mapValue := marked[attr.Val]
						if !mapValue {
							marked[attr.Val] = true
                            // Concurrent goroutine for parsing assets
							assets_wg.Add(1)
							go parseAssets(attr.Val, targetUrlString, &assets_wg, depth)
						}
					}
				}
            }
			if (dataAtom == "img" || dataAtom == "image" || dataAtom == "script") {
				for _, attr := range token.Attr {
					if attr.Key == "src" {
						_, mapValue := marked[attr.Val]
						if !mapValue {
							marked[attr.Val] = true
                            // Concurrent goroutine for parsing static assets
							assets_wg.Add(1)
							go parseStaticAssets(attr.Val, targetUrlString, &assets_wg)
						}
					}
				}
			}
		}
	}
}

func parseAssets(childURL string, parentURL string, waitgroup *sync.WaitGroup, depth int) {
	defer (*waitgroup).Done()
	parsedURL, err := url.Parse(childURL)
	if err != nil {
		log.Error("Invalid URL :", err)
		os.Exit(1)
	}
	currentURL, _ := url.Parse(parentURL)
	newURL := currentURL.ResolveReference(parsedURL)
	if newURL.Host != currentURL.Host {
		return
	}
	newURL.Fragment = ""

    // mark links as visited
	visitedMutex.Lock()
	_, mapValue := visited[newURL.String()]
	if mapValue {
		visitedMutex.Unlock()
		return
	}
	visited[newURL.String()] = true
	visitedMutex.Unlock()

    // append page assets with locks for concurrency handling in maps
    if (childURL != "/") {
        pageAssetMutex.Lock()
        pageAssets[parentURL] = append(pageAssets[parentURL], childURL)	
        pageAssetMutex.Unlock()
    }

    // Recursive crawl with new goroutine
	wg.Add(1)
	go startCrawler(newURL.String(), depth-1)
}

func parseStaticAssets(childURL string, parentURL string, waitgroup *sync.WaitGroup) {
	defer (*waitgroup).Done()
	parsedURL, err := url.Parse(childURL)
	if err != nil {
		log.Error("Invalid URL :", err)
		os.Exit(1)
	}
	currentURL, _ := url.Parse(parentURL)
	newURL := currentURL.ResolveReference(parsedURL)
	newURL.Fragment = ""

    // append static page assets with locks for concurrency handling in maps
    if (newURL.String() != "/") {
        pageStaticAssetMutex.Lock()
        pageStaticAssets[parentURL] = append(pageStaticAssets[parentURL], newURL.String())
        pageStaticAssetMutex.Unlock()
    }
}

func displayCrawledInfo(url string, marked map[string] bool, depth int) {
	fmt.Printf("%s└── %s\n", strings.Repeat(" ", depth), url)
	marked[url] = true
	if len(pageStaticAssets[url]) > 0 {
		for idx, asset := range pageStaticAssets[url] {
			marker := "├"
			if idx == len(pageStaticAssets[url])-1 {
				marker = "└"
			}
			fmt.Printf("%s│%s%s── %s\n", strings.Repeat(" ", depth+2), strings.Repeat(" ", depth+2), marker, asset)
		}
	}

	if len(pageAssets[url]) > 0 {
		for _, link := range pageAssets[url] {
            _, exists := marked[link]
			if !exists {
				displayCrawledInfo(link, marked, depth+2)
			}
		}
	}
}
