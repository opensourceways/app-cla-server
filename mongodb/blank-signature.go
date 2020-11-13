package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

func (this *client) UploadBlankSignature(language string, pdf []byte) error {
	f := func(ctx context.Context) error {
		_, err := this.newDoc(
			ctx, this.blankSigCollection, bson.M{"language": language},
			bson.M{
				"language": language,
				"pdf":      pdf,
			},
		)
		return err
	}

	return withContext(f)
}

func (this *client) DownloadBlankSignature(language string) ([]byte, error) {
	var v struct {
		PDF []byte `bson:"pdf"`
	}

	f := func(ctx context.Context) error {
		return this.getDoc(ctx, this.blankSigCollection, bson.M{"language": language}, bson.M{"pdf": 1}, &v)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	return v.PDF, nil
}
