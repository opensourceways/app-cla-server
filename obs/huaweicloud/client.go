package huaweicloud

import (
	"bytes"
	"fmt"

	sdk "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"

	"github.com/opensourceways/app-cla-server/obs"
)

const plugin = "huaweicloud-obs"

type client struct {
	c *sdk.ObsClient

	// region     string
	bucket     string
	ssecHeader sdk.ISseHeader
}

func init() {
	obs.Register(plugin, &client{})
}

func (cli *client) Initialize(path, bucket string) error {
	cfg, err := loadConfig(path)
	if err != nil {
		return err
	}

	c, err := sdk.New(cfg.AccessKey, cfg.SecretKey, cfg.Endpoint)
	if err != nil {
		return err
	}

	if _, err = c.GetBucketLocation(bucket); err != nil {
		return err
	}

	h, err := newSSECHeader(cfg.ObjectEncryptionKey)
	if err != nil {
		return err
	}

	cli.c = c
	// cli.region = cfg.Region
	cli.bucket = bucket
	cli.ssecHeader = h

	return nil
}

func (cli *client) WriteObject(path string, data []byte) error {
	input := sdk.PutObjectInput{Body: bytes.NewReader(data)}
	input.Bucket = cli.bucket
	input.Key = path
	input.SseHeader = cli.ssecHeader
	input.ContentMD5 = sdk.Base64Md5(data)

	_, err := cli.c.PutObject(&input)
	return err
}

func (cli *client) ReadObject(path, localPath string) obs.OBSError {
	input := sdk.DownloadFileInput{DownloadFile: localPath}
	input.Bucket = cli.bucket
	input.Key = path
	input.SseHeader = cli.ssecHeader

	_, err := cli.c.DownloadFile(&input)
	if err == nil {
		return nil
	}

	return obsError{err: err}
}

func (cli *client) HasObject(path string) (bool, error) {
	input := sdk.GetObjectMetadataInput{
		Bucket:    cli.bucket,
		Key:       path,
		SseHeader: cli.ssecHeader,
	}
	_, err := cli.c.GetObjectMetadata(&input)
	if err == nil {
		return true, nil
	}

	e := obsError{err: err}
	if e.IsObjectNotFound() {
		return false, nil
	}

	return false, err
}

func newSSECHeader(key string) (sdk.ISseHeader, error) {
	h := sdk.SseCHeader{
		Key: sdk.Base64Encode([]byte(key)),
	}

	v := h.GetKeyMD5()
	if v == "" {
		return nil, fmt.Errorf("build md5 of object key failed")
	}
	h.KeyMD5 = v

	return h, nil
}

type obsError struct {
	err error
}

func (e obsError) IsObjectNotFound() bool {
	er, ok := e.err.(sdk.ObsError)
	return ok && er.StatusCode == 404
}

func (e obsError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}
