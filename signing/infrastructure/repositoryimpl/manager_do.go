package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

const fieldId = "id"

func toManagerDO(m *domain.Manager) managerDO {
	return managerDO{
		Id:    m.Id,
		RepDO: toRepDO(&m.Representative),
	}
}

// managerDO
type managerDO struct {
	Id string `bson:"id" json:"id"`

	RepDO `bson:",inline"`
}

func (do *managerDO) isEmpty() bool {
	return do.Id == ""
}

func (do *managerDO) toManager() domain.Manager {
	if do.isEmpty() {
		return domain.Manager{}
	}

	return domain.Manager{
		Id:             do.Id,
		Representative: do.RepDO.toRep(),
	}
}

func (do *managerDO) toDoc() (bson.M, error) {
	return genDoc(do)
}
