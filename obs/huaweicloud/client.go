/*
 * Copyright (C) 2021. Huawei Technologies Co., Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package huaweicloud

import (
	"bytes"

	sdk "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"

	"github.com/opensourceways/app-cla-server/obs"
)

const plugin = "huaweicloud-obs"

type client struct {
	c *sdk.ObsClient

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

	cli.c = c
	cli.bucket = bucket
	cli.ssecHeader = newSSECHeader(cfg.ObjectEncryptionKey)

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

func (cli *client) ListObject(pathPrefix string) ([]string, error) {
	input := sdk.ListObjectsInput{
		Bucket: cli.bucket,
	}
	if pathPrefix != "" {
		input.Prefix = pathPrefix
	}

	r := make([]string, 0, 100)
	for {
		output, err := cli.c.ListObjects(&input)
		if err != nil {
			return nil, err
		}

		for i := range output.Contents {
			r = append(r, output.Contents[i].Key)
		}

		if output.IsTruncated {
			input.Marker = output.NextMarker
		} else {
			break
		}
	}

	return r, nil
}

func newSSECHeader(key string) sdk.ISseHeader {
	if key == "" {
		return nil
	}

	h := sdk.SseCHeader{
		Key: sdk.Base64Encode([]byte(key)),
	}
	h.KeyMD5 = h.GetKeyMD5()

	return h
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
