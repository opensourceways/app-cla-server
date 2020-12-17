package mongodb

func (this *client) DownloadOrgSignature(orgCLAID string) ([]byte, error) {
	/*
		oid, err := toObjectID(orgCLAID)
		if err != nil {
			return nil, err
		}

		var v OrgCLA

		f := func(ctx context.Context) error {
			return this.getDoc(
				ctx, this.linkCollection, filterOfDocID(oid),
				bson.M{fieldOrgSignature: 1},
				&v,
			)
		}

		if withContext(f); err != nil {
			return nil, err
		}

		return v.OrgSignature, nil
	*/
	return nil, nil
}
