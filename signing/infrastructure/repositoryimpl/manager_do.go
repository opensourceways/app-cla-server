package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

func toManagerDO(m *domain.Manager) managerDO {
	return managerDO{
		Id:    m.Id,
		RepDO: toRepDO(&m.Representative),
	}
}

// managerDO
type managerDO struct {
	Id string `bson:"id" json:"account"`

	RepDO `bson:",inline"`
}

func (do *managerDO) isEmpty() bool {
	return do.Id == ""
}

func (do *managerDO) toManager() (m domain.Manager, err error) {
	if do.isEmpty() {
		return
	}

	if m.Representative, err = do.RepDO.toRep(); err != nil {
		return
	}

	m.Id = do.Id

	return
}

func (do *managerDO) toDoc() (bson.M, error) {
	return genDoc(do)
}
