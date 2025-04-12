package main

import (
	"fmt"
	"net/url"
	"net/http"
	"strings"
	"strconv"
	"sync"
	"os"
	"io"

	"golang.org/x/net/html"
)

type config struct {
	pages       map[string]int
	baseURL     string
	mutex       *sync.Mutex
	controlChan chan struct{}
	waitGroup   *sync.WaitGroup
	maxPages    int
}

func isValidAbsoluteURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil { return false }
	return parsedURL.IsAbs()
}

func normalizeURL(inputURL string) (string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", err
	}

	host := parsedURL.Host
	path := strings.TrimRight(parsedURL.Path, "/")

	normalizedURL := fmt.Sprintf("%s%s", host, path)

	return normalizedURL, nil
}

func getURLsFromHTML(htmlStr, rawBaseURL string) ([]string, error) {
	var hrefs []string

	var getAnchorsHrefs func(*html.Node)
	getAnchorsHrefs = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attr := range node.Attr {
				if attr.Key != "href" {
					continue
				}
				hrefs = append(hrefs, attr.Val)
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			getAnchorsHrefs(c)
		}
	}

	docRoot, err := html.Parse(strings.NewReader(htmlStr))

	if err != nil {
		return nil, err
	}
	getAnchorsHrefs(docRoot)

	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, err
	}

	var URLs []string
	for _, href := range hrefs {
		parsedURL, err := url.Parse(href)
		if err != nil {
			continue
		}

		if parsedURL.IsAbs() {
			URLs = append(URLs, parsedURL.String())
		} else {
			resolved := baseURL.ResolveReference(parsedURL)
			URLs = append(URLs, resolved.String())
		}
	}

	return URLs, nil
}

func getHTML(rawURL string) (string, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf(resp.Status)
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return "", fmt.Errorf("Content type is not html")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool) {
	cfg.mutex.Lock()
	defer cfg.mutex.Unlock()
	cfg.pages[normalizedURL]++
	if cfg.pages[normalizedURL] >= 2 {
		return false
	}
	return true
}

func (cfg *config) isMaxPages() (isFirst bool) {
	cfg.mutex.Lock()
	defer cfg.mutex.Unlock()
	return len(cfg.pages) >= cfg.maxPages
}

func (cfg *config) crawlPage(rawCurrentURL string) {
	defer cfg.waitGroup.Done()
	
	if cfg.isMaxPages() {
		return
	}

	normalizedURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Printf("func normalizeURL error: %v\n", err)
		return
	}

	if !cfg.addPageVisit(normalizedURL) {
		return
	}

	htmlStr, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Printf("func getHTML error: %v\n", err)
		return
	}

	URLs, err := getURLsFromHTML(htmlStr, cfg.baseURL)
	if err != nil {
		fmt.Printf("func getURLsFromHTML error: %v\n", err)
		return
	}

	fmt.Println(rawCurrentURL)
	for _, URL := range URLs {
		if strings.HasPrefix(URL, cfg.baseURL) {
			cfg.waitGroup.Add(1)

			select {
			case cfg.controlChan <- struct{}{}:
				go func(urlToCrawl string) {
					cfg.crawlPage(urlToCrawl)
					<-cfg.controlChan 
				}(URL)
			default:
				cfg.crawlPage(URL)
				continue
			}
		}
	}
}

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Usage: %s <URL> <MAX_PAGES> <MAX_CHANELS>\n", os.Args[0])
		os.Exit(1)
	}

	baseURL := os.Args[1]
	if !isValidAbsoluteURL(baseURL) {
		fmt.Printf("'%s' is not a valid URL\n", os.Args[1])
		os.Exit(1)
	}

	maxPages, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("MAX_PAGES expects a number\n")
	}

	maxChans, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Printf("MAX_CHANELS expects a number\n")
	}

	fmt.Println("Starting web crawl at:", baseURL)

	cfg := &config{
		pages:       make(map[string]int),
		baseURL:     baseURL,
		mutex:       &sync.Mutex{},
		controlChan: make(chan struct{}, maxChans),
		waitGroup:   &sync.WaitGroup{},
		maxPages:    maxPages,
	}

	cfg.waitGroup.Add(1)
	cfg.crawlPage(baseURL)
	cfg.waitGroup.Wait()

	fmt.Println("\nPages with visit count:")
	for key, val := range cfg.pages {
		fmt.Printf("%s: %d\n", key, val)
	}
}
