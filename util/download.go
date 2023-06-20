package util

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

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

func head(url string, fileType string, maxSize int) error {
	return try(func() error {
		resp, err := http.Head(url)
		if err != nil {
			return err
		}

		if c := resp.StatusCode; !(c >= 200 && c < 300) {
			return errors.New("can't detect")
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

		return nil
	})
}

func DownloadFile(url, fileType string, maxSize int) ([]byte, error) {
	if err := head(url, fileType, maxSize); err != nil {
		return nil, err
	}

	var content []byte

	err := try(func() error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}

		content, err = ioutil.ReadAll(resp.Body)

		resp.Body.Close()

		return err
	})

	return content, err
}
