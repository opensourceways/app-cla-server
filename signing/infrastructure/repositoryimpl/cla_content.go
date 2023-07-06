package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

type claContent struct {
	dao dao
}

func (impl *claContent) docFilter(linkId, claId string) bson.M {
	filter := linkIdFilter(linkId)
	filter[fieldCLAId] = claId

	return filter
}

func (impl *claContent) add(linkId string, v *domain.CLA) error {
	doc := impl.docFilter(linkId, v.Id)
	doc[fieldText] = v.Text

	_, err := impl.dao.ReplaceDoc(impl.docFilter(linkId, v.Id), doc)

	return err
}
