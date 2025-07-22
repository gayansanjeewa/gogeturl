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

// ExtractTitle returns the content of the <title> tag from the HTML body string
func ExtractTitle(body string) string {
	var title string
	tokenizer := html.NewTokenizer(strings.NewReader(body))
	for {
		tokType := tokenizer.Next()
		if tokType == html.ErrorToken {
			break
		}
		if tokType == html.StartTagToken {
			token := tokenizer.Token()
			if token.Data == "title" {
				tokType = tokenizer.Next()
				if tokType == html.TextToken {
					title = tokenizer.Token().Data
					break
				}
			}
		}
	}
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

// AnalyzeLinks parses links from the HTML body, resolving relative URLs and handling base tags,
// and counts internal, external, and broken links. It skips mailto:, tel:, and javascript: schemes.
func AnalyzeLinks(body, baseURL string) (internal, external, broken int, err error) {
	var baseParsed *url.URL
	var links []string

	tokenizer := html.NewTokenizer(strings.NewReader(body))
	for {
		tokType := tokenizer.Next()
		if tokType == html.ErrorToken {
			break
		}
		if tokType != html.StartTagToken && tokType != html.SelfClosingTagToken {
			continue
		}

		token := tokenizer.Token()
		switch token.Data {
		case "base":
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					parsedBaseURL, err := url.Parse(attr.Val)
					if err == nil {
						baseParsed = parsedBaseURL
					}
				}
			}
		case "a", "link":
			for _, attr := range token.Attr {
				if attr.Key != "href" {
					continue
				}
				link := attr.Val
				parsedLink, err := url.Parse(link)
				if err != nil {
					continue
				}
				// Skip mailto, tel, javascript schemes
				if parsedLink.Scheme == "mailto" || parsedLink.Scheme == "tel" || parsedLink.Scheme == "javascript" {
					continue
				}
				// Resolve relative URLs with base if present
				if baseParsed != nil && !parsedLink.IsAbs() {
					link = baseParsed.ResolveReference(parsedLink).String()
				}
				links = append(links, link)
			}
		}
	}

	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return 0, 0, 0, err
	}
	client := http.Client{Timeout: 5 * time.Second}
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, link := range links {
		wg.Add(1)
		go func(link string) {
			defer wg.Done()
			parsed, err := url.Parse(link)
			if err != nil {
				mu.Lock()
				broken++
				mu.Unlock()
				return
			}

			// Classify as internal or external
			if parsed.Host != "" && parsed.Host != parsedBaseURL.Host {
				mu.Lock()
				external++
				mu.Unlock()
			} else {
				mu.Lock()
				internal++
				mu.Unlock()
			}

			// Resolve the absolute URL for checking
			if !parsed.IsAbs() && baseParsed != nil {
				parsed = baseParsed.ResolveReference(parsed)
			} else if !parsed.IsAbs() {
				parsed = parsedBaseURL.ResolveReference(parsed)
			}

			resp, err := client.Head(parsed.String())
			if err != nil || resp.StatusCode >= 400 {
				mu.Lock()
				broken++
				mu.Unlock()
			}
		}(link)
	}
	wg.Wait()

	return internal, external, broken, nil
}
