package benchmark

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	// time out requests after 20 seconds
	timeout = 3 * time.Minute
)

type Client struct {
	http *http.Client
	url  string
}

func NewFissionClient(routerURL string) *Client {
	return &Client{
		url: routerURL,
		http: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c Client) execFn(fnName string) (string, error) {
	url, err := url.Parse(mkRequestURL(fnName, c.url))

	if err != nil {
		return "", err
	}

	req := &http.Request{
		Method: "GET",
		URL:    url,
		Header: map[string][]string{
			"x-consent-id": {"Bar"},
		},
	}
	//resp, err := c.http.Get(mkRequestURL(fnName, c.url))
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}

	if resp.Status != "200 OK" {
		return "", fmt.Errorf("Fission returned %v status code", resp.Status)
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(result), nil
}

func mkRequestURL(fnName string, router string) string {
	return fmt.Sprintf("%s/%s", router, fnName)
}
