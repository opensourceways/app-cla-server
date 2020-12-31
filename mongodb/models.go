package mongodb

const (
	fieldLinkID = "link_id"
)

type dCorpSigningPDF struct {
	LinkID string `bson:"link_id" json:"link_id" required:"true"`
	CorpID string `bson:"corp_id" json:"corp_id" required:"true"`
	PDF    []byte `bson:"pdf" json:"pdf,omitempty"`
}
