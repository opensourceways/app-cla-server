package repositoryimpl

func NewOrg(dao dao) *orgImpl {
	return &orgImpl{
		dao: dao,
	}
}

type orgImpl struct {
	dao dao
}

func (impl *orgImpl) Find(platform string) ([]string, error) {
	var v orgsDO

	err := impl.dao.GetDoc(nil, nil, &v)
	if err != nil {
		if impl.dao.IsDocNotExists(err) {
			return nil, nil
		}

		return nil, err
	}

	return v.get(platform), nil
}

type orgsDO struct {
	Orgs []orgDO `bson:"orgs"`
}

func (do *orgsDO) get(platform string) []string {
	for i := range do.Orgs {
		if do.Orgs[i].Platform == platform {
			return do.Orgs[i].Orgs
		}
	}

	return nil
}

type orgDO struct {
	Platform string   `bson:"platform"`
	Orgs     []string `bson:"orgs"`
}
