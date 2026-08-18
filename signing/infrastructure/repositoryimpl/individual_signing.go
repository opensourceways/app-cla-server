package repositoryimpl

import (
	"fmt"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

const (
	fieldLogs = "logs"
)

// IndividualSigningIndexes returns the index definitions that should exist on
// the individual_signing collection. Pass the result to mongodb.EnsureIndexes
// on startup.
func IndividualSigningIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		// Support FindByDomains: match active individual signings of one
		// link by email domain (case-insensitive regex).
		{Keys: bson.D{{Key: fieldLinkId, Value: 1}, {Key: fieldDomain, Value: 1}}},
	}
}

func NewIndividualSigning(dao dao) *individualSigning {
	return &individualSigning{
		dao: dao,
	}
}

type individualSigning struct {
	dao dao
}

func (impl *individualSigning) Add(is *domain.IndividualSigning) error {
	do := toIndividualSigningDO(is)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}
	doc[fieldVersion] = 0

	filter := linkIdFilter(is.Link.Id)
	filter[fieldEmail] = is.Rep.EmailAddr.EmailAddr()
	filter[fieldDeleted] = false

	_, err = impl.dao.InsertDocIfNotExists(filter, doc)
	if err != nil && impl.dao.IsDocExists(err) {
		err = commonRepo.NewErrorDuplicateCreating(err)
	}

	return err
}

func (impl *individualSigning) AddForMigrate(is *domain.IndividualSigning) error {
	do := toIndividualSigningDO(is)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}
	doc[fieldVersion] = 0

	filter := linkIdFilter(is.Link.Id)
	filter[fieldEmail] = is.Rep.EmailAddr.EmailAddr()
	filter[fieldDeleted] = false

	_, err = impl.dao.InsertDocIfNotExists(filter, doc)
	if err != nil && impl.dao.IsDocExists(err) {
		err = commonRepo.NewErrorDuplicateCreating(err)
	}

	return err
}

func (impl *individualSigning) FindSignedCLA(linkId string, email dp.EmailAddr) (string, dp.Language, error) {
	filter := linkIdFilter(linkId)
	filter[fieldEmail] = email.EmailAddr()
	filter[fieldDeleted] = false

	var do individualSigningDO

	if err := impl.dao.GetDoc(filter, nil, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			return "", nil, nil
		}

		return "", nil, err
	}

	return do.CLAId, dp.CreateLanguage(do.Language), nil
}

func (impl *individualSigning) Find(linkId string, email dp.EmailAddr) (domain.IndividualSigning, error) {
	filter := linkIdFilter(linkId)
	filter[fieldEmail] = email.EmailAddr()
	filter[fieldDeleted] = false

	var do individualSigningDO

	if err := impl.dao.GetDoc(filter, nil, &do); err != nil {
		return domain.IndividualSigning{}, err
	}

	return do.toIndividualSigning(), nil
}

// toDomainRegexConditions converts email domains to anchored
// case-insensitive regex conditions, escaping regex metacharacters so that
// a domain is matched literally.
func toDomainRegexConditions(domains []string) bson.A {
	conditions := make(bson.A, 0, len(domains))

	for _, d := range domains {
		conditions = append(conditions, bson.M{
			"$regex": fmt.Sprintf("(?i)^%s$", regexp.QuoteMeta(d)),
		})
	}

	return conditions
}

func (impl *individualSigning) FindByDomains(linkId string, domains []string) ([]domain.IndividualSigning, error) {
	if len(domains) == 0 {
		return nil, nil
	}

	filter := linkIdFilter(linkId)
	filter[fieldDeleted] = false
	filter[fieldDomain] = bson.M{mongodbCmdIn: toDomainRegexConditions(domains)}

	var dos []individualSigningDO

	if err := impl.dao.GetDocs(filter, nil, &dos); err != nil {
		return nil, err
	}

	result := make([]domain.IndividualSigning, len(dos))
	for i := range dos {
		result[i] = dos[i].toIndividualSigning()
	}

	return result, nil
}

func (impl *individualSigning) FindAll(linkId string) ([]domain.IndividualSigning, error) {
	filter := linkIdFilter(linkId)
	filter[fieldDeleted] = false

	var dos []individualSigningDO

	if err := impl.dao.GetDocs(filter, nil, &dos); err != nil {
		return nil, err
	}

	var result []domain.IndividualSigning
	for _, do := range dos {
		result = append(result, do.toIndividualSigning())
	}

	return result, nil
}

func (impl *individualSigning) FindAllWithPagination(linkId string, offset, limit int) ([]domain.IndividualSigning, error) {
	filter := linkIdFilter(linkId)
	filter[fieldDeleted] = false

	var dos []individualSigningDO
	if err := impl.dao.GetDocsWithPagination(filter, nil, offset, limit, &dos); err != nil {
		return nil, err
	}
	result := make([]domain.IndividualSigning, len(dos))
	for i := range dos {
		result[i] = dos[i].toIndividualSigning()
	}

	return result, nil
}

func (impl *individualSigning) CountByLinkId(linkId string) (int64, error) {
	filter := linkIdFilter(linkId)
	filter[fieldDeleted] = false
	return impl.dao.CountDocs(filter)
}

func (impl *individualSigning) HasSignedLink(linkId string) (bool, error) {
	filter := linkIdFilter(linkId)

	var do individualSigningDO

	if err := impl.dao.GetDoc(filter, bson.M{fieldLinkId: 1}, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (impl *individualSigning) HasSignedCLA(index *domain.CLAIndex) (bool, error) {
	filter := linkIdFilter(index.LinkId)
	filter[fieldCLAId] = index.CLAId

	var do individualSigningDO

	if err := impl.dao.GetDoc(filter, bson.M{fieldLinkId: 1}, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (impl *individualSigning) SaveNewCLA(is *domain.IndividualSigning) error {
	filter := linkIdFilter(is.Link.Id)
	filter[fieldEmail] = is.Rep.EmailAddr.EmailAddr()
	filter[fieldDeleted] = false

	do := toIndividualSigningLogsDO(is.Logs)
	doc, err := genDoc(do)
	if err != nil {
		return err
	}
	doc[fieldCLAId] = is.Link.CLAId
	doc[fieldCLANotify] = ""
	doc["cla_notify_count"] = 0
	doc["cla_notify_time"] = int64(0)

	return impl.dao.UpdateDoc(filter, doc, is.Version)
}

func (impl *individualSigning) UpdateCLANotify(is *domain.IndividualSigning) error {
	filter := linkIdFilter(is.Link.Id)
	filter[fieldEmail] = is.Rep.EmailAddr.EmailAddr()
	filter[fieldDeleted] = false

	doc := bson.M{
		fieldCLANotify:      is.ClaNotify,
		"cla_notify_count":  is.ClaNotifyCount,
		"cla_notify_time":   is.ClaNotifyTime,
	}

	return impl.dao.UpdateDoc(filter, doc, is.Version)
}
