package util

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	// 覆盖正常下载、文件类型不匹配、文件过大、ContentLength为-1、状态码非2xx等关键路径
	cases := []struct {
		name     string
		handler  http.HandlerFunc
		fileType string
		maxSize  int
		wantErr  bool
	}{
		{
			name: "正常下载PDF",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/pdf")
				w.Write([]byte("%PDF-1.7"))
			},
			fileType: "pdf",
			maxSize:  1024,
			wantErr:  false,
		},
		{
			name: "文件类型不匹配",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("{}"))
			},
			fileType: "pdf",
			maxSize:  1024,
			wantErr:  true,
		},
		{
			name: "文件超过maxSize",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/pdf")
				w.Write(make([]byte, 100))
			},
			fileType: "pdf",
			maxSize:  10,
			wantErr:  true,
		},
		{
			name: "ContentLength为-1",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/pdf")
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				w.Write([]byte("%PDF-1.7"))
			},
			fileType: "pdf",
			maxSize:  1024,
			wantErr:  true,
		},
		{
			name: "状态码非2xx",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			fileType: "pdf",
			maxSize:  1024,
			wantErr:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ts := httptest.NewServer(c.handler)
			defer ts.Close()

			_, err := DownloadFile(ts.URL, c.fileType, c.maxSize)
			if (err != nil) != c.wantErr {
				t.Errorf("wantErr=%v, got err=%v", c.wantErr, err)
			}
		})
	}
}
