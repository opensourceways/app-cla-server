package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

var (
	errNoDBRecord1 = dbError{code: dbmodels.ErrNoDBRecord, err: fmt.Errorf("no record")}
)

func withContext1(f func(context.Context) dbmodels.IDBError) dbmodels.IDBError {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return f(ctx)
}

func structToMap1(info interface{}) (bson.M, dbmodels.IDBError) {
	body, err := golangsdk.BuildRequestBody(info, "")
	if err != nil {
		return nil, newDBError(dbmodels.ErrMarshalDataFaield, err)
	}
	return bson.M(body), nil
}
