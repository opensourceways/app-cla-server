package util

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const userAgent = "Mozilla/5.0"

func try(f func() error) (err error) {
	if err = f(); err == nil {
		return
	}

	for i := 1; i < 3; i++ {
		time.Sleep(time.Millisecond * time.Duration(10))

		if err = f(); err == nil {
			return
		}
	}

	return
}

func DownloadFile(url, fileType string, maxSize int) ([]byte, error) {
	var content []byte

	err := try(func() error {
		cli := newClient(10)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", userAgent)

		resp, err := cli.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if c := resp.StatusCode; !(c >= 200 && c < 300) {
			return errors.New("can't download")
		}

		ct := resp.Header.Get("content-type")
		if !strings.Contains(strings.ToLower(ct), strings.ToLower(fileType)) {
			return errors.New("unknown file type")
		}

		if resp.ContentLength == -1 {
			return errors.New("unknown file size")
		}

		if resp.ContentLength > int64(maxSize) {
			return errors.New("big file")
		}

		content, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return nil
	})

	return content, err
}
