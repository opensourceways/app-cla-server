package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func NewLink(dao dao, claContentDao dao) *link {
	return &link{
		dao:        dao,
		claContent: claContent{claContentDao},
	}
}

type link struct {
	dao        dao
	claContent claContent
}

func (impl *link) docFilter(linkId string) bson.M {
	return bson.M{
		fieldId: linkId,
	}
}

func (impl *link) NewLinkId() string {
	return impl.dao.NewDocId()
}

func (impl *link) Add(v *domain.Link) error {
	for i := range v.CLAs {
		if err := impl.claContent.add(v.Id, &v.CLAs[i]); err != nil {
			return err
		}
	}

	do := toLinkDO(v)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}
	doc[fieldVersion] = 0
	doc[fieldRemoved] = bson.A{}

	org := &v.Org
	filter := bson.M{
		fieldDeleted:                        false,
		childField(fieldOrg, fieldOrgAlias): org.Alias,
	}

	_, err = impl.dao.InsertDocIfNotExists(filter, doc)

	if err != nil && impl.dao.IsDocExists(err) {
		err = commonRepo.NewErrorDuplicateCreating(err)
	}

	return err
}

func (impl *link) Remove(link *domain.Link) error {
	v := impl.docFilter(link.Id)
	v[fieldDeleted] = false

	err := impl.dao.UpdateDoc(v, bson.M{fieldDeleted: true}, link.Version)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}

func (impl *link) Find(linkId string) (r domain.Link, err error) {
	var do linkDO

	err = impl.dao.GetDoc(impl.docFilter(linkId), bson.M{fieldRemoved: 0}, &do)
	if err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}

		return
	}

	return do.toLink(), nil
}

func (impl *link) FindAll(userId string) ([]repository.LinkSummary, error) {
	filter := bson.M{
		fieldDeleted:   false,
		fieldSubmitter: userId,
	}

	var dos []linkDO

	project := bson.M{
		fieldCLAs:    0,
		fieldRemoved: 0,
	}

	err := impl.dao.GetDocs(filter, project, &dos)
	if err != nil || len(dos) == 0 {
		return nil, err
	}

	r := make([]repository.LinkSummary, len(dos))
	for i := range dos {
		item := &dos[i]

		r[i] = repository.LinkSummary{
			Id:        item.Id,
			Org:       item.Org.toOrgInfo(),
			Email:     item.Email.toEmailInfo(),
			Submitter: item.Submitter,
		}
	}

	return r, nil
}
