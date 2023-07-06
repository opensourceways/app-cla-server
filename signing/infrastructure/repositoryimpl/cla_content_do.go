package repositoryimpl

const (
	fieldText  = "text"
	fieldCLAId = "cla_id"
)

// claContentDO
type claContentDO struct {
	LinkId string `bson:"link_id" json:"link_id" required:"true"`
	CLAId  string `bson:"cla_id"  json:"cla_id"  required:"true"`
	Text   []byte `bson:"text"    json:"-"`
}
