package obsimpl

import (
	"bytes"
	"fmt"

	"github.com/beego/beego/v2/core/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

// PDFStorage defines the interface for PDF object storage operations.
type PDFStorage interface {
	Upload(corpSigningId string, data []byte) (string, error)
	Download(key string) ([]byte, error)
}

func NewPDFStorage(cfg *Config) (PDFStorage, error) {
	client, err := obs.New(cfg.AccessKey, cfg.SecretKey, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create obs client: %w", err)
	}

	return &obsPDFStorage{
		client: client,
		bucket: cfg.Bucket,
		cfg:    cfg,
	}, nil
}

type obsPDFStorage struct {
	client *obs.ObsClient
	bucket string
	cfg   *Config
}

func (s *obsPDFStorage) Upload(corpSigningId string, data []byte) (string, error) {
	key := s.cfg.objectKey(corpSigningId)

	input := &obs.PutObjectInput{
		PutObjectBasicInput: obs.PutObjectBasicInput{
			ObjectOperationInput: obs.ObjectOperationInput{
				Bucket: s.bucket,
				Key:    key,
			},
			HttpHeader: obs.HttpHeader{
				ContentType: "application/pdf",
			},
		},
		Body: bytes.NewReader(data),
	}

	output, err := s.client.PutObject(input)
	if err != nil {
		logs.Error("obs upload failed, key=%s, err=%s", key, err.Error())
		return "", fmt.Errorf("obs upload failed: %w", err)
	}

	logs.Info("obs upload succeeded, key=%s, status=%d", key, output.StatusCode)

	return key, nil
}

func (s *obsPDFStorage) Download(key string) ([]byte, error) {
	input := &obs.GetObjectInput{
		GetObjectMetadataInput: obs.GetObjectMetadataInput{
			Bucket: s.bucket,
			Key:    key,
		},
	}

	output, err := s.client.GetObject(input)
	if err != nil {
		logs.Error("obs download failed, key=%s, err=%s", key, err.Error())
		return nil, fmt.Errorf("obs download failed: %w", err)
	}
	defer output.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(output.Body); err != nil {
		logs.Error("obs read body failed, key=%s, err=%s", key, err.Error())
		return nil, fmt.Errorf("obs read body failed: %w", err)
	}

	logs.Info("obs download succeeded, key=%s, size=%d", key, buf.Len())

	return buf.Bytes(), nil
}
