package analyzer

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// FetchHTML fetches the HTML content of the page and returns it as a string
func FetchHTML(targetURL string) (string, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("received non-2xx status code: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("failed to read response body: " + err.Error())
	}

	return string(body), nil
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

// CountHeadings counts the number of headers in the HTML document, sorted by type.
// It accepts the body of the HTML document as a string and returns a map of header types to their respective counts.
func CountHeadings(body string) map[string]int {
	headers := make(map[string]int)
	re := regexp.MustCompile(`^h[1-6]$`)

	tokenizer := html.NewTokenizer(strings.NewReader(body))
	for {
		tokType := tokenizer.Next()

		if tokType == html.ErrorToken {
			break // stop at the end of the document
		}

		if tokType == html.StartTagToken {
			tok := tokenizer.Token()
			if re.MatchString(tok.Data) {
				headerType := strings.ToLower(tok.Data)
				headers[headerType]++
			}
		}
	}

	return headers
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
