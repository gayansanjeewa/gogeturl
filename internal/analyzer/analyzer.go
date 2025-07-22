package analyzer

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// FetchAndParse fetches the HTML page and parses it into a DOM
func FetchAndParse(targetURL string) (*html.Node, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("received non-2xx status code: " + resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, errors.New("failed to parse HTML: " + err.Error())
	}

	return doc, nil
}

// ExtractTitle returns the content of the <title> tag
func ExtractTitle(doc *html.Node) string {
	var title string
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "title" && node.FirstChild != nil {
			title = node.FirstChild.Data
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}
	traverse(doc)
	return title
}

// CountHeadings returns a map of heading levels and their counts (e.g., h1, h2...)
func CountHeadings(doc *html.Node) map[string]int {
	headings := map[string]int{}
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "h1", "h2", "h3", "h4", "h5", "h6":
				headings[node.Data]++
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}
	traverse(doc)
	return headings
}

// AnalyzeLinks counts internal and external links and collects broken links
func AnalyzeLinks(doc *html.Node, baseURL string) (internal, external, broken int, err error) {
	var links []string
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attribute := range node.Attr {
				if attribute.Key == "href" && attribute.Val != "" {
					links = append(links, attribute.Val)
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}
	traverse(doc)

	baseParsed, err := url.Parse(baseURL)
	if err != nil {
		return 0, 0, 0, err
	}

	client := http.Client{Timeout: 5 * time.Second}
	var waitGroup sync.WaitGroup
	var mutex sync.Mutex

	for _, link := range links {
		waitGroup.Add(1)
		go func(link string) {
			defer waitGroup.Done()
			parsed, err := url.Parse(link)
			if err != nil {
				mutex.Lock()
				broken++
				mutex.Unlock()
				return
			}
			isExternal := parsed.Host != "" && parsed.Host != baseParsed.Host
			if isExternal {
				mutex.Lock()
				external++
				mutex.Unlock()
			} else {
				mutex.Lock()
				internal++
				mutex.Unlock()
			}

			// Check if link is accessible
			reqURL := parsed
			if !parsed.IsAbs() {
				reqURL = baseParsed.ResolveReference(parsed)
			}
			resp, err := client.Head(reqURL.String())
			if err != nil || resp.StatusCode >= 400 {
				mutex.Lock()
				broken++
				mutex.Unlock()
			}
		}(link)
	}
	waitGroup.Wait()

	return internal, external, broken, nil
}
