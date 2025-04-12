package main

import (
	"net/url"
	"net/http"
	"strings"
	"fmt"
	"os"
	"io"

	"golang.org/x/net/html"
)

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
		if node.Data == "a" {
			for _, attr := range node.Attr {
				if attr.Key != "href" { continue }
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

func crawlPage(rawBaseURL, rawCurrentURL string, pages map[string]int) {
	normalizedURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Printf("func normalizeURL error: %v\n", err)
		return
	}

	pages[normalizedURL]++
	if pages[normalizedURL] >= 2 {
		return
	}

	htmlStr, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Printf("func getHTML error: %v\n", err)
		return
	}

	URLs, err := getURLsFromHTML(htmlStr, rawBaseURL)
	if err != nil {
		fmt.Printf("func getURLsFromHTML error: %v\n", err)
		return
	}

	fmt.Println(rawCurrentURL)
	for _, URL := range URLs {
		if strings.HasPrefix(URL, rawBaseURL) {
			crawlPage(rawBaseURL, URL, pages)
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: %s <URL>", os.Args[0])
		os.Exit(1)
	}
	baseURL := os.Args[1]	
	fmt.Println("Starting web crawl at:", baseURL)

	pages := make(map[string]int)
	crawlPage(baseURL, baseURL, pages)
	
	fmt.Println("\n=========================================")
	for key, val := range pages {
		fmt.Printf("%s: %d\n", key, val)
	}		
}
