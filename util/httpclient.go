package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpClient struct {
	client     http.Client
	maxRetries int
}

func newClient(timeout int) http.Client {
	return http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
}

func NewHttpClient(n, timeout int) HttpClient {
	return HttpClient{
		maxRetries: n,
		client:     newClient(timeout),
	}
}

func (hc *HttpClient) ForwardTo(req *http.Request, jsonResp interface{}) (statusCode int, err error) {
	resp, err := hc.do(req)
	if err != nil || resp == nil {
		return
	}

	defer resp.Body.Close()

	if code := resp.StatusCode; code < 200 || code > 299 {
		statusCode = code

		var rb []byte
		if rb, err = ioutil.ReadAll(resp.Body); err == nil {
			err = fmt.Errorf("response has status:%s and body:%q", resp.Status, rb)
		}

		return
	}

	if jsonResp != nil {
		err = json.NewDecoder(resp.Body).Decode(jsonResp)
	}

	return
}

func (hc *HttpClient) do(req *http.Request) (resp *http.Response, err error) {
	if resp, err = hc.client.Do(req); err == nil {
		return
	}

	maxRetries := hc.maxRetries
	backoff := 10 * time.Millisecond

	for retries := 1; retries < maxRetries; retries++ {
		time.Sleep(backoff)
		backoff *= 2

		if resp, err = hc.client.Do(req); err == nil {
			break
		}
	}
	return
}
