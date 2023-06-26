package repositoryimpl

import (
	"encoding/json"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	mongodbCmdOr = "$or"
)

type anyDoc = map[string]string

type dao interface {
	IsDocNotExists(error) bool
	IsDocExists(error) bool

	NewDocId() string
	DocIdFilter(s string) (bson.M, error)

	PushArrayDoc(filter bson.M, field string, doc bson.M, version int) error
	UpdateDoc(filter bson.M, doc bson.M, version int) error

	InsertDocIfNotExists(filter, doc bson.M) (string, error)

	DeleteDoc(filter bson.M) error

	GetDoc(filter, project bson.M, result interface{}) error
	GetDocs(filter, project bson.M, result interface{}) error
}

func genDoc(doc interface{}) (m bson.M, err error) {
	v, err := json.Marshal(doc)
	if err != nil {
		return
	}

	err = json.Unmarshal(v, &m)

	return
}

func linkIdFilter(v string) bson.M {
	return bson.M{
		fieldLinkId: v,
	}
}

func childField(fields ...string) string {
	return strings.Join(fields, ".")
}
