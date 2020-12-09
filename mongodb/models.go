package mongodb

const (
	fieldLinkStatus  = "link_status"
	fieldOrgIdentity = "org_identity"
	fieldLinkID      = "link_id"

	// 'ready' means the doc is ready to record the signing data currently.
	// 'unready' means the doc is not ready.
	// 'deleted' means the signing data is invalid.
	linkStatusReady   = "ready"
	linkStatusUnready = "unready"
	linkStatusDeleted = "deleted"
)

type dCorpSigningPDF struct {
	LinkID string `bson:"link_id" json:"link_id" required:"true"`
	CorpID string `bson:"corp_id" json:"corp_id" required:"true"`
	PDF    []byte `bson:"pdf" json:"pdf,omitempty"`
}
