package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpClient struct {
	Client     *http.Client
	MaxRetries int
}

func NewHttpClient(n int) HttpClient {
	return HttpClient{
		MaxRetries: n,
		Client:     http.DefaultClient,
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

func (hc *HttpClient) Download(req *http.Request) (r []byte, statusCode int, err error) {
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

	} else {
		r, err = ioutil.ReadAll(resp.Body)
	}

	return
}

func (hc *HttpClient) do(req *http.Request) (resp *http.Response, err error) {
	if resp, err = hc.Client.Do(req); err == nil {
		return
	}

	maxRetries := hc.MaxRetries
	backoff := 10 * time.Millisecond

	for retries := 1; retries < maxRetries; retries++ {
		time.Sleep(backoff)
		backoff *= 2

		if resp, err = hc.Client.Do(req); err == nil {
			break
		}
	}
	return
}

func JsonMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	enc := json.NewEncoder(buffer)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(t); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
