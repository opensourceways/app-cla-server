package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func (impl *corpSigning) SaveCorpPDFURL(cs *domain.CorpSigning, url string) error {
	index, err := impl.toCorpSigningIndex(cs.Id)
	if err != nil {
		return err
	}

	err = impl.dao.UpdateDoc(
		index, bson.M{fieldPDFURL: url, fieldHasPDF: true, fieldTriggered: true}, cs.Version,
	)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}

func (impl *corpSigning) FindCorpPDFURL(csId string) (string, error) {
	filter, err := impl.toCorpSigningIndex(csId)
	if err != nil {
		return "", err
	}

	var do struct {
		PDFURL string `bson:"pdf_url"`
	}

	err = impl.dao.GetDoc(filter, bson.M{fieldPDFURL: 1}, &do)
	if err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}

		return "", err
	}

	return do.PDFURL, nil
}

func (impl *corpSigning) FindCorpPDF(csId string) ([]byte, error) {
	filter, err := impl.toCorpSigningIndex(csId)
	if err != nil {
		return nil, err
	}

	var do corpSigningDO

	err = impl.dao.GetDoc(filter, bson.M{fieldPDF: 1}, &do)
	if err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}

		return nil, err
	}

	return do.PDF, nil
}

func (impl *corpSigning) SetCorpPDFURLAndUnsetPDF(csId, url string) error {
	filter, err := impl.toCorpSigningIndex(csId)
	if err != nil {
		return err
	}

	return impl.dao.UpdateDocsWithoutVersion(filter, bson.M{
		fieldPDFURL: url,
		fieldHasPDF: true,
		fieldPDF:    nil,
	})
}

func (impl *corpSigning) FindCorpSigningIDsForPDFMigration(page, pageSize int) ([]string, int64, error) {
	filter := bson.M{
		fieldHasPDF: true,
		fieldPDFURL: "",
		fieldPDF:    bson.M{"$exists": true, "$ne": nil, "$not": bson.M{"$size": 0}},
	}

	total, err := impl.dao.CountDocs(filter)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return nil, 0, nil
	}

	skip := (page - 1) * pageSize
	if skip < 0 {
		skip = 0
	}

	var dos []struct {
		Id primitive.ObjectID `bson:"_id"`
	}

	err = impl.dao.GetDocsWithPagination(filter, bson.M{"_id": 1}, skip, pageSize, &dos)
	if err != nil {
		return nil, 0, err
	}

	ids := make([]string, len(dos))
	for i := range dos {
		ids[i] = dos[i].Id.Hex()
	}

	return ids, total, nil
}
