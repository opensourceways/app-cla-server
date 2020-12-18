package mongodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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

func structToMap(info interface{}) (bson.M, *dbmodels.DBError) {
	body, err := golangsdk.BuildRequestBody(info, "")
	if err != nil {
		return nil, newDBError(dbmodels.ErrMarshalDataFaield, err)
	}
	return bson.M(body), nil
}

func addCorporationID(email string, body bson.M) {
	body[fieldCorpID] = genCorpID(email)
}

func genCorpID(email string) string {
	return util.EmailSuffix(email)
}

func filterOfCorpID(email string) bson.M {
	return bson.M{fieldCorpID: genCorpID(email)}
}

func filterOfDocID(oid primitive.ObjectID) bson.M {
	return bson.M{"_id": oid}
}

func indexOfCorpManagerAndIndividual(email string) bson.M {
	return bson.M{
		fieldCorpID: genCorpID(email),
		"email":     email,
	}
}

func isErrorOfRecordExists(err error) bool {
	e, ok := err.(*dbmodels.DBError)
	return ok && e.Code == dbmodels.ErrRecordExists
}

func isErrNoDocuments(err error) bool {
	return err.Error() == mongo.ErrNoDocuments.Error()
}

func isErrOfNoDocument(err error) bool {
	return err.Error() == errNoDBRecord.Error()
}

func errorIfMatchingNoDoc(r *mongo.UpdateResult) error {
	if r.MatchedCount == 0 {
		return errNoDBRecord
	}
	return nil
}

func (this *client) pushArrayElems(ctx context.Context, collection, array string, filterOfDoc bson.M, value bson.A) error {
	update := bson.M{"$push": bson.M{array: bson.M{"$each": value}}}

	col := this.collection(collection)
	r, err := col.UpdateOne(ctx, filterOfDoc, update)
	if err != nil {
		return err
	}

	return errorIfMatchingNoDoc(r)
}

func (this *client) pullArrayElem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray bson.M) *dbmodels.DBError {
	update := bson.M{"$pull": bson.M{array: filterOfArray}}

	col := this.collection(collection)
	r, err := col.UpdateOne(ctx, filterOfDoc, update)
	if err != nil {
		return systemError(err)
	}

	if r.MatchedCount == 0 {
		return errNoDBRecord
	}
	return nil
}

// r, _ := col.UpdateOne; r.ModifiedCount == 0 will happen in two case: 1. no matched array item; 2 update repeatedly with same update cmd.
// checkModified = true when it can't exclude any case of above two; otherwise it can be set as false.
func (this *client) updateArrayElem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray, updateCmd bson.M) *dbmodels.DBError {
	cmd := bson.M{}
	for k, v := range updateCmd {
		cmd[fmt.Sprintf("%s.$[i].%s", array, k)] = v
	}

	arrayFilter := bson.M{}
	for k, v := range filterOfArray {
		arrayFilter["i."+k] = v
	}

	col := this.collection(collection)
	r, err := col.UpdateOne(
		ctx, filterOfDoc,
		bson.M{"$set": cmd},
		&options.UpdateOptions{
			ArrayFilters: &options.ArrayFilters{
				Filters: bson.A{
					arrayFilter,
				},
			},
		},
	)
	if err != nil {
		return systemError(err)
	}

	if r.MatchedCount == 0 {
		return errNoDBRecord
	}
	return nil
}

func (this *client) getArrayElem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray, project bson.M, result interface{}) error {
	ma := map[string]bson.M{}
	if len(filterOfArray) > 0 {
		ma[array] = filterOfArray
	}
	return this.getMultiArrays(ctx, collection, filterOfDoc, ma, project, result)
}

func (this *client) getMultiArrays(ctx context.Context, collection string, filterOfDoc bson.M, filterOfArrays map[string]bson.M, project bson.M, result interface{}) error {
	pipeline := bson.A{bson.M{"$match": filterOfDoc}}

	if len(filterOfArrays) > 0 {
		project1 := bson.M{}

		for array, filterOfArray := range filterOfArrays {
			project1[array] = bson.M{"$filter": bson.M{
				"input": fmt.Sprintf("$%s", array),
				"cond":  conditionTofilterArray(filterOfArray),
			}}
		}

		for k, v := range project {
			s := k
			if i := strings.Index(k, "."); i >= 0 {
				s = k[:i]
			}
			if _, ok := filterOfArrays[s]; !ok {
				project1[k] = v
			}
		}

		pipeline = append(pipeline, bson.M{"$project": project1})
	}

	if len(project) > 0 {
		pipeline = append(pipeline, bson.M{"$project": project})
	}

	col := this.collection(collection)
	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	return cursor.All(ctx, result)
}

func conditionTofilterArray(filterOfArray bson.M) bson.M {
	cond := make(bson.A, 0, len(filterOfArray))
	for k, v := range filterOfArray {
		cond = append(cond, bson.M{"$eq": bson.A{"$$this." + k, v}})
	}

	if len(filterOfArray) == 1 {
		return cond[0].(bson.M)
	}

	return bson.M{"$and": cond}
}

func (this *client) replaceDoc(ctx context.Context, collection string, filterOfDoc, docInfo bson.M) (string, error) {
	upsert := true

	col := this.collection(collection)
	r, err := col.ReplaceOne(
		ctx, filterOfDoc, docInfo,
		&options.ReplaceOptions{Upsert: &upsert},
	)
	if err != nil {
		return "", err
	}

	if r.UpsertedID == nil {
		return "", nil
	}

	return toUID(r.UpsertedID)
}

func (this *client) newDocIfNotExist(ctx context.Context, collection string, filterOfDoc, docInfo bson.M) (string, error) {
	upsert := true

	col := this.collection(collection)
	r, err := col.UpdateOne(
		ctx, filterOfDoc, bson.M{"$setOnInsert": docInfo},
		&options.UpdateOptions{Upsert: &upsert},
	)
	if err != nil {
		return "", err
	}

	if r.UpsertedID == nil {
		return "", errRecordExist
	}

	return toUID(r.UpsertedID)
}

func (this *client) newDoc(ctx context.Context, collection string, filterOfDoc, docInfo bson.M) (string, error) {
	upsert := true

	col := this.collection(collection)
	r, err := col.ReplaceOne(
		ctx, filterOfDoc, docInfo,
		&options.ReplaceOptions{Upsert: &upsert},
	)
	if err != nil {
		return "", err
	}

	if r.UpsertedID != nil {
		return toUID(r.UpsertedID)
	}
	return "", nil
}

func (this *client) updateDoc(ctx context.Context, collection string, filterOfDoc, update bson.M) error {
	col := this.collection(collection)
	r, err := col.UpdateOne(ctx, filterOfDoc, bson.M{"$set": update})
	if err != nil {
		return err
	}
	return errorIfMatchingNoDoc(r)
}

func (this *client) getDoc(ctx context.Context, collection string, filterOfDoc, project bson.M, result interface{}) error {
	col := this.collection(collection)

	var sr *mongo.SingleResult
	if len(project) > 0 {
		sr = col.FindOne(ctx, filterOfDoc, &options.FindOneOptions{
			Projection: project,
		})
	} else {
		sr = col.FindOne(ctx, filterOfDoc)
	}

	if err := sr.Decode(result); err != nil {
		if err == mongo.ErrNoDocuments {
			return errNoDBRecord
		}
		return err
	}
	return nil
}

func (this *client) getDocs(ctx context.Context, collection string, filterOfDoc, project bson.M, result interface{}) error {
	col := this.collection(collection)

	var cursor *mongo.Cursor
	var err error
	if len(project) > 0 {
		cursor, err = col.Find(ctx, filterOfDoc, &options.FindOptions{
			Projection: project,
		})
	} else {
		cursor, err = col.Find(ctx, filterOfDoc)
	}

	if err != nil {
		return err
	}
	return cursor.All(ctx, result)
}

func (this *client) insertDoc(ctx context.Context, collection string, docInfo bson.M) (string, error) {
	col := this.collection(collection)
	r, err := col.InsertOne(ctx, docInfo)
	if err != nil {
		return "", err
	}

	return toUID(r.InsertedID)
}

func arrayFilterByElemMatch(array string, exists bool, cond, filter bson.M) {
	match := bson.M{"$elemMatch": cond}
	if exists {
		filter[array] = match
	} else {
		filter[array] = bson.M{"$not": match}
	}
}
