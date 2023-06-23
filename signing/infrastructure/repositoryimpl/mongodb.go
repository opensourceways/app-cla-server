package repositoryimpl

import (
	"encoding/json"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

type anyDoc = map[string]string

type dao interface {
	IsDocNotExists(error) bool
	IsDocExists(error) bool

	InsertDocIfNotExists(filter, doc bson.M) (string, error)
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
