package analyzer

import (
	"errors"
	"io"
	"net/http"
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
