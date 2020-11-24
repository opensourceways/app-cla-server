package mongodb

import (
	"fmt"
	"strings"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func dbValueOfRepo(org, repo string) string {
	if repo == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", org, repo)
}

func toNormalRepo(repo string) string {
	if strings.Contains(repo, "/") {
		return strings.Split(repo, "/")[1]
	}
	return repo
}

func structToMap(info interface{}) (bson.M, error) {
	body, err := golangsdk.BuildRequestBody(info, "")
	if err != nil {
		return nil, dbmodels.DBError{
			ErrCode: util.ErrInvalidParameter,
			Err:     err,
		}
	}
	return bson.M(body), nil
}

func addCorporationID(email string, body bson.M) {
	body[fieldCorporationID] = genCorpID(email)
}

func genCorpID(email string) string {
	return util.EmailSuffix(email)
}

func filterOfCorpID(email string) bson.M {
	return bson.M{fieldCorporationID: genCorpID(email)}
}

func filterOfDocID(oid primitive.ObjectID) bson.M {
	return bson.M{"_id": oid}
}

func indexOfCorpManagerAndIndividual(email string) bson.M {
	return bson.M{
		fieldCorporationID: genCorpID(email),
		"email":            email,
	}
}

func filterOfOrgRepo(platform, org, repo string) (bson.M, error) {
	if !(platform != "" && org != "") {
		return nil, dbmodels.DBError{
			ErrCode: util.ErrNoPlatformOrOrg,
			Err:     fmt.Errorf("platform or org is empty"),
		}
	}

	if repo == "" {
		return bson.M{
			"platform": platform,
			"org_id":   org,
			fieldRepo:  "",
		}, nil
	}

	return bson.M{
		"platform": platform,
		fieldRepo:  dbValueOfRepo(org, repo),
	}, nil
}

func isErrorOfNotSigned(err error) bool {
	e, ok := dbmodels.IsDBError(err)
	return ok && e.ErrCode == util.ErrHasNotSigned
}

func isErrorOfRecordExists(err error) bool {
	e, ok := dbmodels.IsDBError(err)
	return ok && e.ErrCode == util.ErrRecordExists
}

func toObjectID(uid string) (primitive.ObjectID, error) {
	v, err := primitive.ObjectIDFromHex(uid)
	if err != nil {
		return v, dbmodels.DBError{
			ErrCode: util.ErrInvalidParameter,
			Err:     fmt.Errorf("can't convert to object id"),
		}
	}
	return v, err
}

func isErrNoDocuments(err error) bool {
	return err.Error() == mongo.ErrNoDocuments.Error()
}

func arrayFilterByElemMatch(array string, exists bool, cond, filter bson.M) {
	match := bson.M{"$elemMatch": cond}
	if exists {
		filter[array] = match
	} else {
		filter[array] = bson.M{"$not": match}
	}
}
