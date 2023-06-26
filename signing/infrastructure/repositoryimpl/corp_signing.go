package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func NewCorpSigning(dao dao) *corpSigning {
	return &corpSigning{
		dao: dao,
	}
}

type corpSigning struct {
	dao dao
}

func (impl *corpSigning) toCorpSigningIndex(corpSigningId string) (bson.M, error) {
	return impl.dao.DocIdFilter(corpSigningId)
}

func (impl *corpSigning) Add(v *domain.CorpSigning) error {
	do := toCorpSigningDO(v)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}
	doc[fieldVersion] = 0

	docFilter := linkIdFilter(v.Link.Id)
	docFilter["$or"] = bson.A{
		bson.M{childField(fieldRep, fieldEmail): v.Rep.EmailAddr.EmailAddr()},
		bson.M{
			childField(fieldCorp, fieldName):   v.Corp.Name.CorpName(),
			childField(fieldCorp, fieldDomain): v.Corp.PrimaryEmailDomain,
		},
	}

	_, err = impl.dao.InsertDocIfNotExists(docFilter, doc)
	if err != nil && impl.dao.IsDocExists(err) {
		err = commonRepo.NewErrorDuplicateCreating(err)
	}

	return err
}

func (impl *corpSigning) Find(index string) (cs domain.CorpSigning, err error) {
	filter, err := impl.toCorpSigningIndex(index)
	if err != nil {
		return
	}

	var do corpSigningDO

	if err = impl.dao.GetDoc(filter, nil, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}
	} else {
		err = do.toCorpSigning(&cs)
	}

	return
}

func (impl *corpSigning) Count(linkId, domain string) (int, error) {
	filter := linkIdFilter(linkId)
	filter[childField(fieldCorp, fieldDomain)] = domain

	var dos []struct {
		LinkId string `bson:"link_id"`
	}

	err := impl.dao.GetDocs(filter, bson.M{fieldLinkId: 1}, &dos)

	return len(dos), err
}
