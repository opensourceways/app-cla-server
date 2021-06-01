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

package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/dbmodels"
)

var _ dbmodels.IModel = (*client)(nil)

type client struct {
	c       *mongo.Client
	db      *mongo.Database
	encrypt encryption

	vcCollection                string
	orgEmailCollection          string
	corpPDFCollection           string
	linkCollection              string
	corpSigningCollection       string
	individualSigningCollection string
}

func Initialize(cfg *config.MongodbConfig, encryptionKey, nonce string) (*client, error) {
	c, err := mongo.NewClient(options.Client().ApplyURI(cfg.MongodbConn))
	if err != nil {
		return nil, err
	}

	if err = withContext(c.Connect); err != nil {
		return nil, err
	}

	// verify if database connection is created successfully
	err = withContext(func(ctx context.Context) error {
		return c.Ping(ctx, nil)
	})
	if err != nil {
		return nil, err
	}

	e, err := newEncryption(encryptionKey, nonce)
	if err != nil {
		return nil, err
	}

	cli := &client{
		c:       c,
		db:      c.Database(cfg.DBName),
		encrypt: e,

		vcCollection:                cfg.VCCollection,
		orgEmailCollection:          cfg.OrgEmailCollection,
		corpPDFCollection:           cfg.CorpPDFCollection,
		linkCollection:              cfg.LinkCollection,
		corpSigningCollection:       cfg.CorpSigningCollection,
		individualSigningCollection: cfg.IndividualSigningCollection,
	}
	return cli, nil
}

func (this *client) Close() error {
	return withContext(this.c.Disconnect)
}

func (c *client) collection(name string) *mongo.Collection {
	return c.db.Collection(name)
}

func (this *client) doTransaction(f func(mongo.SessionContext) error) error {

	callback := func(sc mongo.SessionContext) (interface{}, error) {
		return nil, f(sc)
	}

	s, err := this.c.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start mongodb session: %s", err.Error())
	}

	ctx := context.Background()
	defer s.EndSession(ctx)

	_, err = s.WithTransaction(ctx, callback)
	return err
}

func objectIDToUID(oid primitive.ObjectID) string {
	return oid.Hex()
}

func toUID(oid interface{}) (string, error) {
	v, ok := oid.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("retrieve id failed")
	}
	return v.Hex(), nil
}

func withContext(f func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return f(ctx)
}
