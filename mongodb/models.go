package mongodb

const (
	fieldLinkID = "link_id"
)

type dCorpSigningPDF struct {
	LinkID string `bson:"link_id" json:"link_id" required:"true"`
	CorpID string `bson:"corp_id" json:"corp_id" required:"true"`
	PDF    []byte `bson:"pdf" json:"pdf,omitempty"`
}

type cVerificationCode struct {
	Email   string `bson:"email" json:"email" required:"true"`
	Code    string `bson:"code" json:"code" required:"true"`
	Purpose string `bson:"purpose" json:"purpose" required:"true"`
	Expiry  int64  `bson:"expiry" json:"expiry" required:"true"`
}
