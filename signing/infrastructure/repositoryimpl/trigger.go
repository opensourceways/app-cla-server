package repositoryimpl

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type TriggeredCorp struct {
	Id      string
	LinkId  string
	Domains []string
	Version int
}

func (impl *corpSigning) ListTriggered() ([]TriggeredCorp, error) {
	filter := bson.M{
		fieldTriggered: true,
	}

	project := bson.M{
		fieldLinkId:                         1,
		fieldVersion:                        1,
		childField(fieldCorp, fieldDomains): 1,
	}

	var dos []corpSigningDO

	if err := impl.dao.GetDocs(filter, project, &dos); err != nil {
		return nil, err
	}

	v := make([]TriggeredCorp, len(dos))
	for i := range dos {
		item := &dos[i]

		v[i] = TriggeredCorp{
			Id:      item.Id.Hex(),
			LinkId:  item.LinkId,
			Domains: item.Corp.Domains,
			Version: item.Version,
		}
	}

	return v, nil
}

func (impl *corpSigning) ResetTriggered(csId string, version int) error {
	filter, err := impl.dao.DocIdFilter(csId)
	if err != nil {
		return err
	}

	doc := bson.M{fieldTriggered: false}

	return impl.dao.UpdateDoc(filter, doc, version)
}

func (impl *individualSigning) RemoveAll(linkId string, domains []string) error {
	return impl.dao.UpdateDocsWithoutVersion(
		bson.M{
			fieldLinkId:  linkId,
			fieldDeleted: false,
			fieldDomain:  bson.M{mongodbCmdIn: domains},
		},
		bson.M{
			fieldDeleted:   true,
			fieldDeletedAt: time.Now().Unix(),
		},
	)
}
