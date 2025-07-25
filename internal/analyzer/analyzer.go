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

// HTTPClient is an interface to allow mocking HTTP requests in tests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Analyzer interface {
	FetchHTML(targetURL string) (string, error)
	ExtractTitle(body string) string
	CountHeadings(body string) map[string]int
	AnalyzeLinks(body, baseURL string) (internal, external, broken int, err error)
	DetectLoginForm(body string) bool
	DetectHTMLVersion(body string) string
}

type DefaultAnalyzer struct {
	Client HTTPClient
}

func NewAnalyzer(client HTTPClient) Analyzer {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &DefaultAnalyzer{Client: client}
}

const maxWorkers = 10

// FetchHTML fetches the HTML content of the page and returns it as a string.
func (analyser *DefaultAnalyzer) FetchHTML(targetURL string) (string, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := analyser.Client.Do(req)
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
func (analyser *DefaultAnalyzer) ExtractTitle(body string) string {
	var title string
	tokenizer := html.NewTokenizer(strings.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			break
		}
		if tokenType == html.StartTagToken {
			token := tokenizer.Token()
			if token.Data == "title" {
				tokenType = tokenizer.Next()
				if tokenType == html.TextToken {
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
func (analyser *DefaultAnalyzer) CountHeadings(body string) map[string]int {
	headers := make(map[string]int)
	headerRegex := regexp.MustCompile(`^h[1-6]$`)

	tokenizer := html.NewTokenizer(strings.NewReader(body))
	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			break // stop at the end of the document
		}

		if tokenType == html.StartTagToken {
			token := tokenizer.Token()
			if headerRegex.MatchString(token.Data) {
				headerType := strings.ToLower(token.Data)
				headers[headerType]++
			}
		}
	}

	return headers
}

// AnalyzeLinks parses links from the HTML body, resolves relative URLs, handles <base> tags,
// and counts internal, external, and broken links on the page.
// It uses concurrency to efficiently check the accessibility of each link.
func (analyser *DefaultAnalyzer) AnalyzeLinks(body, baseURL string) (internal, external, broken int, err error) {
	var baseParsed *url.URL
	var links []string

	tokenizer := html.NewTokenizer(strings.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			break
		}
		if tokenType != html.StartTagToken && tokenType != html.SelfClosingTagToken {
			continue
		}

		token := tokenizer.Token()
		switch token.Data {
		case "base":
			// Parse the <base> tag to resolve relative URLs correctly
			baseParsed = extractBaseHref(token)
		case "a", "link":
			links = extractLinks(token, baseParsed, links)
		}
	}

	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return 0, 0, 0, err
	}

	type linkResult struct {
		isInternal bool
		isBroken   bool
	}

	// Use buffered channels for jobs and results to handle concurrency
	jobs := make(chan string, len(links))
	results := make(chan linkResult, len(links))

	// Spawn worker goroutines to check link accessibility
	var waitGroup sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for link := range jobs {
				resolvedURL, err := url.Parse(link)
				if err != nil {
					results <- linkResult{false, true}
					continue
				}

				resolved := parsedBaseURL.ResolveReference(resolvedURL)
				isInternal := sameHost(parsedBaseURL, resolved)
				isBroken := analyser.checkLinkBroken(resolved.String())
				results <- linkResult{isInternal, isBroken}
			}
		}()
	}

	// Feed links to workers via the jobs channel
	for _, link := range links {
		jobs <- link
	}
	close(jobs) // Close jobs channel to signal no more links

	waitGroup.Wait()
	close(results) // Close results channel after all workers finish

	// Aggregate results from workers
	for res := range results {
		if res.isInternal {
			internal++
		} else {
			external++
		}
		if res.isBroken {
			broken++
		}
	}

	return internal, external, broken, nil
}

// extractLinks filters and extracts href attributes from <a> and <link> tags,
// ignoring mailto:, tel:, and javascript: schemes to avoid non-http links.
func extractLinks(token html.Token, baseParsed *url.URL, links []string) []string {
	for _, attr := range token.Attr {
		if attr.Key != "href" {
			continue
		}

		link := attr.Val
		parsed, err := url.Parse(link)

		if err != nil {
			return links
		}

		if parsed.Scheme == "mailto" || parsed.Scheme == "tel" || parsed.Scheme == "javascript" {
			return links
		}

		if baseParsed != nil && !parsed.IsAbs() {
			link = baseParsed.ResolveReference(parsed).String()
		}

		links = append(links, link)
	}
	return links
}

// extractBaseHref parses a <base> tag and returns the resolved *url.URL if present.
func extractBaseHref(token html.Token) *url.URL {
	for _, attr := range token.Attr {
		if attr.Key == "href" {
			if parsed, err := url.Parse(attr.Val); err == nil {
				return parsed
			}
		}
	}
	return nil
}

// checkLinkBroken sends a HEAD request (or fallback GET) and returns whether the link is broken.
func (analyser *DefaultAnalyzer) checkLinkBroken(link string) bool {
	// Try HEAD request to check link quickly without downloading the body
	req, err := http.NewRequest("HEAD", link, nil)
	if err != nil {
		return true
	}

	resp, err := analyser.Client.Do(req)
	if err != nil {
		return true
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	// Fallback to GET if HEAD not allowed, since some servers do not support HEAD requests
	if resp.StatusCode == http.StatusMethodNotAllowed {
		req, err = http.NewRequest("GET", link, nil)
		if err != nil {
			return true
		}
		resp, err = analyser.Client.Do(req)
		if err != nil {
			return true
		}
		defer func() {
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
		}()
	}

	return resp.StatusCode >= 400
}

// DetectLoginForm checks if the HTML body contains a form with an input of a type "password"
// or any attribute matches the word "login"
func (analyser *DefaultAnalyzer) DetectLoginForm(body string) bool {
	tokenizer := html.NewTokenizer(strings.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			break
		}
		if tokenType != html.StartTagToken && tokenType != html.SelfClosingTagToken {
			continue
		}
		token := tokenizer.Token()
		switch token.Data {
		case "input":
			// Detect login form by presence of input[type=password]
			inputType := strings.ToLower(getAttributeValue(token, "type"))
			if inputType == "password" {
				return true
			}
		case "form":
			// Also detect login forms by checking if form's action or class attribute contains "login"
			action := strings.ToLower(getAttributeValue(token, "action"))
			class := strings.ToLower(getAttributeValue(token, "class"))
			if strings.Contains(action, "login") || strings.Contains(class, "login") {
				return true
			}
		}
	}
	return false
}

// getAttributeValue retrieves the value of a given attribute key from an HTML token.
// If the attribute does not exist, it returns an empty string.
func getAttributeValue(tag html.Token, key string) string {
	for _, attr := range tag.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// DetectHTMLVersion determines the HTML version by matching known DOCTYPE declarations in the HTML body.
func (analyser *DefaultAnalyzer) DetectHTMLVersion(body string) string {
	htmlDeclarations := map[string]string{
		"HTML 5":                 "<!DOCTYPE html>",
		"HTML 4.01 Strict":       "-//W3C//DTD HTML 4.01//EN",
		"HTML 4.01 Transitional": "-//W3C//DTD HTML 4.01 Transitional//EN",
		"HTML 4.01 Frameset":     "-//W3C//DTD HTML 4.01 Frameset//EN",
		"HTML 4.0 Strict":        "-//W3C//DTD HTML 4.0//EN",
		"HTML 3.2":               "-//W3C//DTD HTML 3.2//EN",
		"HTML 2.0":               "-//IETF//DTD HTML 2.0//EN",
		"HTML 1.0":               "-//IETF//DTD HTML 1.0//EN",
		"XHTML 1.0 Strict":       "-//W3C//DTD XHTML 1.0 Strict//EN",
		"XHTML 1.0 Transitional": "-//W3C//DTD XHTML 1.0 Transitional//EN",
		"XHTML 1.0 Frameset":     "-//W3C//DTD XHTML 1.0 Frameset//EN",
		"XHTML 1.1":              "-//W3C//DTD XHTML 1.1//EN",
	}

	lowerBody := strings.ToLower(body)

	for version, declaration := range htmlDeclarations {
		if strings.Contains(lowerBody, strings.ToLower(declaration)) {
			return version
		}
	}

	return "Unknown"
}

func sameHost(base, other *url.URL) bool {
	return base.Hostname() == other.Hostname()
}
