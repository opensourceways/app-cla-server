package mongodb

const (
	fieldLinkStatus  = "link_status"
	fieldOrgIdentity = "org_identity"

	// 'ready' means the doc is ready to record the signing data currently.
	// 'unready' means the doc is not ready.
	// 'deleted' means the signing data is invalid.
	linkStatusReady   = "ready"
	linkStatusUnready = "unready"
	linkStatusDeleted = "deleted"
)

type dCorpSigningPDF struct {
	OrgIdentity string `bson:"org_identity" json:"org_identity" required:"true"`
	LinkStatus  string `bson:"link_status" json:"link_status"`

	CorpID string `bson:"corp_id" json:"corp_id" required:"true"`
	PDF    []byte `bson:"pdf" json:"pdf,omitempty"`
}
